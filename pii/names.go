package pii

import (
	piictx "github.com/adminveez/Veez-pii-guard/pii/context"
)

// runNamePass implements the contextual second-pass name detector described
// in ADR-004. Off by default; opt in via Config.DetectNames.
//
// Heuristic: a token is a candidate person name if it is capitalized AND found
// in the embedded given-name dictionary. Confidence is 0.65 when isolated and
// 0.85 when followed by another capitalized token (presumed surname). The
// span returned covers the first name token; if a surname follows it is
// included in the span.
func (d *Detector) runNamePass(text string, result *Result) {
	tokens := piictx.Tokenize(text)
	for i, tok := range tokens {
		if !piictx.IsCapitalized(tok.Text) {
			continue
		}
		if !piictx.IsLikelyFirstName(tok.Text) {
			continue
		}
		end := tok.End
		conf := 0.65
		if i+1 < len(tokens) {
			next := tokens[i+1]
			if piictx.IsCapitalized(next.Text) && !piictx.IsLikelyFirstName(next.Text) {
				end = next.End
				conf = 0.85
			}
		}
		result.Detections = append(result.Detections, Detection{
			Type:           TypePersonName,
			Text:           text[tok.Start:end],
			Start:          tok.Start,
			End:            end,
			Confidence:     conf,
			Method:         "contextual",
			Source:         "context/names",
			RequiresReview: true,
		})
	}
}
