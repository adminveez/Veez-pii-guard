<pre align="center">
__     _______ _____ _____
\ \   / / ____| ____|__  /
 \ \ / /|  _| |  _|   / /
  \ V / | |___| |___ / /_
   \_/  |_____|_____/____|
</pre>

<h1 align="center">VEEZ</h1>

<h3 align="center">veez-pii-guard</h3>

<p align="center">Offline PII detection and anonymization for LLM-bound text. Pure Go, plugin-based, stream-capable.</p>

<p align="center">
  <a href="https://github.com/adminveez/Veez-pii-guard/actions/workflows/ci.yml"><img src="https://github.com/adminveez/Veez-pii-guard/actions/workflows/ci.yml/badge.svg?branch=main" alt="Build status"></a>
  <a href="https://github.com/adminveez/Veez-pii-guard"><img src="https://img.shields.io/badge/coverage-88.8%25-brightgreen" alt="Coverage"></a>
  <a href="https://go.dev/doc/devel/release#go1.22.0"><img src="https://img.shields.io/badge/go-1.22-00ADD8?logo=go" alt="Go version"></a>
  <a href="https://github.com/adminveez/Veez-pii-guard/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-green.svg" alt="MIT License"></a>
  <a href="https://goreportcard.com/report/github.com/adminveez/Veez-pii-guard"><img src="https://goreportcard.com/badge/github.com/adminveez/Veez-pii-guard" alt="Go Report Card"></a>
</p>

---

## What's new in v0.2

- **PatternPlugin** API — register custom detectors at runtime ([`pii/engine.go`](pii/engine.go))
- **Built-in plugins** — French SSN (NIR), SIRET/SIREN with Luhn, EU driver license
- **Stream scanner** — detect PII in arbitrarily large inputs without loading into memory ([`pii/stream.go`](pii/stream.go))
- **WASM build** — same engine in the browser, ~3.5 MB ([`examples/wasm-demo/`](examples/wasm-demo/))
- **LSP server** — `pii-guard-lsp` highlights PII as diagnostics in any LSP-capable editor ([`cmd/pii-guard-lsp/`](cmd/pii-guard-lsp/))
- **Pre-commit hook** — block commits leaking PII ([`.pre-commit-hooks.yaml`](.pre-commit-hooks.yaml))
- **Optional Rust backend** — opt-in via `-tags veezrust` for high-volume regex passes ([`engine-rust/`](engine-rust/))
- **Benchmark harness** — reproducible comparison against Presidio / spaCy ([`bench/README.md`](bench/README.md))
- **Property tests** — invariants verified with `pgregory.net/rapid` on 1000+ random inputs
- **Breaking change** — `NewDetector(cfg) (*Detector, error)`. Use `MustNewDetector` for the previous panic-on-error behavior.

See [`CHANGELOG.md`](CHANGELOG.md) and [`docs/ARCHITECTURE_DECISIONS.md`](docs/ARCHITECTURE_DECISIONS.md).

---

## Table of Contents

