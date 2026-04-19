package pii

import (
	"context"
	"strings"
	"sync"
	"testing"
)

func buildWords(count int) string {
	var builder strings.Builder
	for i := 0; i < count; i++ {
		builder.WriteString("mot ")
	}
	return builder.String()
}

func BenchmarkScanShort100Words(b *testing.B) {
	d := NewDetector(DefaultConfig())
	text := buildWords(100) + " contact: bench.short@example.com"
	ctx := context.Background()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = d.Scan(ctx, text)
	}
}

func BenchmarkScanMedium1000Words(b *testing.B) {
	d := NewDetector(DefaultConfig())
	text := buildWords(1000) + " contact: bench.medium@example.com"
	ctx := context.Background()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = d.Scan(ctx, text)
	}
}

func BenchmarkScanLong10000Words(b *testing.B) {
	d := NewDetector(DefaultConfig())
	text := buildWords(10000) + " contact: bench.long@example.com"
	ctx := context.Background()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = d.Scan(ctx, text)
	}
}

func BenchmarkScanParallel1000Texts(b *testing.B) {
	d := NewDetector(DefaultConfig())
	texts := make([]string, 1000)
	for i := range texts {
		texts[i] = buildWords(100) + " contact: parallel@example.com"
	}
	ctx := context.Background()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(len(texts))
		for _, txt := range texts {
			go func(s string) {
				defer wg.Done()
				_ = d.Scan(ctx, s)
			}(txt)
		}
		wg.Wait()
	}
}
