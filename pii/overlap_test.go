package pii

import "testing"

func TestResolveOverlaps_KeepsHighestConfidence(t *testing.T) {
	in := []Detection{
		{Type: TypeEmail, Start: 0, End: 10, Confidence: 0.95},
		{Type: TypeSecret, Start: 5, End: 15, Confidence: 0.99},
	}
	out := resolveOverlaps(in)
	if len(out) != 1 {
		t.Fatalf("expected 1, got %d", len(out))
	}
	if out[0].Type != TypeSecret {
		t.Errorf("expected SECRET to win, got %s", out[0].Type)
	}
}

func TestResolveOverlaps_TieBreakOnLength(t *testing.T) {
	in := []Detection{
		{Type: TypeEmail, Start: 0, End: 5, Confidence: 0.9},
		{Type: TypePhone, Start: 0, End: 10, Confidence: 0.9},
	}
	out := resolveOverlaps(in)
	if len(out) != 1 || out[0].Type != TypePhone {
		t.Errorf("longer span should win on tie; got %+v", out)
	}
}

func TestResolveOverlaps_NonOverlappingKept(t *testing.T) {
	in := []Detection{
		{Type: TypeEmail, Start: 0, End: 10, Confidence: 0.9},
		{Type: TypePhone, Start: 20, End: 30, Confidence: 0.9},
	}
	out := resolveOverlaps(in)
	if len(out) != 2 {
		t.Errorf("expected 2 detections, got %d", len(out))
	}
}

func TestResolveOverlaps_DropsZeroLength(t *testing.T) {
	in := []Detection{
		{Type: TypeEmail, Start: 5, End: 5, Confidence: 0.9},
		{Type: TypePhone, Start: 0, End: 10, Confidence: 0.9},
	}
	out := resolveOverlaps(in)
	if len(out) != 1 || out[0].Type != TypePhone {
		t.Errorf("zero-length match should be dropped; got %+v", out)
	}
}

func TestResolveOverlaps_DeterministicOnDoubleTie(t *testing.T) {
	in := []Detection{
		{Type: TypePhone, Start: 0, End: 10, Confidence: 0.9},
		{Type: TypeEmail, Start: 0, End: 10, Confidence: 0.9},
	}
	a := resolveOverlaps(in)
	b := resolveOverlaps(in)
	if len(a) != 1 || len(b) != 1 || a[0].Type != b[0].Type {
		t.Errorf("non-deterministic tie-break: %v vs %v", a, b)
	}
}
