package pii

import (
	"context"
	"strings"
	"testing"
)

func FuzzScan(f *testing.F) {
	seeds := []string{
		"",
		"hello world",
		"john@example.com",
		"+33 6 12 34 56 78",
		"4111111111111111",
		"FR1420041010050500013M02606",
		"api_key=ABCDEFGH12345678IJKLMNOP",
		strings.Repeat("a", 4096),
	}
	for _, s := range seeds {
		f.Add(s)
	}
	d := MustNewDetector(DefaultConfig())
	f.Fuzz(func(t *testing.T, s string) {
		_ = d.Scan(context.Background(), s)
	})
}

func FuzzAnonymize(f *testing.F) {
	seeds := []string{"", "x", "alice@x.com bob@y.com", strings.Repeat("e@e.io ", 50)}
	for _, s := range seeds {
		f.Add(s)
	}
	d := MustNewDetector(DefaultConfig())
	f.Fuzz(func(t *testing.T, s string) {
		res := d.Scan(context.Background(), s)
		_ = Anonymize(s, res.Detections)
	})
}

func FuzzReidentify(f *testing.F) {
	seeds := []string{"alice@x.com", "no pii here", "Contact alice@x.com or +33612345678"}
	for _, s := range seeds {
		f.Add(s)
	}
	d := MustNewDetector(DefaultConfig())
	f.Fuzz(func(t *testing.T, s string) {
		res := d.Scan(context.Background(), s)
		anon, mappings := AnonymizeWithMap(s, res.Detections)
		got := Reidentify(anon, mappings)
		// Round-trip must be identity when no PII overlaps survive ambiguity.
		if len(res.Detections) > 0 && got == "" && s != "" {
			t.Errorf("reidentify produced empty for non-empty input: %q -> %q -> %q", s, anon, got)
		}
	})
}

func FuzzStream(f *testing.F) {
	seeds := []string{"", "alice@x.com", strings.Repeat("token ", 200) + "alice@x.com"}
	for _, s := range seeds {
		f.Add(s)
	}
	scanner, err := NewStreamScanner(StreamOptions{ChunkSize: 64, Overlap: 16})
	if err != nil {
		f.Fatal(err)
	}
	f.Fuzz(func(t *testing.T, s string) {
		ch, _ := scanner.Scan(context.Background(), strings.NewReader(s))
		for range ch {
		}
	})
}
