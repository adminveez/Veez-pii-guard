// Package eudriverlicense detects driver-license number formats from major EU
// jurisdictions. Patterns are intentionally conservative — license formats
// vary by issuing authority and year, and false positives matter for a tool
// running on prompts.
//
// Coverage: FR (12 digits, post-2013 format), DE (alphanumeric 11 chars),
// IT (UV + 7 digits + L), ES (8 digits + letter).
package eudriverlicense

import (
	"regexp"

	"github.com/veez-ai/veez-pii-guard/pii"
)

var detectors = []*regexp.Regexp{
	regexp.MustCompile(`\b\d{12}\b`),                   // FR (12 digits, ambiguous → low confidence)
	regexp.MustCompile(`\b[A-Z0-9]{2}\d{6}[A-Z0-9]{3}\b`), // DE-ish 11-char form
	regexp.MustCompile(`\b[Uu][Vv]\d{7}[A-Za-z]\b`),    // IT
	regexp.MustCompile(`\b\d{8}[A-Z]\b`),               // ES
}

// Plugin detects EU driver licenses.
type Plugin struct{}

// New returns the plugin.
func New() *Plugin { return &Plugin{} }

// Name identifies this detector.
func (Plugin) Name() string { return "veez/eu-driver-license" }

// Confidence is moderate; license number shapes overlap with other IDs.
func (Plugin) Confidence() float64 { return 0.70 }

// Detect returns matches across the supported jurisdictions.
func (Plugin) Detect(text string) []pii.Match {
	var out []pii.Match
	seen := map[[2]int]bool{}
	for _, re := range detectors {
		for _, m := range re.FindAllStringIndex(text, -1) {
			k := [2]int{m[0], m[1]}
			if seen[k] {
				continue
			}
			seen[k] = true
			out = append(out, pii.Match{
				Type:  pii.TypeEUDriverLicense,
				Start: m[0],
				End:   m[1],
				Text:  text[m[0]:m[1]],
			})
		}
	}
	return out
}
