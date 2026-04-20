<div align="center">

<pre>
██╗   ██╗███████╗███████╗███████╗
██║   ██║██╔════╝██╔════╝╚══███╔╝
██║   ██║█████╗  █████╗    ███╔╝ 
╚██╗ ██╔╝██╔══╝  ██╔══╝   ███╔╝  
 ╚████╔╝ ███████╗███████╗███████╗
  ╚═══╝  ╚══════╝╚══════╝╚══════╝
</pre>

# veez-pii-guard

**Offline PII detection and anonymization for LLM-bound text.**  
Pure Go. Plugin-based. Stream-capable. No cloud, no dependency.

[![Build](https://github.com/adminveez/Veez-pii-guard/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/adminveez/Veez-pii-guard/actions/workflows/ci.yml)
[![Coverage](https://img.shields.io/badge/coverage-88.8%25-brightgreen)](#benchmarks)
[![Go](https://img.shields.io/badge/go-1.22+-00ADD8?logo=go)](https://go.dev/doc/devel/release#go1.22.0)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/adminveez/Veez-pii-guard)](https://goreportcard.com/report/github.com/adminveez/Veez-pii-guard)

</div>

---

## Overview

`veez-pii-guard` scans any text for personally identifiable information and anonymizes it before it reaches an LLM, a log sink, or any external API. It runs entirely offline — no data ever leaves the machine.

It is designed for production: plugin-extensible, stream-native, WASM-compatible, and equipped with a Language Server for in-editor detection.

---

## What's new in v0.2

| Area | Change |
| --- | --- |
| **API** | `NewDetector` now returns `(*Detector, error)`. Use `MustNewDetector` for the old behavior. |
| **Plugins** | `PatternPlugin` interface — register custom detectors at runtime |
| **Built-in plugins** | French SSN (NIR), SIRET/SIREN with Luhn, EU driver license |
| **Stream** | `StreamScanner` handles arbitrarily large inputs without loading into memory |
| **WASM** | Same engine compiled for the browser (~3.5 MB) |
| **LSP** | `pii-guard-lsp` publishes PII diagnostics in any LSP-capable editor |
| **Pre-commit** | Block commits leaking PII via `.pre-commit-hooks.yaml` |
| **Rust backend** | Optional Rust acceleration, opt-in via `-tags veezrust` |
| **Benchmark** | Reproducible harness with precision / recall / F1 and latency metrics |
| **Property tests** | Structural invariants verified with `pgregory.net/rapid` on 1000+ random inputs |

See [`CHANGELOG.md`](CHANGELOG.md) and [`docs/ARCHITECTURE_DECISIONS.md`](docs/ARCHITECTURE_DECISIONS.md) for full details.

---

## Table of Contents

- [Demo](#demo)
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [CLI Usage](#cli-usage)
- [Plugins](#plugins)
- [Stream Scanner](#stream-scanner)
- [WASM](#wasm)
- [LSP Server](#lsp-server)
- [Pre-commit Hook](#pre-commit-hook)
- [Architecture](#architecture)
- [Benchmarks](#benchmarks)
- [Why this exists](#why-this-exists)
- [Contributing](#contributing)
- [Built as part of something bigger](#built-as-part-of-something-bigger)
- [License](#license)

---

## Demo

**Before** — what your app would send to OpenAI / Anthropic / Mistral:

```text
Marie Dupont, marie.dupont@cabinet-legal.fr, +33 6 12 34 56 78,
IBAN FR76 3000 4000 0312 3456 7890 143, SIRET 552 100 554 00013
```

**After** — what `veez-pii-guard` lets through:

```text
[NAME_1], [EMAIL_1], [PHONE_1],
IBAN [IBAN_1], SIRET [SIRET_1]
```

**One command:**

```text
$ pii-guard scan --text "Marie Dupont, marie.dupont@cabinet-legal.fr, +33 6 12 34 56 78"

detections: 3
blocked:    false
anonymized:
  Marie Dupont, [EMAIL_1], [PHONE_1]
```

**Detail mode** — every span, every confidence:

```text
$ pii-guard explain --text "Authorization: Bearer eyJhbGc..."

Found 1 detection:

  [1] BEARER_TOKEN
      span:       [16, 47]
      text:       eyJhbGc...
      confidence: 0.95
      source:     regex:bearer_token
```

**Reversible** — keep the mapping, reidentify after the LLM responds:

```go
res := d.Scan(ctx, input)              // [EMAIL_1], [PHONE_1] ...
llmReply := callOpenAI(res.AnonymizedText)
final := pii.Reidentify(llmReply, res.Mappings)  // back to real values
```

---

## Features

- Zero runtime dependencies — pure Go standard library
- Offline processing — data never leaves the process
- Deterministic placeholders — `[EMAIL_1]`, `[PHONE_2]`, reversible via `Reidentify`
- Runtime plugin system — register custom detectors at startup or at test time
- Stream scanner — detect PII in multi-gigabyte inputs chunk by chunk
- WASM build — same detection logic in the browser
- LSP server — PII highlights as editor diagnostics
- Optional Rust acceleration — high-volume regex, opt-in, no forced dependency
- Fuzz-tested — 4 fuzz targets covering scan, anonymize, reidentify, stream
- Property-tested — idempotence, bounds, mapping consistency on random inputs

---

## Installation

### CLI

```bash
go install github.com/adminveez/Veez-pii-guard/cmd/pii-guard@latest
go install github.com/adminveez/Veez-pii-guard/cmd/pii-guard-lsp@latest
```

### As a library

```bash
go get github.com/adminveez/Veez-pii-guard
```

### Docker

```bash
docker build -t veez-pii-guard .
docker run --rm veez-pii-guard scan --help
```

---

## Quick Start

```go
package main

import (
    "context"
    "fmt"

    "github.com/adminveez/Veez-pii-guard/pii"
)

func main() {
    d, err := pii.NewDetector(pii.DefaultConfig())
    if err != nil {
        panic(err)
    }

    res := d.Scan(context.Background(), "Call me at +33 6 12 34 56 78 or john@example.com")
    fmt.Println(res.AnonymizedText)
    // Call me at [PHONE_1] or [EMAIL_1]

    // Reverse the anonymization
    original := pii.Reidentify(res.AnonymizedText, res.Mappings)
    fmt.Println(original)
}
```

---

## CLI Usage

```
pii-guard <command> [flags]

Commands:
  scan        Detect PII. Optionally block (exit 2) on policy violations.
  anonymize   Print anonymized text with deterministic placeholders.
  explain     Print every detection: type, span, confidence, source pattern.
  stream      Stream-scan stdin, emit one JSON object per chunk.
  version     Print the build version.
```

**Common flags:**

| Flag | Default | Description |
| --- | --- | --- |
| `--text` | — | Inline text to scan |
| `--file` | — | Path to input file |
| `--format` | `text` | Output format: `text` or `json` |
| `--block` | `false` | Block (exit 2) on any PII — shortcut for pre-commit |
| `--block-on-secrets` | `true` | Block when API keys or tokens are found |
| `--detect-names` | `false` | Enable contextual given-name detection (opt-in) |

**Examples:**

```bash
# Block a commit if any file contains PII
pii-guard scan --block contracts/nda.md

# Output JSON for downstream processing
pii-guard scan --file logs.jsonl --format json

# Stream large files without loading into memory
cat access.log | pii-guard stream --format json
```

---

## Plugins

Plugins implement the `PatternPlugin` interface and are registered at runtime. They participate in the same overlap-resolution pipeline as built-in detectors.

```go
type MyPlugin struct{}

func (MyPlugin) Name() string { return "my-plugin" }

func (MyPlugin) Detect(text string) []pii.Match {
    // return spans with type, text, start, end, confidence
    return nil
}

d, _ := pii.NewDetector(pii.DefaultConfig())
d.Register(MyPlugin{})
```

**Built-in plugins** (under [`pii/plugins/`](pii/plugins/)):

| Plugin | Type emitted | Notes |
| --- | --- | --- |
| `frenchssn` | `FRENCH_SSN` | NIR with mod-97 checksum validation |
| `siretsiren` | `SIRET`, `SIREN` | Luhn mod-10 validation |
| `eudriverlicense` | `EU_DRIVER_LICENSE` | Multi-country card format |

---

## Stream Scanner

For large files, use `StreamScanner` to avoid loading the entire input into memory. It uses a tail-overlap algorithm to detect PII that spans chunk boundaries.

```go
scanner, _ := pii.NewStreamScanner(pii.StreamOptions{
    ChunkSize: 4096,
    Overlap:   256,
})

chunks, errCh := scanner.Scan(context.Background(), file)
for chunk := range chunks {
    fmt.Printf("offset %d: %d detections\n", chunk.Offset, len(chunk.Detections))
}
if err := <-errCh; err != nil {
    log.Fatal(err)
}
```

---

## WASM

Build the engine for the browser:

```bash
make wasm
# Output: wasm/veez-pii.wasm (~3.5 MB)
```

The exposed JavaScript API:

```js
veezPiiAnonymize(text)   // returns anonymized string
veezPiiScan(text)        // returns JSON detections
veezPiiVersion()         // returns version string
```

See [`examples/wasm-demo/index.html`](examples/wasm-demo/index.html) for a working vanilla HTML demo.

---

## LSP Server

`pii-guard-lsp` implements a minimal Language Server Protocol server. It publishes PII detections as diagnostics with severity levels (error for secrets, warning for PII).

Compatible with VS Code, Neovim, Helix, Emacs, and any editor that supports LSP.

```bash
go install github.com/adminveez/Veez-pii-guard/cmd/pii-guard-lsp@latest
```

Configure your editor to launch `pii-guard-lsp` as a language server for the file types you want monitored. No configuration file required.

---

## Pre-commit Hook

Add to your `.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/adminveez/Veez-pii-guard
    rev: v0.2.0
    hooks:
      - id: pii-guard
```

The hook scans every staged text file and fails the commit (exit code 2) if any PII is detected.

---

## Architecture

```
veez-pii-guard/
├── pii/
│   ├── patterns/           regex packs split by family (email, phone, financial, ...)
│   ├── plugins/            built-in plugins (frenchssn, siretsiren, eudriverlicense)
│   └── context/            embedded firstname dictionary (~250 names)
├── cmd/
│   ├── pii-guard/          CLI binary
│   └── pii-guard-lsp/      LSP server binary
├── wasm/                   browser build entry point
├── engine-rust/            optional Rust acceleration crate (cdylib)
├── bench/                  reproducible benchmark harness + dataset
├── docs/                   8 Architecture Decision Records
└── .github/workflows/      CI: Go 1.22+1.23 x 3 OS, lint, fuzz, wasm, rust
```

**Architecture Decision Records** in [`docs/ARCHITECTURE_DECISIONS.md`](docs/ARCHITECTURE_DECISIONS.md):

| ADR | Decision |
| --- | --- |
| 001 | Pure Go first, Rust opt-in |
| 002 | Decoupled pattern engine |
| 003 | Runtime plugin interface |
| 004 | No neural NER — regex + context only |
| 005 | Tail-overlap stream algorithm |
| 006 | JSON wire format for FFI |
| 007 | Reproducible benchmark methodology |
| 008 | WASM via Go standard library |

---

## Benchmarks

Measured on a synthetic corpus of 200 samples (AMD EPYC 7763, pure Go, no Rust):

| Metric | Value |
| --- | ---: |
| Throughput | 5.3 M chars / sec |
| p50 latency | 0.018 ms |
| p95 latency | 0.038 ms |
| p99 latency | 0.045 ms |

Reproduce:

```bash
go run ./bench/cmd/run --engine=veez --samples=1000 --out=bench/results/veez.json
```

The harness outputs precision / recall / F1 per PII type and can run against external engines (Presidio, spaCy). See [`bench/README.md`](bench/README.md) for the full methodology.

---

## Why this exists

### The legal pressure is real, and it lands in 2026

| Date | Regulation | What it means for LLM apps |
| --- | --- | --- |
| **Already in force** | **GDPR (since 2018)** | Sending personal data to a US-hosted LLM without a legal basis = up to **4% of worldwide revenue** in fines. |
| **Feb 2, 2025** | **EU AI Act — prohibitions** | Some uses (social scoring, manipulation) are already banned. |
| **Aug 2, 2025** | **EU AI Act — GPAI obligations** | Foundation models (GPT, Claude, Mistral, Llama) must publish training data summaries and respect EU copyright. |
| **Aug 2, 2026** | **EU AI Act — high-risk systems** | Full obligations apply. Any AI handling HR, healthcare, credit, justice, education must log inputs and prove data minimisation. |
| **Aug 2, 2027** | **EU AI Act — full enforcement** | All provisions in force. Fines up to **€35M or 7% of revenue**. |

### What unfiltered LLM calls actually leak

Names. Emails. Phone numbers. National IDs (NIR, SSN). Bank accounts (IBAN). Company
identifiers (SIRET, SIREN, VAT). Driver licences. API keys and bearer tokens.
Sometimes entire contracts pasted in a chat box "to summarise".

Most of it ends up sitting on US infrastructure, often used for retraining,
subject to the **CLOUD Act** — which means a US authority can subpoena it without
notifying you or your users.

### What this module does about it

`veez-pii-guard` sits **between your code and the LLM provider**. Every prompt
passes through it first. PII is replaced by deterministic placeholders
(`[EMAIL_1]`, `[NAME_1]`, ...). The LLM sees the structure, not the data.
The original mapping stays in your process, so you can reidentify the response
locally.

It runs offline. Zero network. Zero dependency. Pure Go standard library.
You can audit every line.

This is not a complete compliance solution. It is the **first line of defence**
any serious EU AI deployment will need — and the cheapest one to add today.

---

## Contributing

Contributions should be focused, tested, and easy to review.

- Found a bug? Open an issue with a minimal input sample, the expected result, and the actual output.
- Want to contribute? Start with a scoped change, explain the impact, and include the commands you used to validate locally.
- New pattern or plugin? Add tests that cover both positive matches and near-miss rejections.

See [`CONTRIBUTING.md`](CONTRIBUTING.md) if it exists, or follow standard Go project conventions.

---

## Built as part of something bigger

<div align="center">

<pre>
██╗   ██╗███████╗███████╗███████╗
██║   ██║██╔════╝██╔════╝╚══███╔╝
██║   ██║█████╗  █████╗    ███╔╝ 
╚██╗ ██╔╝██╔══╝  ██╔══╝   ███╔╝  
 ╚████╔╝ ███████╗███████╗███████╗
  ╚═══╝  ╚══════╝╚══════╝╚══════╝
</pre>

**Sovereign AI infrastructure for Europe.**

</div>

`veez-pii-guard` is the **first public building block of VEEZ** — a sovereign AI
platform built for European companies that cannot afford to send their data
through American clouds, and increasingly cannot legally do so.

### What VEEZ is building

- **Local-first AI runtime** — your data, your machines, your control
- **EU-hosted gateway** — multi-tenant SaaS for teams, on EU sovereign cloud
- **Compliance by design** — GDPR, EU AI Act, NIS2, DORA aligned from day one
- **Open building blocks** — like this one, MIT-licensed, so anyone can audit and reuse

### Who is building it

One founder. Solo. From scratch. Shipping in public.

No VC. No marketing team. No countdown. Just code that works, released as it
stabilises. `veez-pii-guard` is the first piece you can use today.

### Where to find the rest

- GitHub: [github.com/adminveez](https://github.com/adminveez)
- Topics: `#sovereign-ai` `#eu-ai-act` `#gdpr` `#made-in-france` `#privacy-first`

If this resonates with you — engineer, operator, investor, or simply curious —
open an issue, star the repo, or reach out.

---

## License

[MIT License](LICENSE)
