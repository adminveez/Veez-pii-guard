// Package siretsiren detects French SIRET (14 digits) and SIREN (9 digits)
// company identifiers and validates them with the Luhn checksum on the digit
// stream (mod-10). SIRET is SIREN + 5-digit NIC; we report both.
package siretsiren

import (
	"regexp"

	"github.com/veez-ai/veez-pii-guard/pii"
)

var (
	siretRE = regexp.MustCompile(`\b\d{3}[\s.-]?\d{3}[\s.-]?\d{3}[\s.-]?\d{5}\b`)
	sirenRE = regexp.MustCompile(`\b\d{3}[\s.-]?\d{3}[\s.-]?\d{3}\b`)
)

// Plugin detects SIRET/SIREN.
type Plugin struct{}

// New returns the plugin.
func New() *Plugin { return &Plugin{} }

// Name identifies this detector.
func (Plugin) Name() string { return "veez/siret-siren" }

// Confidence is moderate; raw 9/14-digit groups have some false-positive surface.
func (Plugin) Confidence() float64 { return 0.85 }

// Detect returns matches for both SIRET (preferred) and SIREN.
func (Plugin) Detect(text string) []pii.Match {
	out := make([]pii.Match, 0, 4)
	taken := map[int]bool{}

	for _, m := range siretRE.FindAllStringIndex(text, -1) {
		raw := stripSep(text[m[0]:m[1]])
		if len(raw) != 14 || !luhnMod10(raw) {
			continue
		}
		out = append(out, pii.Match{
			Type:  pii.TypeSIRET,
			Start: m[0],
			End:   m[1],
			Text:  text[m[0]:m[1]],
		})
		for i := m[0]; i < m[1]; i++ {
			taken[i] = true
		}
	}
	for _, m := range sirenRE.FindAllStringIndex(text, -1) {
		// Skip if already covered by a SIRET match.
		if taken[m[0]] {
			continue
		}
		raw := stripSep(text[m[0]:m[1]])
		if len(raw) != 9 || !luhnMod10(raw) {
			continue
		}
		out = append(out, pii.Match{
			Type:  pii.TypeSIREN,
			Start: m[0],
			End:   m[1],
			Text:  text[m[0]:m[1]],
		})
	}
	return out
}

func stripSep(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			out = append(out, c)
		}
	}
	return string(out)
}

func luhnMod10(digits string) bool {
	sum := 0
	alt := false
	for i := len(digits) - 1; i >= 0; i-- {
		d := int(digits[i] - '0')
		if alt {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		alt = !alt
	}
	return sum%10 == 0
}
