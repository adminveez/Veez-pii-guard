// Command pii-guard-lsp is a minimal Language Server Protocol implementation
// that highlights PII detections as diagnostics in any LSP-capable editor
// (VS Code, Neovim, Helix, Emacs).
//
// Scope is intentionally minimal:
//   - textDocument/didOpen, didChange, didClose, didSave
//   - publishDiagnostics
//
// No completions, no quickfixes, no workspace symbols. We are an extra
// pair of eyes for PII, not a refactoring tool.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/veez-ai/veez-pii-guard/pii"
)

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type position struct {
	Line      int `json:"line"`
	Character int `json:"character"`
}

type lspRange struct {
	Start position `json:"start"`
	End   position `json:"end"`
}

type diagnostic struct {
	Range    lspRange `json:"range"`
	Severity int      `json:"severity"` // 1=error, 2=warn, 3=info, 4=hint
	Source   string   `json:"source"`
	Message  string   `json:"message"`
}

type publishDiagnosticsParams struct {
	URI         string       `json:"uri"`
	Diagnostics []diagnostic `json:"diagnostics"`
}

type didOpenParams struct {
	TextDocument struct {
		URI  string `json:"uri"`
		Text string `json:"text"`
	} `json:"textDocument"`
}

type didChangeParams struct {
	TextDocument struct {
		URI string `json:"uri"`
	} `json:"textDocument"`
	ContentChanges []struct {
		Text string `json:"text"`
	} `json:"contentChanges"`
}

type server struct {
	mu       sync.Mutex
	out      *bufio.Writer
	detector *pii.Detector
}

func main() {
	log.SetOutput(os.Stderr)
	cfg := pii.DefaultConfig()
	cfg.AnonymizeOutput = false
	d := pii.MustNewDetector(cfg)
	s := &server{
		out:      bufio.NewWriter(os.Stdout),
		detector: d,
	}
	if err := s.serve(os.Stdin); err != nil {
		log.Fatal(err)
	}
}

func (s *server) serve(r io.Reader) error {
	br := bufio.NewReader(r)
	for {
		body, err := readMessage(br)
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		var req rpcRequest
		if err := json.Unmarshal(body, &req); err != nil {
			log.Printf("bad request: %v", err)
			continue
		}
		s.dispatch(&req)
	}
}

func (s *server) dispatch(req *rpcRequest) {
	switch req.Method {
	case "initialize":
		s.respond(req.ID, map[string]any{
			"capabilities": map[string]any{
				"textDocumentSync": 1, // full sync
			},
			"serverInfo": map[string]string{
				"name":    "pii-guard-lsp",
				"version": "0.2.0",
			},
		})
	case "initialized", "shutdown":
		if req.ID != nil {
			s.respond(req.ID, nil)
		}
	case "exit":
		os.Exit(0)
	case "textDocument/didOpen":
		var p didOpenParams
		_ = json.Unmarshal(req.Params, &p)
		s.publish(p.TextDocument.URI, p.TextDocument.Text)
	case "textDocument/didChange":
		var p didChangeParams
		_ = json.Unmarshal(req.Params, &p)
		if len(p.ContentChanges) > 0 {
			s.publish(p.TextDocument.URI, p.ContentChanges[len(p.ContentChanges)-1].Text)
		}
	case "textDocument/didSave":
		// Reuse last didChange — nothing to do without contentChanges.
	}
}

func (s *server) publish(uri, text string) {
	res := s.detector.Scan(context.Background(), text)
	diags := make([]diagnostic, 0, len(res.Detections))
	for _, det := range res.Detections {
		startLine, startCh := lineCol(text, det.Start)
		endLine, endCh := lineCol(text, det.End)
		severity := 2
		if det.Type == pii.TypeAPIKey || det.Type == pii.TypeBearerToken || det.Type == pii.TypeSecret {
			severity = 1
		}
		diags = append(diags, diagnostic{
			Range:    lspRange{Start: position{startLine, startCh}, End: position{endLine, endCh}},
			Severity: severity,
			Source:   "veez-pii-guard",
			Message:  fmt.Sprintf("%s detected (confidence %.2f, source %s)", det.Type, det.Confidence, det.Source),
		})
	}
	s.notify("textDocument/publishDiagnostics", publishDiagnosticsParams{URI: uri, Diagnostics: diags})
}

func (s *server) respond(id json.RawMessage, result any) {
	s.write(rpcResponse{JSONRPC: "2.0", ID: id, Result: result})
}

func (s *server) notify(method string, params any) {
	b, _ := json.Marshal(params)
	s.write(map[string]any{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  json.RawMessage(b),
	})
}

func (s *server) write(payload any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("marshal: %v", err)
		return
	}
	_, _ = fmt.Fprintf(s.out, "Content-Length: %d\r\n\r\n", len(body))
	_, _ = s.out.Write(body)
	_ = s.out.Flush()
}

// readMessage reads one LSP framed JSON message.
func readMessage(br *bufio.Reader) ([]byte, error) {
	var contentLength int
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			break
		}
		if strings.HasPrefix(line, "Content-Length:") {
			n, err := strconv.Atoi(strings.TrimSpace(line[len("Content-Length:"):]))
			if err != nil {
				return nil, fmt.Errorf("bad Content-Length: %w", err)
			}
			contentLength = n
		}
	}
	if contentLength <= 0 {
		return nil, fmt.Errorf("missing Content-Length")
	}
	body := make([]byte, contentLength)
	if _, err := io.ReadFull(br, body); err != nil {
		return nil, err
	}
	return body, nil
}

func lineCol(text string, offset int) (line, col int) {
	if offset > len(text) {
		offset = len(text)
	}
	line, col = 0, 0
	for i := 0; i < offset; i++ {
		if text[i] == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}
	return line, col
}
