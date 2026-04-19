package pii

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
)

func TestScanDetectEmail(t *testing.T) {
	d := NewDetector(DefaultConfig())
	res := d.Scan(context.Background(), "Contact john@example.com")
	if res.PIICount == 0 {
		t.Fatal("expected at least one detection")
	}
	found := false
	for _, det := range res.Detections {
		if det.Type == TypeEmail {
			found = true
		}
	}
	if !found {
		t.Fatal("expected email detection")
	}
}

func TestOnePIIEachType(t *testing.T) {
	d := NewDetector(DefaultConfig())
	text := strings.Join([]string{
		"john@example.com",
		"06 12 34 56 78",
		"+33 6 11 22 33 44",
		"FR76 3000 6000 0112 3456 7890 189",
		"4111111111111111",
		"285012A12345678",
		"192.168.1.12",
		"api_key: sk-abcdefghijklmnopqrstuvwxyz",
		"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
		"password = superSecret1234",
		"Contrat n° 2024-001",
		"Dossier n° AFF-123",
		"Client ID: ACME-001",
	}, " | ")

	res := d.Scan(context.Background(), text)
	if res.PIICount < 12 {
		t.Fatalf("expected broad type coverage, got %d detections", res.PIICount)
	}
}

func TestMultipleSameType(t *testing.T) {
	d := NewDetector(DefaultConfig())
	text := "a@b.com, c@d.com, e@f.com"
	res := d.Scan(context.Background(), text)
	count := 0
	for _, det := range res.Detections {
		if det.Type == TypeEmail {
			count++
		}
	}
	if count != 3 {
		t.Fatalf("expected 3 emails, got %d", count)
	}
}

func TestMultipleMixedTypes(t *testing.T) {
	d := NewDetector(DefaultConfig())
	text := "Email: jane@corp.fr Phone: 06 12 34 56 78 IBAN: FR76 3000 6000 0112 3456 7890 189"
	res := d.Scan(context.Background(), text)
	if res.PIICount < 3 {
		t.Fatalf("expected at least 3 mixed detections, got %d", res.PIICount)
	}
}

func TestFrenchAndEnglishText(t *testing.T) {
	d := NewDetector(DefaultConfig())
	fr := d.Scan(context.Background(), "Bonjour, mon email est marie.dupont@cabinet.fr et mon telephone est 06 12 34 56 78")
	en := d.Scan(context.Background(), "Hello, contact me at john.doe@firm.com and call +33 6 10 20 30 40")
	if fr.PIICount == 0 || en.PIICount == 0 {
		t.Fatalf("expected detections in FR and EN, got FR=%d EN=%d", fr.PIICount, en.PIICount)
	}
}

func TestEmptyText(t *testing.T) {
	d := NewDetector(DefaultConfig())
	res := d.Scan(context.Background(), "")
	if res.PIICount != 0 {
		t.Fatalf("expected 0 detection, got %d", res.PIICount)
	}
	if res.AnonymizedText != "" {
		t.Fatalf("expected empty anonymized text")
	}
}

func TestTextWithoutPIIReturnsIdentical(t *testing.T) {
	d := NewDetector(DefaultConfig())
	input := "Le ciel est bleu aujourd'hui"
	res := d.Scan(context.Background(), input)
	if res.PIICount != 0 {
		t.Fatalf("expected 0 detection, got %d", res.PIICount)
	}
	if res.AnonymizedText != input {
		t.Fatalf("expected same text, got %q", res.AnonymizedText)
	}
}

func TestVeryLongText10000Words(t *testing.T) {
	d := NewDetector(DefaultConfig())
	var b strings.Builder
	for i := 0; i < 10000; i++ {
		b.WriteString("mot ")
	}
	b.WriteString(" contact: test.long@example.com")
	res := d.Scan(context.Background(), b.String())
	if res.PIICount == 0 {
		t.Fatal("expected at least one detection in long text")
	}
}

