package pii

// Type represents a type of personally identifiable information.
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
)

// Detection is one PII match in the source text.
type Detection struct {
	Type           Type    `json:"type"`
	Text           string  `json:"text"`
	Start          int     `json:"start"`
	End            int     `json:"end"`
	Confidence     float64 `json:"confidence"`
	Method         string  `json:"method"`
	RequiresReview bool    `json:"requires_review"`
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

// Config defines scanner behavior.
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

	BlockOnPII     bool
	BlockOnSecrets bool
	BlockThreshold int

	AnonymizeOutput bool
}

// DefaultConfig returns sane defaults for general-purpose PII scanning.
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
		BlockOnPII:                false,
		BlockOnSecrets:            true,
		BlockThreshold:            100,
		AnonymizeOutput:           true,
	}
}
