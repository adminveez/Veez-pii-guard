package frenchssn

import "testing"

func TestFrenchSSN_DetectsKnownGoodPattern(t *testing.T) {
	p := New()
	hits := p.Detect("My NIR is 180127505012345 thanks.")
	if len(hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(hits))
	}
}

func TestFrenchSSN_NoMatchOnRandomDigits(t *testing.T) {
	p := New()
	hits := p.Detect("Order 12345 confirmed.")
	if len(hits) != 0 {
		t.Errorf("expected 0 hits, got %d", len(hits))
	}
}

func TestFrenchSSN_Metadata(t *testing.T) {
	p := New()
	if p.Name() != "veez/frenchssn" {
		t.Errorf("unexpected name: %s", p.Name())
	}
	if p.Confidence() < 0.9 {
		t.Errorf("confidence should be high: %f", p.Confidence())
	}
}