- [Demo](#demo)
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [CLI Usage](#cli-usage)
- [Plugins](#plugins)
- [WASM](#wasm)
- [LSP server](#lsp-server)
- [Architecture](#architecture)
- [Benchmarks](#benchmarks)
- [Why this exists](#why-this-exists)
- [Built as part of something bigger](#built-as-part-of-something-bigger)
- [License](#license)

---

## Demo

```text
$ pii-guard scan --text "Marie Dupont, marie.dupont@cabinet-legal.fr, 06 12 34 56 78"
detections: 2
blocked: false
anonymized:
Marie Dupont, [EMAIL_1], [PHONE_1]
```

---

## Features

- Pure Go core, zero runtime dependencies
- Offline processing — text never leaves the machine
- Deterministic placeholders + reversible mapping
- Runtime-pluggable detectors via `PatternPlugin`
- Stream scanner for arbitrarily large inputs
- WASM build for the browser
- LSP server for editor integration
- Optional Rust acceleration backend (opt-in)

---

## Installation

```bash
go install github.com/veez-ai/veez-pii-guard/cmd/pii-guard@latest
go install github.com/veez-ai/veez-pii-guard/cmd/pii-guard-lsp@latest
```

Docker:

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
    d, err := pii.NewDetector(pii.DefaultConfig())
    if err != nil {
        panic(err)
    }
    res := d.Scan(context.Background(), "contact john@example.com")
    fmt.Println(res.AnonymizedText) // contact [EMAIL_1]
}
```

---

## CLI Usage

```text
pii-guard <command> [flags]

Commands:
  scan        Detect PII and optionally block on policy violations.
  anonymize   Print anonymized text with deterministic placeholders.
  explain     Print every detection with its source pattern, span and confidence.
  stream      Stream-scan stdin and emit JSON per chunk.
  version     Print the build version.
```

Pre-commit shortcut:

```bash
pii-guard scan --block file1.go file2.md   # exits 2 if any PII is found
```

---

## Plugins

```go
type MyPlugin struct{}

func (MyPlugin) Name() string { return "my-plugin" }
func (MyPlugin) Detect(text string) []pii.Match {
    // return []pii.Match{...}
    return nil
}

d, _ := pii.NewDetector(pii.DefaultConfig())
d.Register(MyPlugin{})
```

Built-in plugins live under [`pii/plugins/`](pii/plugins/):

- `frenchssn` — NIR (French social security number) with checksum validation
- `siretsiren` — SIRET / SIREN with Luhn mod-10
- `eudriverlicense` — EU driver license card numbers

---

## WASM

```bash
make wasm
# produces wasm/veez-pii.wasm (~3.5 MB)
```

See [`examples/wasm-demo/index.html`](examples/wasm-demo/index.html) for a vanilla HTML/JS demo.
The exposed JS API: `veezPiiAnonymize(text)`, `veezPiiScan(text)`, `veezPiiVersion()`.

---

## LSP server

`pii-guard-lsp` is a minimal Language Server that publishes PII detections as diagnostics. It works in any LSP-capable editor (VS Code, Neovim, Helix, Emacs).

```bash
go install github.com/veez-ai/veez-pii-guard/cmd/pii-guard-lsp@latest
```

Configure your editor to launch `pii-guard-lsp` for the file types you want monitored.

---

## Architecture

```text
veez-pii-guard/
├── pii/                 ← core engine, plugins, stream, names
│   ├── patterns/        ← regex packs split by family
│   ├── plugins/         ← built-in plugins (frenchssn, siretsiren, eudriverlicense)
│   └── context/         ← embedded firstname dictionary
├── cmd/
│   ├── pii-guard/       ← CLI
│   └── pii-guard-lsp/   ← LSP server
├── wasm/                ← browser build
├── engine-rust/         ← optional Rust acceleration crate
├── bench/               ← reproducible benchmark harness
├── docs/                ← ADRs
└── .github/             ← CI matrix (Go 1.22+1.23 × 3 OS, lint, fuzz, wasm, rust)
```

See [`docs/ARCHITECTURE_DECISIONS.md`](docs/ARCHITECTURE_DECISIONS.md) for the 8 ADRs covering plugin design, FFI, WASM, and benchmark methodology.

---

## Benchmarks

Local synthetic harness (200 samples, AMD EPYC 7763, pure Go):

| Metric | Value |
| --- | ---: |
| Throughput | `5.3 M chars/sec` |
| p50 latency | `0.018 ms` |
| p95 latency | `0.038 ms` |
| p99 latency | `0.045 ms` |

Reproduce:

```bash
go run ./bench/cmd/run --engine=veez --samples=1000 --out=bench/results/veez.json
```

Comparison methodology against Presidio and spaCy is documented in [`bench/README.md`](bench/README.md).

---

## Why this exists

On August 2, 2026, the EU AI Act starts applying to high-risk AI systems.
Every unfiltered LLM request that contains personal data is a potential GDPR violation.
This module is a first practical protection layer you can deploy immediately in front of LLM calls.
It runs offline, with zero cloud dependency, so personal data does not have to leave the machine.

---

## Built as part of something bigger

`veez-pii-guard` is the first public building block of **VEEZ**.

VEEZ is a sovereign AI infrastructure for European companies that cannot afford to send their data through American clouds — and increasingly, can't legally do so either.

It's being built by one developer. Quietly. From scratch.

The rest is coming. No roadmap, no countdown — just work.

→ If this resonates with you — engineer, operator, or simply curious — reach out.

---

## License

[MIT License](LICENSE)
