package context

import "testing"

func TestTokenize_Basic(t *testing.T) {
	toks := Tokenize("Hello, world! Jean-Claude est là.")
	if len(toks) < 4 {
		t.Fatalf("expected >=4 tokens, got %d", len(toks))
	}
	// "Jean-Claude" should stay as one token thanks to hyphen handling.
	found := false
	for _, tk := range toks {
		if tk.Text == "Jean-Claude" {
			found = true
		}
	}
	if !found {
		t.Errorf("hyphenated token not preserved: %+v", toks)
	}
}

func TestTokenize_OffsetsAreByteAccurate(t *testing.T) {
	text := "Marie est ici"
	toks := Tokenize(text)
	for _, tk := range toks {
		if text[tk.Start:tk.End] != tk.Text {
			t.Errorf("offset mismatch for %q: %d-%d", tk.Text, tk.Start, tk.End)
		}
	}
}

func TestIsCapitalized(t *testing.T) {
	if !IsCapitalized("Marie") {
		t.Error("Marie should be capitalized")
	}
	if IsCapitalized("marie") {
		t.Error("marie should not be capitalized")
	}
	if IsCapitalized("") {
		t.Error("empty should not be capitalized")
	}
}

func TestIsLikelyFirstName(t *testing.T) {
	if !IsLikelyFirstName("Marie") {
		t.Error("Marie should be in dictionary")
	}
	if !IsLikelyFirstName("john") {
		t.Error("john should be in dictionary")
	}
	if IsLikelyFirstName("Quetzalcoatl") {
		t.Error("Quetzalcoatl should not be in dictionary")
	}
}

func TestFirstNameSet_NotEmpty(t *testing.T) {
	if len(FirstNameSet) < 100 {
		t.Errorf("dictionary too small: %d entries", len(FirstNameSet))
	}
}
