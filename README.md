# veez-pii-guard

Anonymize PII before sending data to LLMs.

## Features

- Regex-based detection for common sensitive data types
- Semantic placeholders (example: `[EMAIL_1]`, `[PHONE_1]`)
- Optional blocking strategy for secrets and PII thresholds
- Utility masking and validation helpers (Luhn, API-key heuristics)
- Zero dependency runtime

## Demo

```text
$ pii-guard --text "Bonjour, je suis Marie Dupont, mon email est marie.dupont@cabinet-legal.fr et mon numéro est le 06 12 34 56 78."

✓ 3 PII detected in 0.019ms

Anonymized:
"Bonjour, je suis [NAME_1], mon email est [EMAIL_1] et mon numéro est le [PHONE_1]."

Map:
NAME_1   → Marie Dupont
EMAIL_1  → marie.dupont@cabinet-legal.fr
PHONE_1  → 06 12 34 56 78
```

## PII types (initial)

- EMAIL
- PHONE, PHONE_E164
- IBAN
- CREDIT_CARD
- FRENCH_SSN
- IP_ADDRESS
- API_KEY, BEARER_TOKEN, SECRET
- CONTRACT_REF, CASE_REF, CLIENT_ID

## Install

```bash
go install github.com/veez-ai/veez-pii-guard/cmd/pii-guard@latest
```

## Library usage

```go
package main

import (
	"context"
	"fmt"

	"github.com/veez-ai/veez-pii-guard/pii"
)

func main() {
cfg := pii.DefaultConfig()
detector := pii.NewDetector(cfg)
res := detector.Scan(context.Background(), "Contact john@example.com")
fmt.Println(res.AnonymizedText) // Contact [EMAIL_1]
}
```

## CLI usage

```bash
pii-guard --text "Contact john@example.com" --format json
echo "Contact: paul.martin@avocat.fr" | pii-guard
pii-guard --file contrat.txt
```

Exit code:

- `0` scan succeeded and not blocked
- `2` blocked by policy
- `1` invalid input or runtime error

## Benchmarks

Measured on `AMD EPYC 7763` with `go test -bench`:

- `BenchmarkScanShort100Words`: `19315 ns/op` (~0.019 ms)
- `BenchmarkScanMedium1000Words`: `159304 ns/op` (~0.159 ms)
- `BenchmarkScanLong10000Words`: `2091130 ns/op` (~2.091 ms)
- `BenchmarkScanParallel1000Texts`: `12659217 ns/op` (~12.659 ms)

All thresholds validated in tests:

- 100 words: `< 1 ms`
- 1000 words: `< 5 ms`
- 10000 words: `< 50 ms`
- 1000 parallel texts: `< 2 s`

## Security note

This project relies on regex/heuristics and can produce false positives/false negatives. For regulated use-cases, combine with human review and domain-specific controls.

## Why this exists

On August 2, 2026, the EU AI Act starts applying to high-risk AI systems.
Every unfiltered LLM request that contains personal data is a potential GDPR violation.
This module is a first practical protection layer you can deploy immediately in front of LLM calls.
It runs offline, with zero cloud dependency, so personal data does not have to leave the machine.

## License

[MIT License](LICENSE)
