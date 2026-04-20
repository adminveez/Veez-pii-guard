package pii

import (
	"fmt"
	"sort"
)

// Anonymize replaces detected PII with semantic placeholders.
//
// The mapping is deterministic and stable: the i-th occurrence of a given
// type T gets the placeholder [T_i] (or a type-specific shape, see
// semanticPlaceholder). For reversible anonymization, use AnonymizeWithMap.
func Anonymize(text string, detections []Detection) string {
	if len(detections) == 0 {
		return text
	}

	filtered := resolveOverlaps(detections)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Start < filtered[j].Start
	})

	typeCounter := map[Type]int{}
	placeholders := make([]string, len(filtered))
	for i, d := range filtered {
		typeCounter[d.Type]++
		placeholders[i] = semanticPlaceholder(d.Type, typeCounter[d.Type])
	}

	result := text
	for i := len(filtered) - 1; i >= 0; i-- {
		d := filtered[i]
		if d.Start < 0 || d.End > len(result) || d.Start >= d.End {
			continue
		}
		result = result[:d.Start] + placeholders[i] + result[d.End:]
	}
	return result
}

func semanticPlaceholder(piiType Type, index int) string {
	switch piiType {
	case TypeContractRef:
		return "[CONTRACT_REF]"
	case TypeClientID:
		return "[CLIENT_ID]"
	case TypeCaseRef:
		return fmt.Sprintf("[CASE_REF_%d]", index)
	default:
		return fmt.Sprintf("[%s_%d]", piiType, index)
	}
}
