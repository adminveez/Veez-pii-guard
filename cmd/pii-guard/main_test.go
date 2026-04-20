package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveInputText(t *testing.T) {
	got, err := resolveInput("hello", "")
	if err != nil {
		t.Fatalf("resolveInput text error: %v", err)
	}
	if got != "hello" {
		t.Fatalf("unexpected text: %q", got)
	}
}

func TestResolveInputFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "input.txt")
	if err := os.WriteFile(path, []byte("abc"), 0o600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	got, err := resolveInput("", path)
	if err != nil {
		t.Fatalf("resolveInput file error: %v", err)
	}
	if got != "abc" {
		t.Fatalf("unexpected content: %q", got)
	}
}

func TestRunScanAndAnonymize(t *testing.T) {
	code := runScan([]string{"--text", "Contact john@example.com", "--format", "json"})
	if code != 0 {
		t.Fatalf("runScan returned code %d", code)
	}

	code = runAnonymize([]string{"--text", "Contact john@example.com", "--format", "json"})
	if code != 0 {
		t.Fatalf("runAnonymize returned code %d", code)
	}
}

func TestRunAnonymizeInvalidFormat(t *testing.T) {
	code := runAnonymize([]string{"--text", "hello", "--format", "xml"})
	if code != 1 {
		t.Fatalf("expected code 1, got %d", code)
	}
}

func TestRunScanBlocked(t *testing.T) {
	code := runScan([]string{"--text", "password=superSecret1234", "--block-on-secrets=true"})
	if code != 2 {
		t.Fatalf("expected blocked code 2, got %d", code)
	}
}

func TestRunScanInvalidFormat(t *testing.T) {
	code := runScan([]string{"--text", "hello", "--format", "xml"})
	if code != 1 {
		t.Fatalf("expected invalid format code 1, got %d", code)
	}
}

func TestResolveInputMissing(t *testing.T) {
	_, err := resolveInput("", "")
	if err == nil {
		t.Fatal("expected error when no input")
	}
}

func TestResolveInputFromPipe(t *testing.T) {
	origStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe error: %v", err)
	}
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = origStdin
		_ = r.Close()
	})

	_, _ = w.WriteString("from-pipe@example.com")
	_ = w.Close()

	got, err := resolveInput("", "")
	if err != nil {
		t.Fatalf("resolve from pipe error: %v", err)
	}
	if got != "from-pipe@example.com" {
		t.Fatalf("unexpected piped input: %q", got)
	}
}

func TestHasPipedInput(t *testing.T) {
	origStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe error: %v", err)
	}
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = origStdin
		_ = r.Close()
		_ = w.Close()
	})

	if !hasPipedInput() {
		t.Fatal("expected stdin pipe to be detected")
	}
}

func TestUsageConstant(t *testing.T) {
	if usage == "" {
		t.Fatal("usage string must not be empty")
	}
}