func TestPIIAgainstPunctuation(t *testing.T) {
	d := NewDetector(DefaultConfig())
	text := "(marie@email.fr), fin."
	res := d.Scan(context.Background(), text)
	if res.PIICount == 0 {
		t.Fatal("expected detection near punctuation")
	}
}

func TestCaseSensitivity(t *testing.T) {
	d := NewDetector(DefaultConfig())
	text := "A@B.COM a@b.com Mixed.Case@Example.FR"
	res := d.Scan(context.Background(), text)
	if res.PIICount < 3 {
		t.Fatalf("expected 3 email detections, got %d", res.PIICount)
	}
}

func TestConcurrentCalls100Goroutines(t *testing.T) {
	d := NewDetector(DefaultConfig())
	ctx := context.Background()
	input := "Email: john@example.com Tel: 06 12 34 56 78"

	var wg sync.WaitGroup
	errCh := make(chan error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res := d.Scan(ctx, input)
			if res.PIICount < 2 {
				errCh <- fmt.Errorf("expected >=2 detections, got %d", res.PIICount)
			}
		}()
	}

	wg.Wait()
	close(errCh)
	for err := range errCh {
		t.Fatal(err)
	}
}

func TestScanAnonymize(t *testing.T) {
	d := NewDetector(DefaultConfig())
	text := "Contact john@example.com for details"
	res := d.Scan(context.Background(), text)
	if strings.Contains(res.AnonymizedText, "john@example.com") {
		t.Fatal("email should be anonymized")
	}
	if !strings.Contains(res.AnonymizedText, "[EMAIL_1]") {
		t.Fatalf("expected [EMAIL_1], got %q", res.AnonymizedText)
	}
}

func TestScanBlockOnSecret(t *testing.T) {
	cfg := DefaultConfig()
	cfg.BlockOnSecrets = true
	d := NewDetector(cfg)
	res := d.Scan(context.Background(), `api_key: sk-abcdefghijklmnopqrstuvwxyz`)
	if !res.Blocked {
		t.Fatal("expected blocked result")
	}
	if res.BlockReason == "" {
		t.Fatal("expected block reason")
	}
}

func TestSemanticPlaceholders(t *testing.T) {
	d := NewDetector(DefaultConfig())
	text := "Contrat n° 2024-001. Dossier n° AFF-123. Client ID: ACME-001"
	res := d.Scan(context.Background(), text)
	if !strings.Contains(res.AnonymizedText, "[CONTRACT_REF]") {
		t.Fatalf("expected [CONTRACT_REF], got %q", res.AnonymizedText)
	}
	if !strings.Contains(res.AnonymizedText, "[CASE_REF_1]") {
		t.Fatalf("expected [CASE_REF_1], got %q", res.AnonymizedText)
	}
	if !strings.Contains(res.AnonymizedText, "[CLIENT_ID]") {
		t.Fatalf("expected [CLIENT_ID], got %q", res.AnonymizedText)
	}
}

func TestMapAndReidentifyIntegrity(t *testing.T) {
	d := NewDetector(DefaultConfig())
	input := "Bonjour, je suis Marie Dupont, mon email est marie.dupont@cabinet-legal.fr et mon numero est 06 12 34 56 78"
	res := d.Scan(context.Background(), input)
	anonymized, mappings := AnonymizeWithMap(input, res.Detections)

	if len(mappings) == 0 {
		t.Fatal("expected non-empty mappings")
	}
	for placeholder, original := range mappings {
		if !strings.Contains(anonymized, placeholder) {
			t.Fatalf("expected anonymized text to contain placeholder %s", placeholder)
		}
		if original == "" {
			t.Fatalf("expected original value for placeholder %s", placeholder)
		}
	}

	restored := Reidentify(anonymized, mappings)
	if restored != input {
		t.Fatalf("expected restored text to equal original\nwant: %q\ngot:  %q", input, restored)
	}
}
