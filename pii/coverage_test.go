package pii

import (
	"context"
	"errors"
	"testing"
)

func TestMask_Email(t *testing.T) {
	out := Mask("alice@example.com", TypeEmail)
	if out == "alice@example.com" || out == "" {
		t.Errorf("mask should partially obscure, got %q", out)
	}
}

func TestMask_TooShort(t *testing.T) {
	if Mask("ab", TypeEmail) != "***" {
		t.Errorf("expected *** for invalid email")
	}
	if Mask("123", TypeCreditCard) != "***" {
		t.Errorf("expected *** for short card")
	}
}

func TestMask_CreditCard(t *testing.T) {
	out := Mask("4111111111111111", TypeCreditCard)
	if len(out) != 16 {
		t.Errorf("mask length should be preserved: %s", out)
	}
}

func TestMask_DefaultType(t *testing.T) {
	out := Mask("hello world", TypeContractRef)
	if out == "hello world" {
		t.Errorf("expected masking, got %q", out)
	}
}

func TestPlugins_Snapshot(t *testing.T) {
	d := MustNewDetector(DefaultConfig())
	if got := d.Plugins(); len(got) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(got))
	}
	_ = d.Register(fakePlugin{name: "p1", conf: 0.5})
	_ = d.Register(fakePlugin{name: "p2", conf: 0.6})
	got := d.Plugins()
	if len(got) != 2 {
		t.Errorf("expected 2 plugins, got %d", len(got))
	}
}

func TestNewDetector_RejectsInvalidConfig(t *testing.T) {
	cfg := DefaultConfig()
	cfg.BlockThreshold = -1
	if _, err := NewDetector(cfg); err == nil || !errors.Is(err, ErrInvalidConfig) {
		t.Errorf("expected ErrInvalidConfig, got %v", err)
	}
}

func TestMustNewDetector_PanicsOnBadConfig(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	cfg := DefaultConfig()
	cfg.BlockThreshold = -1
	_ = MustNewDetector(cfg)
}

func TestApplyBlocking_PIIThreshold(t *testing.T) {
	cfg := DefaultConfig()
	cfg.BlockOnPII = true
	d := MustNewDetector(cfg)
	r := d.Scan(context.Background(), "alice@x.com bob@y.io carol@z.net")
	if !r.Blocked {
		t.Error("expected block on PII")
	}
}
