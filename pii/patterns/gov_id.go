package patterns

import "regexp"

// FrenchSSN matches the French INSEE social-security number (NIR), accounting for
// Corsican department codes (2A/2B). Note: this is also exposed as a built-in
// PatternPlugin under pii/plugins/frenchssn for documentation purposes.
var FrenchSSN = Compiled{
	Name:       "FRENCH_SSN",
	Source:     "regex/french-ssn",
	Confidence: 0.98,
	Regexp:     regexp.MustCompile(`\b[12][0-9]{2}(?:0[1-9]|1[0-2])(?:(?:0[1-9]|[1-8][0-9]|9[0-5])|2[AB])[0-9]{3}[0-9]{3}(?:\s?[0-9]{2})?\b`),
}
