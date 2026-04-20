// Package frenchssn is a built-in PatternPlugin example detecting the French
// social-security number (NIR). It mirrors the built-in regex but is exposed
// as a plugin to serve as a 30-line template for third-party authors.
package frenchssn

import (
	"regexp"

	"github.com/adminveez/Veez-pii-guard/pii"
)

var re = regexp.MustCompile(`\b[12][0-9]{2}(?:0[1-9]|1[0-2])(?:(?:0[1-9]|[1-8][0-9]|9[0-5])|2[AB])[0-9]{3}[0-9]{3}(?:\s?[0-9]{2})?\b`)

// Plugin is the registered detector instance.
type Plugin struct{}

// New returns a ready-to-register plugin.
func New() *Plugin { return &Plugin{} }

// Name returns the plugin identifier.
func (Plugin) Name() string { return "veez/frenchssn" }

// Confidence is high — NIR has strict structural rules.
func (Plugin) Confidence() float64 { return 0.97 }

// Detect returns SSN matches in text.
func (Plugin) Detect(text string) []pii.Match {
	idx := re.FindAllStringIndex(text, -1)
	out := make([]pii.Match, 0, len(idx))
	for _, m := range idx {
		out = append(out, pii.Match{
			Type:  pii.TypeFrenchSSN,
			Start: m[0],
			End:   m[1],
			Text:  text[m[0]:m[1]],
		})
	}
	return out
}
