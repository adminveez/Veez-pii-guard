package pii

// AnonymizeWithMap replaces detected PII with placeholders and returns a mapping placeholder->original value.
func AnonymizeWithMap(text string, detections []Detection) (string, map[string]string) {
	if len(detections) == 0 {
		return text, map[string]string{}
	}

	filtered := removeOverlapsByPriority(detections)
	result := text
	mappings := make(map[string]string, len(filtered))

	// First pass to generate deterministic placeholder names in reading order.
	placeholders := buildPlaceholders(filtered)

	// Replace in reverse order to preserve indices.
	for i := len(placeholders) - 1; i >= 0; i-- {
		d := placeholders[i].Detection
		if d.Start < 0 || d.End > len(result) || d.Start >= d.End {
			continue
		}
		result = result[:d.Start] + placeholders[i].Placeholder + result[d.End:]
		mappings[placeholders[i].Placeholder] = d.Text
	}

	return result, mappings
}

// Reidentify replaces placeholders using placeholder->original mappings.
func Reidentify(anonymized string, mappings map[string]string) string {
	result := anonymized
	for placeholder, original := range mappings {
		result = replaceAll(result, placeholder, original)
	}
	return result
}

type placeholderBinding struct {
	Detection   Detection
	Placeholder string
}

func buildPlaceholders(detections []Detection) []placeholderBinding {
	// Reuse helper by creating synthetic anonymized text order from detections.
	ordered := make([]Detection, len(detections))
	copy(ordered, detections)
	// Stable insertion sort by Start to avoid importing sort here.
	for i := 1; i < len(ordered); i++ {
		j := i
		for j > 0 && ordered[j-1].Start > ordered[j].Start {
			ordered[j-1], ordered[j] = ordered[j], ordered[j-1]
			j--
		}
	}
	counts := map[Type]int{}
	out := make([]placeholderBinding, 0, len(ordered))
	for _, d := range ordered {
		counts[d.Type]++
		out = append(out, placeholderBinding{
			Detection:   d,
			Placeholder: semanticPlaceholder(d.Type, counts[d.Type]),
		})
	}
	return out
}

func replaceAll(s, old, new string) string {
	if old == "" {
		return s
	}
	for {
		idx := indexOf(s, old)
		if idx < 0 {
			break
		}
		s = s[:idx] + new + s[idx+len(old):]
	}
	return s
}

func indexOf(s, sub string) int {
	if len(sub) == 0 || len(sub) > len(s) {
		return -1
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
