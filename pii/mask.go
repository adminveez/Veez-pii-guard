package pii

import "strings"

// Mask partially masks a PII value for safe display.
func Mask(value string, piiType Type) string {
	switch piiType {
	case TypeEmail:
		parts := strings.Split(value, "@")
		if len(parts) != 2 {
			return "***"
		}
		return maskString(parts[0], 1, 0) + "@" + maskString(parts[1], 1, 4)
	case TypePhone, TypePhoneE164, TypeIBAN, TypeCreditCard:
		if len(value) < 8 {
			return "***"
		}
		return value[:4] + strings.Repeat("*", len(value)-8) + value[len(value)-4:]
	default:
		return maskString(value, 2, 2)
	}
}

func maskString(s string, keepStart, keepEnd int) string {
	if keepStart < 0 {
		keepStart = 0
	}
	if keepEnd < 0 {
		keepEnd = 0
	}
	if len(s) <= keepStart+keepEnd {
		return strings.Repeat("*", len(s))
	}
	return s[:keepStart] + strings.Repeat("*", len(s)-keepStart-keepEnd) + s[len(s)-keepEnd:]
}
