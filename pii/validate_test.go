package pii

import "testing"

func TestValidateLuhn(t *testing.T) {
	if !ValidateLuhn("4111111111111111") {
		t.Fatal("expected valid visa")
	}
	if ValidateLuhn("4111111111111112") {
		t.Fatal("expected invalid card")
	}
}

func TestLooksLikeAPIKey(t *testing.T) {
	if !LooksLikeAPIKey("AKIAIOSFODNN7EXAMPLE") {
		t.Fatal("expected aws-like key to match")
	}
	if LooksLikeAPIKey("hello") {
		t.Fatal("did not expect plain word to match")
	}
}
