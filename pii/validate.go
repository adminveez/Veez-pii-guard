package pii

import "regexp"

// LooksLikeAPIKey checks if a value resembles common API key formats.
func LooksLikeAPIKey(value string) bool {
	patterns := []string{
		`^sk-[a-zA-Z0-9]{20,}$`,
		`^[a-f0-9]{32}$`,
		`^[A-Za-z0-9_-]{20,}$`,
		`^AIza[A-Za-z0-9_-]{35}$`,
		`^AKIA[A-Z0-9]{16}$`,
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, value); matched {
			return true
		}
	}

	return false
}

// ValidateLuhn returns true when the card number passes the Luhn checksum.
func ValidateLuhn(cardNumber string) bool {
	cleaned := make([]rune, 0, len(cardNumber))
	for _, r := range cardNumber {
		switch r {
		case ' ', '-':
			continue
		default:
			if r < '0' || r > '9' {
				return false
			}
			cleaned = append(cleaned, r)
		}
	}

	if len(cleaned) < 13 || len(cleaned) > 19 {
		return false
	}

	sum := 0
	alternate := false
	for i := len(cleaned) - 1; i >= 0; i-- {
		digit := int(cleaned[i] - '0')
		if alternate {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}
