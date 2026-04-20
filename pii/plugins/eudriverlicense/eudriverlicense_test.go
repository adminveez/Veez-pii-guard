package eudriverlicense

import "testing"

func TestEUDriverLicense_DetectsITFormat(t *testing.T) {
	hits := New().Detect("License: UV1234567A")
	if len(hits) == 0 {
		t.Fatalf("expected IT-format match, got 0")
	}
}

func TestEUDriverLicense_DetectsESFormat(t *testing.T) {
	hits := New().Detect("Permiso 12345678X confirmado.")
	if len(hits) == 0 {
		t.Fatalf("expected ES-format match, got 0")
	}
}

func TestEUDriverLicense_NoMatchOnRandomText(t *testing.T) {
	hits := New().Detect("hello world")
	if len(hits) != 0 {
		t.Errorf("expected 0 hits, got %d", len(hits))
	}
}
