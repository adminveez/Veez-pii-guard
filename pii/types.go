// Package pii provides offline PII detection, anonymization and reversible
// re-identification for LLM prompts and arbitrary text payloads.
//
// Public API stability: the types in this file are part of the v1 contract.
// Anything under pii/internal/, pii/context/, pii/plugins/ may change between
// minor versions until v1.0.
package pii

// Type is the canonical identifier for a category of personally identifiable
// information. Values are stable across releases; new types are added, never
// renamed.
type Type string

const (
	TypeEmail       Type = "EMAIL"
	TypePhone       Type = "PHONE"
	TypePhoneE164   Type = "PHONE_E164"
	TypeIBAN        Type = "IBAN"
	TypeCreditCard  Type = "CREDIT_CARD"
	TypeFrenchSSN   Type = "FRENCH_SSN"
	TypeIPAddress   Type = "IP_ADDRESS"
	TypeAPIKey      Type = "API_KEY"
	TypeBearerToken Type = "BEARER_TOKEN"
	TypeSecret      Type = "SECRET"
	TypeContractRef Type = "CONTRACT_REF"
	TypeCaseRef     Type = "CASE_REF"
	TypeClientID    Type = "CLIENT_ID"
	// TypePersonName is emitted by the contextual name detector (off by default).
	TypePersonName Type = "PERSON_NAME"
	// Plugin-emitted types (see pii/plugins/).
	TypeSIRET           Type = "SIRET"
	TypeSIREN           Type = "SIREN"
	TypeEUDriverLicense Type = "EU_DRIVER_LICENSE"
)

// Detection is one PII match in the source text.
type Detection struct {
	Type           Type    `json:"type"`
	Text           string  `json:"text"`
	Start          int     `json:"start"`
	End            int     `json:"end"`
	Confidence     float64 `json:"confidence"`
	Method         string  `json:"method"`
	Source         string  `json:"source,omitempty"`
	RequiresReview bool    `json:"requires_review"`
}

// Match is the value type returned by a PatternPlugin or an internal matcher.
// It carries no anonymization concerns — that is the engine's job.
type Match struct {
	Type       Type
	Start      int
	End        int
	Text       string
	Confidence float64
}

// Result is the output of a scan.
type Result struct {
	Detections     []Detection       `json:"detections"`
	AnonymizedText string            `json:"anonymized_text"`
	Mappings       map[string]string `json:"mappings,omitempty"`
	PIICount       int               `json:"pii_count"`
	Blocked        bool              `json:"blocked"`
	BlockReason    string            `json:"block_reason,omitempty"`
}

// Config defines scanner behavior. Construct via DefaultConfig() and tweak.
type Config struct {
	DetectEmails              bool
	DetectPhones              bool
	DetectPhonesInternational bool
	DetectIBANs               bool
	DetectCreditCards         bool
	DetectFrenchSSN           bool
	DetectIPAddresses         bool
	DetectSecrets             bool
	DetectContractRef         bool
	DetectCaseRef             bool
	DetectClientID            bool

	// DetectNames enables the contextual second-pass name detector
	// (heuristic + embedded FR/EN given-name dictionary). Off by default
	// because of inherent false-positive rate on common nouns. See ADR-004.
	DetectNames bool

	BlockOnPII     bool
	BlockOnSecrets bool
	BlockThreshold int

	AnonymizeOutput bool
}

// DefaultConfig returns sane defaults for general-purpose PII scanning.
// Name detection is disabled (see ADR-004 for the trade-off rationale).
func DefaultConfig() Config {
	return Config{
		DetectEmails:              true,
		DetectPhones:              true,
		DetectPhonesInternational: true,
		DetectIBANs:               true,
		DetectCreditCards:         true,
		DetectFrenchSSN:           true,
		DetectIPAddresses:         true,
		DetectSecrets:             true,
		DetectContractRef:         true,
		DetectCaseRef:             true,
		DetectClientID:            true,
		DetectNames:               false,
		BlockOnPII:                false,
		BlockOnSecrets:            true,
		BlockThreshold:            100,
		AnonymizeOutput:           true,
	}
}

// PatternPlugin is the public extension point for third-party detectors.
// Implementations MUST be safe for concurrent calls and MUST NOT panic;
// panics are recovered by the engine and reported as a skipped detection.
//
// See docs/PLUGINS.md for a 30-line "your first plugin" template.
type PatternPlugin interface {
	Name() string
	Detect(text string) []Match
	Confidence() float64
}
