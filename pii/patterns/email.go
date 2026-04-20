package patterns

import "regexp"

// Email matches RFC-5322-ish email addresses. Conservative on the local part
// (no quoted strings) to keep the false-positive rate low in prose.
var Email = Compiled{
	Name:       "EMAIL",
	Source:     "regex/email",
	Confidence: 0.95,
	Regexp:     regexp.MustCompile(`(?i)[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`),
}
