// Package patterns holds the regex families used by the default Go engine.
// Each file groups conceptually related patterns. Patterns are package-level
// vars compiled at init time; this is acceptable because the literals are
// known at compile time (see ADR-001).
package patterns

import "regexp"

// Compiled is one compiled pattern with metadata used by the engine.
type Compiled struct {
	Name       string
	Source     string // family name, surfaces in Detection.Source
	Confidence float64
	Regexp     *regexp.Regexp
}
