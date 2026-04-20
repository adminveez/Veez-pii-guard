package pii

import "sort"

// resolveOverlaps removes overlapping detections, keeping higher-confidence
// matches first; on ties the longer span wins. This is the single source of
// truth for overlap arbitration; see overlap_test.go for invariants.
//
// Stable in the input order for deterministic output across runs.
func resolveOverlaps(detections []Detection) []Detection {
	if len(detections) <= 1 {
		out := make([]Detection, len(detections))
		copy(out, detections)
		return out
	}

	sorted := make([]Detection, len(detections))
	copy(sorted, detections)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Confidence != sorted[j].Confidence {
			return sorted[i].Confidence > sorted[j].Confidence
		}
		li := sorted[i].End - sorted[i].Start
		lj := sorted[j].End - sorted[j].Start
		if li != lj {
			return li > lj
		}
		// Final tiebreaker for determinism.
		if sorted[i].Start != sorted[j].Start {
			return sorted[i].Start < sorted[j].Start
		}
		return sorted[i].Type < sorted[j].Type
	})

	chosen := make([]Detection, 0, len(sorted))
	for _, candidate := range sorted {
		if candidate.End <= candidate.Start {
			continue
		}
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
