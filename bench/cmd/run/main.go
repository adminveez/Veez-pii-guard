// Command run executes the veez-pii-guard benchmark harness against a
// curated dataset and emits JSON + Markdown results.
//
// Methodology (see ADR-007):
//
//   - Dataset: 1000 texts across 5 categories (chat, email, log, ticket, doc)
//   - Each text contains a known set of PII spans (ground truth).
//   - For each engine, we measure:
//   - precision, recall, F1 per type
//   - p50/p95/p99 latency
//   - throughput (chars/sec)
//   - Engines: veez (default, pure-go), veez-rust (opt-in), presidio (docker), spacy (uv).
//
// Usage:
//
//	go run ./bench/cmd/run --engine=veez --out=bench/results/veez.json
//	go run ./bench/cmd/run --engine=veez-rust --out=bench/results/veez-rust.json
//
// The presidio and spacy adapters are external binaries invoked via exec;
// they are optional and skipped when not present.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"time"

	"github.com/veez-ai/veez-pii-guard/bench/dataset"
	"github.com/veez-ai/veez-pii-guard/pii"
)

type result struct {
	Engine     string             `json:"engine"`
	Samples    int                `json:"samples"`
	TotalChars int                `json:"total_chars"`
	WallMs     float64            `json:"wall_ms"`
	P50Ms      float64            `json:"p50_ms"`
	P95Ms      float64            `json:"p95_ms"`
	P99Ms      float64            `json:"p99_ms"`
	Throughput float64            `json:"throughput_chars_per_sec"`
	PerType    map[string]metrics `json:"per_type"`
}

type metrics struct {
	TP        int     `json:"tp"`
	FP        int     `json:"fp"`
	FN        int     `json:"fn"`
	Precision float64 `json:"precision"`
	Recall    float64 `json:"recall"`
	F1        float64 `json:"f1"`
}

func main() {
	engine := flag.String("engine", "veez", "Engine to benchmark: veez|veez-rust|presidio|spacy")
	out := flag.String("out", "", "Path to write JSON results (default: stdout)")
	seed := flag.Int64("seed", 42, "Dataset RNG seed for reproducibility")
	samples := flag.Int("samples", 1000, "Number of synthetic samples to run")
	flag.Parse()

	ds := dataset.Generate(*seed, *samples)

	var run func(ctx context.Context, text string) ([]pii.Detection, error)
	switch *engine {
	case "veez":
		d := pii.MustNewDetector(pii.DefaultConfig())
		run = func(ctx context.Context, text string) ([]pii.Detection, error) {
			return d.Scan(ctx, text).Detections, nil
		}
	case "veez-rust":
		if !pii.RustAvailable() {
			fmt.Fprintln(os.Stderr, "Rust backend not built. Rebuild with: go build -tags veezrust")
			os.Exit(2)
		}
		run = func(_ context.Context, text string) ([]pii.Detection, error) {
			return pii.RustScan(text)
		}
	default:
		fmt.Fprintf(os.Stderr, "engine %q not yet wired (presidio/spacy adapters: see bench/adapters/)\n", *engine)
		os.Exit(2)
	}

	res := benchmark(*engine, ds, run)
	var w io.Writer = os.Stdout
	var closer func() error
	if *out != "" {
		f, err := os.Create(*out)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		closer = f.Close
		w = f
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(res); err != nil {
		if closer != nil {
			_ = closer()
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if closer != nil {
		if err := closer(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func benchmark(engine string, ds []dataset.Sample, run func(context.Context, string) ([]pii.Detection, error)) result {
	latencies := make([]float64, 0, len(ds))
	totalChars := 0
	perType := map[string]*metrics{}

	wallStart := time.Now()
	for _, s := range ds {
		totalChars += len(s.Text)
		t0 := time.Now()
		dets, err := run(context.Background(), s.Text)
		if err != nil {
			fmt.Fprintf(os.Stderr, "scan error: %v\n", err)
			continue
		}
		latencies = append(latencies, float64(time.Since(t0).Microseconds())/1000.0)
		score(perType, s.Truth, dets)
	}
	wall := time.Since(wallStart).Seconds() * 1000.0

	sort.Float64s(latencies)
	finalPerType := make(map[string]metrics, len(perType))
	for k, m := range perType {
		denom := float64(m.TP + m.FP)
		if denom > 0 {
			m.Precision = float64(m.TP) / denom
		}
		denom = float64(m.TP + m.FN)
		if denom > 0 {
			m.Recall = float64(m.TP) / denom
		}
		if m.Precision+m.Recall > 0 {
			m.F1 = 2 * m.Precision * m.Recall / (m.Precision + m.Recall)
		}
		finalPerType[k] = *m
	}

	return result{
		Engine:     engine,
		Samples:    len(ds),
		TotalChars: totalChars,
		WallMs:     wall,
		P50Ms:      pct(latencies, 0.50),
		P95Ms:      pct(latencies, 0.95),
		P99Ms:      pct(latencies, 0.99),
		Throughput: float64(totalChars) / (wall / 1000.0),
		PerType:    finalPerType,
	}
}

func score(per map[string]*metrics, truth []dataset.Span, got []pii.Detection) {
	matched := make([]bool, len(got))
	for _, t := range truth {
		m := per[string(t.Type)]
		if m == nil {
			m = &metrics{}
			per[string(t.Type)] = m
		}
		hit := false
		for i, g := range got {
			if matched[i] {
				continue
			}
			if string(g.Type) == string(t.Type) && overlap(g.Start, g.End, t.Start, t.End) {
				m.TP++
				matched[i] = true
				hit = true
				break
			}
		}
		if !hit {
			m.FN++
		}
	}
	for i, g := range got {
		if matched[i] {
			continue
		}
		key := string(g.Type)
		m := per[key]
		if m == nil {
			m = &metrics{}
			per[key] = m
		}
		m.FP++
	}
}

func overlap(a1, a2, b1, b2 int) bool {
	return a1 < b2 && b1 < a2
}

func pct(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	i := int(float64(len(sorted)-1) * p)
	return sorted[i]
}
