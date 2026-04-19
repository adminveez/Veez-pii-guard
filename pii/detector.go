package pii

import (
	"context"
	"regexp"
	"strings"
)

// Detector performs regex-based PII detection.
type Detector struct {
	cfg      Config
	patterns map[Type]*regexp.Regexp
}

// NewDetector creates a detector with precompiled patterns.
func NewDetector(cfg Config) *Detector {
	d := &Detector{
		cfg:      cfg,
		patterns: make(map[Type]*regexp.Regexp),
	}
	d.compilePatterns()
	return d
}

func (d *Detector) compilePatterns() {
	d.patterns[TypeEmail] = regexp.MustCompile(`(?i)[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}`)
	d.patterns[TypePhone] = regexp.MustCompile(`(?:\+33[\s.-]?|0033[\s.-]?|0)[1-9][\s.-]?(?:[0-9][\s.-]?){7,12}[0-9]`)
	d.patterns[TypePhoneE164] = regexp.MustCompile(`\+[\s.-]?[0-9]{1,4}[\s.-]?[0-9]{1,4}[\s.-]?[0-9]{1,4}[\s.-]?[0-9]{2,4}`)
	d.patterns[TypeIBAN] = regexp.MustCompile(`(?i)[A-Z]{2}\d{2}[\s-]?(?:[A-Z0-9]{4}[\s-]?){4,7}[A-Z0-9]{1,4}`)
	d.patterns[TypeCreditCard] = regexp.MustCompile(`\b(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|3[47][0-9]{13}|6(?:011|5[0-9]{2})[0-9]{12})\b`)
	d.patterns[TypeFrenchSSN] = regexp.MustCompile(`\b[12][0-9]{2}(?:0[1-9]|1[0-2])(?:(?:0[1-9]|[1-8][0-9]|9[0-5])|2[AB])[0-9]{3}[0-9]{3}(?:\s?[0-9]{2})?\b`)
	d.patterns[TypeIPAddress] = regexp.MustCompile(`\b(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`)
	d.patterns[TypeAPIKey] = regexp.MustCompile(`(?i)(?:api[_-]?key|apikey|api_secret|api-secret)[\s:="']+([a-zA-Z0-9_-]{20,})`)
	d.patterns[TypeBearerToken] = regexp.MustCompile(`(?i)bearer\s+[a-zA-Z0-9_.-]{20,}`)
	d.patterns[TypeSecret] = regexp.MustCompile(`(?i)(?:password|secret|token|passwd|pwd)[\s:="']+[^\s"']{8,}`)
	d.patterns[TypeContractRef] = regexp.MustCompile(`(?:Contrat\s*n[°ºo]?\s*\d{4}[-_]?\d*|Contract\s*#?\s*\d+)`)
	d.patterns[TypeCaseRef] = regexp.MustCompile(`(?:Dossier\s*n?[°]?\s*[\w-]+|Affaire\s*\d+/\d{4}|Case\s*#?\s*[\w-]+|Réf\.\s*[\w.-]+)`)
	d.patterns[TypeClientID] = regexp.MustCompile(`(?:Client\s*ID|N°\s*client|Numéro\s*client|Customer\s*ID|Client\s*#)[\s:=-]*[\w.-]+`)
}

