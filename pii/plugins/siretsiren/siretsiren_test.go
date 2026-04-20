package siretsiren

import "testing"

func TestSIRET_ValidLuhnPasses(t *testing.T) {
	// 73282932000074 — known valid SIRET (CRDA Lyon).
	hits := New().Detect("Société: 732 829 320 00074")
	if len(hits) == 0 {
		t.Fatalf("expected a SIRET hit, got 0")
	}
}

func TestSIRET_InvalidLuhnRejected(t *testing.T) {
	hits := New().Detect("Société: 111 111 111 11111")
	for _, h := range hits {
		if h.Type == "SIRET" {
			t.Errorf("invalid Luhn must be rejected, got %+v", h)
		}
	}
}

func TestPlugin_Metadata(t *testing.T) {
	p := New()
	if p.Name() == "" {
		t.Error("name must be non-empty")
	}
	c := p.Confidence()
	if c < 0 || c > 1 {
		t.Errorf("confidence out of range: %f", c)
	}
}
