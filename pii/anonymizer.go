package pii

import (
	"fmt"
	"sort"
)

// Anonymize replaces detected PII with semantic placeholders.
func Anonymize(text string, detections []Detection) string {
	if len(detections) == 0 {
		return text
	}

	filtered := removeOverlapsByPriority(detections)
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

func removeOverlapsByPriority(detections []Detection) []Detection {
	out := make([]Detection, len(detections))
	copy(out, detections)

	// Higher confidence first; if equal, longer match first.
	sort.Slice(out, func(i, j int) bool {
		if out[i].Confidence == out[j].Confidence {
			return (out[i].End - out[i].Start) > (out[j].End - out[j].Start)
		}
		return out[i].Confidence > out[j].Confidence
	})

	chosen := make([]Detection, 0, len(out))
	for _, candidate := range out {
		overlaps := false
		for _, kept := range chosen {
			if candidate.Start < kept.End && kept.Start < candidate.End {
				overlaps = true
				break
			}
		}
		if !overlaps {
			chosen = append(chosen, candidate)
		}
	}

	return chosen
}