// Scan analyzes text and returns detections and optional anonymized output.
func (d *Detector) Scan(ctx context.Context, text string) Result {
	_ = ctx
	result := Result{
		Detections:     []Detection{},
		AnonymizedText: text,
	}

	if d.cfg.DetectEmails {
		d.detect(text, TypeEmail, 0.95, &result)
	}
	if d.cfg.DetectPhones {
		d.detect(text, TypePhone, 0.95, &result)
	}
	if d.cfg.DetectPhonesInternational {
		d.detect(text, TypePhoneE164, 0.90, &result)
	}
	if d.cfg.DetectIBANs {
		d.detect(text, TypeIBAN, 0.98, &result)
	}
	if d.cfg.DetectCreditCards {
		d.detect(text, TypeCreditCard, 0.95, &result)
	}
	if d.cfg.DetectFrenchSSN {
		d.detect(text, TypeFrenchSSN, 0.98, &result)
	}
	if d.cfg.DetectIPAddresses {
		d.detect(text, TypeIPAddress, 0.95, &result)
	}
	if d.cfg.DetectSecrets {
		d.detect(text, TypeAPIKey, 0.99, &result)
		d.detect(text, TypeBearerToken, 0.99, &result)
		d.detect(text, TypeSecret, 0.90, &result)
	}
	if d.cfg.DetectContractRef {
		d.detect(text, TypeContractRef, 0.95, &result)
	}
	if d.cfg.DetectCaseRef {
		d.detect(text, TypeCaseRef, 0.95, &result)
	}
	if d.cfg.DetectClientID {
		d.detect(text, TypeClientID, 0.95, &result)
	}

	result.PIICount = len(result.Detections)

	if d.cfg.BlockOnSecrets {
		for _, det := range result.Detections {
			if det.Type == TypeAPIKey || det.Type == TypeBearerToken || det.Type == TypeSecret {
				result.Blocked = true
				result.BlockReason = "Secret detected in payload"
				break
			}
		}
	}

	if d.cfg.BlockOnPII && result.PIICount > 0 {
		result.Blocked = true
		if result.BlockReason == "" {
			result.BlockReason = "PII detected in payload"
		}
	}

	if result.PIICount >= d.cfg.BlockThreshold {
		result.Blocked = true
		if result.BlockReason == "" {
			result.BlockReason = "PII count exceeds threshold"
		}
	}

	if d.cfg.AnonymizeOutput {
		result.AnonymizedText = Anonymize(text, result.Detections)
	}

	return result
}

func (d *Detector) detect(text string, piiType Type, confidence float64, result *Result) {
	if !shouldScanType(text, piiType) {
		return
	}
	pattern, ok := d.patterns[piiType]
	if !ok {
		return
	}

	matches := pattern.FindAllStringIndex(text, -1)
	for _, match := range matches {
		detectedText := text[match[0]:match[1]]
		if piiType == TypeCreditCard && !ValidateLuhn(detectedText) {
			continue
		}
		result.Detections = append(result.Detections, Detection{
			Type:       piiType,
			Text:       detectedText,
			Start:      match[0],
			End:        match[1],
			Confidence: confidence,
			Method:     "regex",
		})
	}
}

func shouldScanType(text string, piiType Type) bool {
	switch piiType {
	case TypeEmail:
		return strings.Contains(text, "@")
	case TypePhone, TypePhoneE164, TypeCreditCard, TypeFrenchSSN:
		return hasDigit(text)
	case TypeIBAN:
		return hasDigit(text) && hasUpper(text)
	case TypeIPAddress:
		return strings.Contains(text, ".") && hasDigit(text)
	case TypeAPIKey:
		lower := strings.ToLower(text)
		return strings.Contains(lower, "api") || strings.Contains(lower, "key")
	case TypeBearerToken:
		return strings.Contains(strings.ToLower(text), "bearer")
	case TypeSecret:
		lower := strings.ToLower(text)
		return strings.Contains(lower, "secret") || strings.Contains(lower, "password") || strings.Contains(lower, "token")
	case TypeContractRef:
		lower := strings.ToLower(text)
		return strings.Contains(lower, "contrat") || strings.Contains(lower, "contract")
	case TypeCaseRef:
		lower := strings.ToLower(text)
		return strings.Contains(lower, "dossier") || strings.Contains(lower, "affaire") || strings.Contains(lower, "case") || strings.Contains(lower, "réf")
	case TypeClientID:
		lower := strings.ToLower(text)
		return strings.Contains(lower, "client") || strings.Contains(lower, "customer")
	default:
		return true
	}
}

func hasDigit(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			return true
		}
	}
	return false
}

func hasUpper(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			return true
		}
	}
	return false
}
