// Command pii-guard scans, anonymizes and explains PII in text input.
// See "pii-guard help" for usage.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/veez-ai/veez-pii-guard/pii"
)

const usage = `pii-guard <command> [flags]

Commands:
  scan        Detect PII and optionally block on policy violations.
  anonymize   Print anonymized text with deterministic placeholders.
  explain     Print every detection with its source pattern, span and confidence.
  stream      Stream-scan stdin and emit JSON per chunk.
  version     Print the build version.
  help        Print this help.

Run "pii-guard <command> -h" for command-specific flags.`

var version = "0.2.0-dev"

func main() {
	if len(os.Args) == 1 {
		if hasPipedInput() {
			os.Exit(runScan([]string{}))
		}
		fmt.Println(usage)
		os.Exit(0)
	}
	switch os.Args[1] {
	case "-h", "--help", "help":
		fmt.Println(usage)
	case "version", "--version", "-v":
		fmt.Println("pii-guard", version)
	case "scan":
		os.Exit(runScan(os.Args[2:]))
	case "anonymize":
		os.Exit(runAnonymize(os.Args[2:]))
	case "explain":
		os.Exit(runExplain(os.Args[2:]))
	case "stream":
		os.Exit(runStream(os.Args[2:]))
	default:
		os.Exit(runScan(os.Args[1:]))
	}
}

func runScan(args []string) int {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	text := fs.String("text", "", "Text to scan")
	file := fs.String("file", "", "Path to input file")
	format := fs.String("format", "text", "Output format: text|json")
	blockOnSecrets := fs.Bool("block-on-secrets", true, "Block when secrets are detected")
	blockOnPII := fs.Bool("block-on-pii", false, "Block when any PII is detected")
	threshold := fs.Int("block-threshold", 100, "Block when detection count reaches this threshold")
	detectNames := fs.Bool("detect-names", false, "Enable contextual name detection (off by default; see ADR-004)")
	block := fs.Bool("block", false, "Shortcut: --block-on-secrets --block-on-pii (used by pre-commit hook)")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	cfg := pii.DefaultConfig()
	cfg.BlockOnSecrets = *blockOnSecrets
	cfg.BlockOnPII = *blockOnPII
	cfg.BlockThreshold = *threshold
	cfg.DetectNames = *detectNames
	if *block {
		cfg.BlockOnSecrets = true
		cfg.BlockOnPII = true
	}
	d, err := pii.NewDetector(cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	// pre-commit invokes "pii-guard scan --block <files...>" with positional args.
	files := fs.Args()
	if *file != "" {
		files = append(files, *file)
	}
	if *text == "" && len(files) == 0 {
		input, err := resolveInput(*text, "")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
		return scanOne(d, "<stdin>", input, *format)
	}
	if *text != "" {
		return scanOne(d, "<text>", *text, *format)
	}
	exit := 0
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			exit = 1
			continue
		}
		code := scanOne(d, f, string(b), *format)
		if code > exit {
			exit = code
		}
	}
	return exit
}

func scanOne(d *pii.Detector, name, input, format string) int {
	res := d.Scan(context.Background(), input)
	res.AnonymizedText, res.Mappings = pii.AnonymizeWithMap(input, res.Detections)
	if format == "text" && name != "<text>" && name != "<stdin>" {
		fmt.Printf("== %s ==\n", name)
	}
	if err := printResult(&res, format); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if res.Blocked {
		return 2
	}
	return 0
}

func runAnonymize(args []string) int {
	fs := flag.NewFlagSet("anonymize", flag.ContinueOnError)
	text := fs.String("text", "", "Text to anonymize")
	file := fs.String("file", "", "Path to input file")
	format := fs.String("format", "text", "Output format: text|json")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	input, err := resolveInput(*text, *file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	cfg := pii.DefaultConfig()
	cfg.BlockOnPII = false
	cfg.BlockOnSecrets = false
	d := pii.MustNewDetector(cfg)
	res := d.Scan(context.Background(), input)
	res.AnonymizedText, res.Mappings = pii.AnonymizeWithMap(input, res.Detections)

	switch *format {
	case "text":
		fmt.Println(res.AnonymizedText)
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(res); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	default:
		fmt.Fprintln(os.Stderr, "invalid format, expected text|json")
		return 1
	}
	return 0
}

func runExplain(args []string) int {
	fs := flag.NewFlagSet("explain", flag.ContinueOnError)
	text := fs.String("text", "", "Text to explain")
	file := fs.String("file", "", "Path to input file")
	detectNames := fs.Bool("detect-names", false, "Include contextual name pass")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	input, err := resolveInput(*text, *file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	cfg := pii.DefaultConfig()
	cfg.AnonymizeOutput = false
	cfg.DetectNames = *detectNames
	d := pii.MustNewDetector(cfg)
	res := d.Scan(context.Background(), input)

	if len(res.Detections) == 0 {
		fmt.Println("No PII detected.")
		return 0
	}
	fmt.Printf("Found %d detection(s):\n\n", len(res.Detections))
	for i, det := range res.Detections {
		fmt.Printf("  #%d %s\n", i+1, det.Type)
		fmt.Printf("     value      : %q\n", det.Text)
		fmt.Printf("     span       : [%d:%d] (%d bytes)\n", det.Start, det.End, det.End-det.Start)
		fmt.Printf("     confidence : %.2f\n", det.Confidence)
		fmt.Printf("     method     : %s\n", det.Method)
		fmt.Printf("     source     : %s\n\n", det.Source)
	}
	return 0
}

func runStream(args []string) int {
	fs := flag.NewFlagSet("stream", flag.ContinueOnError)
	chunkSize := fs.Int("chunk-size", 2048, "Bytes per chunk")
	overlap := fs.Int("overlap", 128, "Tail overlap between chunks")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	scanner, err := pii.NewStreamScanner(pii.StreamOptions{
		ChunkSize: *chunkSize,
		Overlap:   *overlap,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	chunks, errCh := scanner.Scan(context.Background(), os.Stdin)
	enc := json.NewEncoder(os.Stdout)
	for c := range chunks {
		if err := enc.Encode(c); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	}
	if e, ok := <-errCh; ok && e != nil {
		fmt.Fprintln(os.Stderr, e)
		return 1
	}
	return 0
}

func resolveInput(textFlag, fileFlag string) (string, error) {
	if textFlag != "" {
		return textFlag, nil
	}
	if fileFlag != "" {
		b, err := os.ReadFile(fileFlag) // #nosec G304 — CLI accepts user paths intentionally
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	if hasPipedInput() {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(b)), nil
	}
	return "", errors.New("provide --text or --file")
}

func hasPipedInput() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) == 0
}

func printResult(result *pii.Result, format string) error {
	switch format {
	case "text":
		fmt.Printf("detections: %d\n", result.PIICount)
		fmt.Printf("blocked: %t\n", result.Blocked)
		if result.BlockReason != "" {
			fmt.Printf("reason: %s\n", result.BlockReason)
		}
		fmt.Println("anonymized:")
		fmt.Println(result.AnonymizedText)
		return nil
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	default:
		return errors.New("invalid format, expected text|json")
	}
}
