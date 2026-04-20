package pii

import (
	"context"
	"strings"
	"testing"
)

func TestNamePass_DisabledByDefault(t *testing.T) {
	d := MustNewDetector(DefaultConfig())
	res := d.Scan(context.Background(), "Marie Dupont travaille ici.")
	for _, det := range res.Detections {
		if det.Type == TypePersonName {
			t.Errorf("DetectNames default should be off; got %+v", det)
		}
	}
}

func TestNamePass_DetectsKnownFirstNameWithSurname(t *testing.T) {
	cfg := DefaultConfig()
	cfg.DetectNames = true
	d := MustNewDetector(cfg)
	res := d.Scan(context.Background(), "Marie Dupont travaille ici.")
	found := false
	for _, det := range res.Detections {
		if det.Type == TypePersonName && strings.Contains(det.Text, "Marie") {
			found = true
			if det.Confidence < 0.8 {
				t.Errorf("expected ≥0.85 confidence with surname, got %f", det.Confidence)
			}
		}
	}
	if !found {
		t.Fatalf("expected PERSON_NAME detection, got %+v", res.Detections)
	}
}

func TestNamePass_LowerConfidenceForSoloFirstName(t *testing.T) {
	cfg := Config{DetectNames: true, AnonymizeOutput: false}
	d := MustNewDetector(cfg)
	res := d.Scan(context.Background(), "Marie est partie.")
	for _, det := range res.Detections {
		if det.Type == TypePersonName && det.Confidence > 0.7 {
			t.Errorf("solo first name should have ≤0.65 confidence, got %f", det.Confidence)
		}
	}
}

func TestNamePass_IgnoresLowercase(t *testing.T) {
	cfg := Config{DetectNames: true, AnonymizeOutput: false}
	d := MustNewDetector(cfg)
	res := d.Scan(context.Background(), "marie dupont")
	for _, det := range res.Detections {
		if det.Type == TypePersonName {
			t.Errorf("lowercase tokens must not match; got %+v", det)
		}
	}
}
