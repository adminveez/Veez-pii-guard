// Package context implements the second-pass contextual detection
// (currently: person-name detection via tokenizer + dictionary lookup).
//
// See ADR-004 for the rationale (no neural NER, embedded dictionary,
// off-by-default flag, documented false-positive trade-off).
package context

import (
	_ "embed"
	"strings"
	"unicode"
)

//go:embed data/firstnames.txt
var firstNamesRaw string

// FirstNameSet is the set of normalized given names embedded at build time.
// Lookup is case-insensitive against the lowercased form.
var FirstNameSet = func() map[string]struct{} {
	out := make(map[string]struct{}, 1024)
	for _, line := range strings.Split(firstNamesRaw, "\n") {
		name := strings.TrimSpace(strings.ToLower(line))
		if name == "" {
			continue
		}
		out[name] = struct{}{}
	}
	return out
}()

// Token is one word with its byte offsets into the original string.
type Token struct {
	Text  string
	Start int
	End   int
}

// Tokenize splits text into Unicode-aware word tokens with their byte positions.
// Whitespace and punctuation act as separators; hyphenated names like
// "Jean-Claude" are emitted as a single token.
func Tokenize(text string) []Token {
	var tokens []Token
	start := -1
	bytes := []byte(text)
	i := 0
	for i < len(bytes) {
		r, size := decodeRune(bytes[i:])
		if isWordRune(r) {
			if start < 0 {
				start = i
			}
		} else {
			if start >= 0 {
				tokens = append(tokens, Token{Text: text[start:i], Start: start, End: i})
				start = -1
			}
		}
		i += size
	}
	if start >= 0 {
		tokens = append(tokens, Token{Text: text[start:], Start: start, End: len(text)})
	}
	return tokens
}

func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || r == '-' || r == '\''
}

func decodeRune(b []byte) (r rune, size int) {
	for _, ru := range string(b) {
		return ru, len(string(ru))
	}
	return 0, 1
}

// IsCapitalized returns true if the first rune is an uppercase letter.
func IsCapitalized(s string) bool {
	for _, r := range s {
		return unicode.IsUpper(r)
	}
	return false
}

// IsLikelyFirstName looks up the lowercased token in the embedded dictionary.
func IsLikelyFirstName(s string) bool {
	_, ok := FirstNameSet[strings.ToLower(s)]
	return ok
}
