package patterns

import "regexp"

// IBAN matches a permissive IBAN shape (length-validated downstream).
var IBAN = Compiled{
	Name:       "IBAN",
	Source:     "regex/iban",
	Confidence: 0.98,
	Regexp:     regexp.MustCompile(`(?i)[A-Z]{2}\d{2}[\s-]?(?:[A-Z0-9]{4}[\s-]?){4,7}[A-Z0-9]{1,4}`),
}

// CreditCard matches Visa, Mastercard, Amex, Discover. Luhn-validated downstream.
var CreditCard = Compiled{
	Name:       "CREDIT_CARD",
	Source:     "regex/credit-card",
	Confidence: 0.95,
	Regexp:     regexp.MustCompile(`\b(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|3[47][0-9]{13}|6(?:011|5[0-9]{2})[0-9]{12})\b`),
}
