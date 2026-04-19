package pii

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

func generateText(words int, piiSuffix string) string {
	var b strings.Builder
	for i := 0; i < words; i++ {
		b.WriteString("mot ")
	}
	b.WriteString(piiSuffix)
	return b.String()
}

func TestPerformanceThresholds(t *testing.T) {
	d := NewDetector(DefaultConfig())
	ctx := context.Background()

	tests := []struct {
		name      string
		text      string
		threshold time.Duration
	}{
		{name: "short_100_words", text: generateText(100, " email: short@example.com"), threshold: time.Millisecond},
		{name: "medium_1000_words", text: generateText(1000, " email: medium@example.com"), threshold: 5 * time.Millisecond},
		{name: "long_10000_words", text: generateText(10000, " email: long@example.com"), threshold: 50 * time.Millisecond},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			start := time.Now()
			_ = d.Scan(ctx, tc.text)
			elapsed := time.Since(start)
			if elapsed > tc.threshold {
				t.Fatalf("threshold exceeded: %s > %s", elapsed, tc.threshold)
			}
		})
	}

	t.Run("parallel_1000_texts_under_2s", func(t *testing.T) {
		texts := make([]string, 1000)
		for i := range texts {
			texts[i] = generateText(100, " email: parallel@example.com")
		}

		start := time.Now()
		var wg sync.WaitGroup
		wg.Add(len(texts))
		for _, txt := range texts {
			go func(s string) {
				defer wg.Done()
				_ = d.Scan(ctx, s)
			}(txt)
		}
		wg.Wait()
		elapsed := time.Since(start)
		if elapsed > 2*time.Second {
			t.Fatalf("parallel threshold exceeded: %s > 2s", elapsed)
		}
	})
}
