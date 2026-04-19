# veez-pii-guard

<pre align="center">
__     _______ _____ _____
\ \   / / ____| ____|__  /
 \ \ / /|  _| |  _|   / /
  \ V / | |___| |___ / /_
   \_/  |_____|_____/____|
</pre>

<h1 align="center">veez-pii-guard</h1>

<p align="center">Offline PII detection and anonymization for LLM-bound text in pure Go.</p>

<p align="center">
  <a href="https://github.com/adminveez/Veez-pii-guard/actions/workflows/ci.yml"><img src="https://github.com/adminveez/Veez-pii-guard/actions/workflows/ci.yml/badge.svg?branch=main" alt="Build status"></a>
  <a href="https://github.com/adminveez/Veez-pii-guard#benchmarks"><img src="https://img.shields.io/badge/coverage-80.4%25-brightgreen" alt="Coverage"></a>
  <a href="https://go.dev/doc/devel/release#go1.22.0"><img src="https://img.shields.io/badge/go-1.22-00ADD8?logo=go" alt="Go version"></a>
  <a href="https://github.com/adminveez/Veez-pii-guard/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-green.svg" alt="MIT License"></a>
  <a href="https://goreportcard.com/report/github.com/adminveez/Veez-pii-guard"><img src="https://goreportcard.com/badge/github.com/adminveez/Veez-pii-guard" alt="Go Report Card"></a>
  <a href="https://github.com/adminveez/Veez-pii-guard/stargazers"><img src="https://img.shields.io/github/stars/adminveez/Veez-pii-guard?style=social" alt="GitHub stars"></a>
</p>

---

## Table of Contents

- [Demo](#demo)
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [CLI Usage](#cli-usage)
- [Architecture](#architecture)
- [Benchmarks](#benchmarks)
- [Roadmap](#roadmap)
- [Contributing](#contributing)
- [Code of Conduct](#code-of-conduct)
- [Why this exists](#why-this-exists)
- [License](#license)
- [About VEEZ](#about-veez)

---

## Demo

```text
$ pii-guard --text "Bonjour, je suis Marie Dupont, mon email est marie.dupont@cabinet-legal.fr et mon numéro est le 06 12 34 56 78"
detections: 2
blocked: false
anonymized:
Bonjour, je suis Marie Dupont, mon email est [EMAIL_1] et mon numéro est le [PHONE_1]
```

---

## Features

- `🧱` Pure Go core with zero runtime dependencies
- `🔒` Offline processing so text stays on the machine
- `📍` Deterministic placeholders plus reversible mapping support
- `🧪` Tested thresholds for latency, throughput, and concurrency
- `🖥️` CLI for direct text, files, or piped stdin
- `📦` Small surface area that is easy to embed in other Go services

---

## Installation

### Go install

```bash
go install github.com/veez-ai/veez-pii-guard/cmd/pii-guard@latest
```

### Docker

```bash
docker build -t veez-pii-guard .
docker run --rm veez-pii-guard --help
```

---

## Quick Start

```go
package main
import (
    "context"
    "fmt"
    "github.com/veez-ai/veez-pii-guard/pii"
)
func main() {
    r := pii.NewDetector(pii.DefaultConfig()).Scan(context.Background(), "john@example.com")
    fmt.Println(r.AnonymizedText)
}
```

---

## CLI Usage

| Flag | Scope | Description | Example |
| --- | --- | --- | --- |
| `--text` | `scan`, `anonymize` | Scan raw inline text | `pii-guard --text "john@example.com"` |
| `--file` | `scan`, `anonymize` | Read input from a file | `pii-guard --file contract.txt` |
| `--format` | `scan`, `anonymize` | Output format: `text` or `json` | `pii-guard --text "john@example.com" --format json` |
| `--block-on-secrets` | `scan` | Fail when secrets are detected | `pii-guard scan --text "password=secret" --block-on-secrets=true` |
| `--block-on-pii` | `scan` | Fail when any PII is detected | `pii-guard scan --text "john@example.com" --block-on-pii=true` |
| `--block-threshold` | `scan` | Block when detection count reaches the threshold | `pii-guard scan --file batch.txt --block-threshold 10` |

The CLI also accepts piped stdin when no `--text` or `--file` flag is provided.

---

## Architecture

```text
veez-pii-guard/
├── pii/          ← core detection & anonymization engine
├── cmd/          ← CLI entrypoint
├── examples/     ← ready-to-run usage examples
└── .github/      ← CI workflows
```

---

## Benchmarks

| Scenario | Result | Threshold |
| --- | ---: | ---: |
| Short text, 100 words | `19315 ns/op` (`0.019 ms`) | `< 1 ms` |
| Medium text, 1000 words | `159304 ns/op` (`0.159 ms`) | `< 5 ms` |
| Long text, 10000 words | `2091130 ns/op` (`2.091 ms`) | `< 50 ms` |
| 1000 texts in parallel | `12659217 ns/op` (`12.659 ms`) | `< 2 s` |

Environment: `AMD EPYC 7763`, `go test -bench`.

---

## Roadmap

- [ ] Détection des noms propres via NLP léger
- [ ] Support multilingue étendu
- [ ] Mode stream pour les textes longs
- [ ] Plugin Langchain
- [ ] Dashboard de métriques

---

## Contributing

Contributions should stay focused, tested, and easy to review. If you change behavior, update the documentation and include a reproducible validation path.

If you find a bug, open an issue with a minimal input sample, the expected behavior, and the actual output. Reproducible reports are the fastest to fix.

If you want to propose a pull request, start with a scoped change, explain the impact clearly, and include the commands you used to validate it locally.

---

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md).

---

## Why this exists

On August 2, 2026, the EU AI Act starts applying to high-risk AI systems.
Every unfiltered LLM request that contains personal data is a potential GDPR violation.
This module is a first practical protection layer you can deploy immediately in front of LLM calls.
It runs offline, with zero cloud dependency, so personal data does not have to leave the machine.

---

## License

[MIT License](LICENSE)

---

## About VEEZ

veez-pii-guard is the first open building block of VEEZ — a sovereign AI infrastructure built for European businesses. More coming soon.
