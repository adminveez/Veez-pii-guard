package pii

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/veez-ai/veez-pii-guard/pii/patterns"
)

// Detector is the orchestrator. It owns config, the matcher set, and the set
// of registered plugins. It is safe for concurrent use after construction;
// Register must not be called concurrently with Scan.
type Detector struct {
	cfg     Config
	mu      sync.RWMutex
	plugins []PatternPlugin
}

// ErrInvalidConfig is returned by NewDetector when the supplied config has
// internally inconsistent values.
var ErrInvalidConfig = errors.New("pii: invalid config")

// NewDetector validates the config and returns a ready-to-use Detector.
//
// It does not panic. Use MustNewDetector if you prefer a one-liner with
// panic-on-error semantics (acceptable for tests and main, never for libraries).
func NewDetector(cfg Config) (*Detector, error) {
	if cfg.BlockThreshold < 0 {
		return nil, fmt.Errorf("%w: BlockThreshold must be >= 0", ErrInvalidConfig)
	}
	return &Detector{cfg: cfg}, nil
}

// MustNewDetector is the panic-on-error counterpart of NewDetector.
// Use only in tests or main.
func MustNewDetector(cfg Config) *Detector {
	d, err := NewDetector(cfg)
	if err != nil {
		panic(err)
	}
	return d
}

// Register adds a plugin. Returns an error if the plugin is nil, has an
// empty name, has a confidence outside [0, 1], or duplicates an existing name.
//
// Not safe to call concurrently with Scan.
func (d *Detector) Register(p PatternPlugin) error {
	if p == nil {
		return fmt.Errorf("%w: plugin is nil", ErrInvalidConfig)
	}
	name := strings.TrimSpace(p.Name())
	if name == "" {
		return fmt.Errorf("%w: plugin name is empty", ErrInvalidConfig)
	}
	c := p.Confidence()
	if c < 0 || c > 1 {
		return fmt.Errorf("%w: plugin %q confidence %.3f not in [0,1]", ErrInvalidConfig, name, c)
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	for _, existing := range d.plugins {
		if existing.Name() == name {
			return fmt.Errorf("%w: plugin %q already registered", ErrInvalidConfig, name)
		}
	}
	d.plugins = append(d.plugins, p)
	return nil
}

// Plugins returns a snapshot of the registered plugins. For introspection
// only — modifying the slice has no effect on the detector.
func (d *Detector) Plugins() []PatternPlugin {
	d.mu.RLock()
	defer d.mu.RUnlock()
	out := make([]PatternPlugin, len(d.plugins))
	copy(out, d.plugins)
	return out
}

// Scan analyzes text and returns detections + optional anonymized output.
// The context is honored only for cancellation; per-call deadlines apply
// to long inputs (each plugin runs synchronously).
func (d *Detector) Scan(ctx context.Context, text string) Result {
	result := Result{
		Detections:     []Detection{},
		AnonymizedText: text,
	}
	if text == "" {
		return result
	}
	if err := ctx.Err(); err != nil {
		return result
	}

	// Built-in regex pass.
	d.runBuiltins(text, &result)

	// Plugin pass. Each plugin runs in a recover() so a misbehaving one
	// can never crash the host.
	d.runPlugins(text, &result)

	// Optional contextual name pass. Adds Detection{Type: TypePersonName}.
	if d.cfg.DetectNames {
		d.runNamePass(text, &result)
	}

	result.Detections = resolveOverlaps(result.Detections)
	result.PIICount = len(result.Detections)

	d.applyBlocking(&result)

	if d.cfg.AnonymizeOutput {
		result.AnonymizedText = Anonymize(text, result.Detections)
	}
	return result
}

func (d *Detector) runBuiltins(text string, result *Result) {
	type slot struct {
		enabled  bool
		c        patterns.Compiled
		piiType  Type
		validate func(string) bool
	}
	matchers := []slot{
		{d.cfg.DetectEmails, patterns.Email, TypeEmail, nil},
		{d.cfg.DetectPhones, patterns.PhoneFR, TypePhone, nil},
		{d.cfg.DetectPhonesInternational, patterns.PhoneE164, TypePhoneE164, nil},
		{d.cfg.DetectIBANs, patterns.IBAN, TypeIBAN, nil},
		{d.cfg.DetectCreditCards, patterns.CreditCard, TypeCreditCard, ValidateLuhn},
		{d.cfg.DetectFrenchSSN, patterns.FrenchSSN, TypeFrenchSSN, nil},
		{d.cfg.DetectIPAddresses, patterns.IPAddress, TypeIPAddress, nil},
		{d.cfg.DetectSecrets, patterns.APIKey, TypeAPIKey, nil},
		{d.cfg.DetectSecrets, patterns.BearerToken, TypeBearerToken, nil},
		{d.cfg.DetectSecrets, patterns.GenericSecret, TypeSecret, nil},
		{d.cfg.DetectContractRef, patterns.ContractRef, TypeContractRef, nil},
		{d.cfg.DetectCaseRef, patterns.CaseRef, TypeCaseRef, nil},
		{d.cfg.DetectClientID, patterns.ClientID, TypeClientID, nil},
	}

	for _, m := range matchers {
		if !m.enabled {
			continue
		}
		if !shouldScanType(text, m.piiType) {
			continue
		}
		for _, idx := range m.c.Regexp.FindAllStringIndex(text, -1) {
			detected := text[idx[0]:idx[1]]
			if m.validate != nil && !m.validate(detected) {
				continue
			}
			result.Detections = append(result.Detections, Detection{
				Type:       m.piiType,
				Text:       detected,
				Start:      idx[0],
				End:        idx[1],
				Confidence: m.c.Confidence,
				Method:     "regex",
				Source:     m.c.Source,
			})
		}
	}
}

func (d *Detector) runPlugins(text string, result *Result) {
	d.mu.RLock()
	plugins := make([]PatternPlugin, len(d.plugins))
	copy(plugins, d.plugins)
	d.mu.RUnlock()

	for _, p := range plugins {
		matches := safePluginCall(p, text)
		conf := p.Confidence()
		for _, m := range matches {
			if m.End <= m.Start || m.Start < 0 || m.End > len(text) {
				continue
			}
			result.Detections = append(result.Detections, Detection{
				Type:       m.Type,
				Text:       m.Text,
				Start:      m.Start,
				End:        m.End,
				Confidence: conf,
				Method:     "plugin",
				Source:     p.Name(),
			})
		}
	}
}

// safePluginCall isolates plugin panics. A panicking plugin yields zero
// detections; the host stays alive.
func safePluginCall(p PatternPlugin, text string) (out []Match) {
	defer func() {
		if r := recover(); r != nil {
			out = nil
		}
	}()
	return p.Detect(text)
}

func (d *Detector) applyBlocking(result *Result) {
	if d.cfg.BlockOnSecrets {
		for _, det := range result.Detections {
			if det.Type == TypeAPIKey || det.Type == TypeBearerToken || det.Type == TypeSecret {
				result.Blocked = true
				result.BlockReason = "Secret detected in payload"
				break
			}
		}
	}
	if d.cfg.BlockOnPII && result.PIICount > 0 && !result.Blocked {
		result.Blocked = true
		result.BlockReason = "PII detected in payload"
	}
	if d.cfg.BlockThreshold > 0 && result.PIICount >= d.cfg.BlockThreshold && !result.Blocked {
		result.Blocked = true
		result.BlockReason = "PII count exceeds threshold"
	}
}

// shouldScanType is a cheap pre-filter. Avoids running expensive regexes when
// the input cannot possibly contain a given pattern. Documented in ADR-002.
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
