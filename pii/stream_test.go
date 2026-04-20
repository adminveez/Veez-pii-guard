package pii

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestStreamScanner_ParityWithFullScan(t *testing.T) {
	text := "Contact john.doe@example.com or jane@example.org. " +
		strings.Repeat("Lorem ipsum dolor sit amet. ", 200) +
		"Final email: ceo@acme.io"

	full := MustNewDetector(DefaultConfig()).Scan(context.Background(), text)

	scanner, err := NewStreamScanner(StreamOptions{ChunkSize: 256, Overlap: 64})
	if err != nil {
		t.Fatal(err)
	}
	chunks, errCh := scanner.Scan(context.Background(), strings.NewReader(text))
	streamCount := 0
	for c := range chunks {
		streamCount += len(c.Detections)
	}
	if e, ok := <-errCh; ok && e != nil && !errors.Is(e, io.EOF) {
		t.Fatalf("stream error: %v", e)
	}
	if streamCount < full.PIICount {
		t.Errorf("stream missed detections: %d < full %d", streamCount, full.PIICount)
	}
}

func TestStreamScanner_DetectsAcrossBoundary(t *testing.T) {
	// Email straddling the chunk boundary. Use spaces around so the
	// regex anchors cleanly on the local part.
	prefix := strings.Repeat("a ", 125) // 250 bytes ending with a space
	text := prefix + "alice@example.com" + strings.Repeat(" z", 50)

	scanner, err := NewStreamScanner(StreamOptions{ChunkSize: 256, Overlap: 64})
	if err != nil {
		t.Fatal(err)
	}
	chunks, _ := scanner.Scan(context.Background(), strings.NewReader(text))
	found := false
	for c := range chunks {
		for _, d := range c.Detections {
			if d.Type == TypeEmail && strings.Contains(d.Text, "alice@example.com") {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("email straddling chunk boundary was not detected")
	}
}

func TestStreamScanner_RejectsBadOptions(t *testing.T) {
	if _, err := NewStreamScanner(StreamOptions{ChunkSize: 100, Overlap: 100}); err == nil {
		t.Error("expected error: overlap >= chunk")
	}
	if _, err := NewStreamScanner(StreamOptions{ChunkSize: 100, Overlap: -1}); err == nil {
		t.Error("expected error: negative overlap")
	}
}
