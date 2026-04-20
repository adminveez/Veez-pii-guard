package pii

import (
	"context"
	"errors"
	"testing"
)

type fakePlugin struct {
	name string
	conf float64
	hits []Match
}

func (f fakePlugin) Name() string          { return f.name }
func (f fakePlugin) Confidence() float64   { return f.conf }
func (f fakePlugin) Detect(string) []Match { return f.hits }

type panicPlugin struct{}

func (panicPlugin) Name() string          { return "panic-plugin" }
func (panicPlugin) Confidence() float64   { return 0.5 }
func (panicPlugin) Detect(string) []Match { panic("oops") }

func TestRegister_RejectsInvalid(t *testing.T) {
	d := MustNewDetector(DefaultConfig())
	if err := d.Register(nil); err == nil {
		t.Error("expected error for nil plugin")
	}
	if err := d.Register(fakePlugin{name: ""}); err == nil {
		t.Error("expected error for empty name")
	}
	if err := d.Register(fakePlugin{name: "x", conf: 1.5}); err == nil {
		t.Error("expected error for confidence > 1")
	}
	if err := d.Register(fakePlugin{name: "x", conf: -0.1}); err == nil {
		t.Error("expected error for negative confidence")
	}
}

func TestRegister_RejectsDuplicate(t *testing.T) {
	d := MustNewDetector(DefaultConfig())
	if err := d.Register(fakePlugin{name: "dup", conf: 0.9}); err != nil {
		t.Fatal(err)
	}
	err := d.Register(fakePlugin{name: "dup", conf: 0.9})
	if err == nil || !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("expected ErrInvalidConfig duplicate, got %v", err)
	}
}

func TestPlugin_DetectionsAreReported(t *testing.T) {
	cfg := Config{AnonymizeOutput: false}
	d := MustNewDetector(cfg)
	_ = d.Register(fakePlugin{
		name: "test/marker",
		conf: 0.88,
		hits: []Match{{Type: "MARKER", Start: 0, End: 4, Text: "TEST"}},
	})
	got := d.Scan(context.Background(), "TEST hello")
	if got.PIICount != 1 {
		t.Fatalf("expected 1 detection, got %d", got.PIICount)
	}
	if got.Detections[0].Source != "test/marker" {
		t.Errorf("expected source=test/marker, got %s", got.Detections[0].Source)
	}
	if got.Detections[0].Confidence != 0.88 {
		t.Errorf("expected plugin confidence, got %f", got.Detections[0].Confidence)
	}
}

func TestPlugin_PanicIsRecovered(t *testing.T) {
	d := MustNewDetector(DefaultConfig())
	_ = d.Register(panicPlugin{})
	got := d.Scan(context.Background(), "harmless text")
	// Built-in detectors still run; panic plugin contributes 0.
	if got.Blocked {
		t.Errorf("scan should not be blocked by panicking plugin: %+v", got)
	}
}

func TestPlugin_RejectsOutOfBoundsMatches(t *testing.T) {
	d := MustNewDetector(Config{AnonymizeOutput: false})
	_ = d.Register(fakePlugin{
		name: "bad",
		conf: 0.5,
		hits: []Match{{Type: "X", Start: -1, End: 99999, Text: "??"}},
	})
	got := d.Scan(context.Background(), "short")
	if got.PIICount != 0 {
		t.Errorf("expected 0 detections, got %d", got.PIICount)
	}
}
