# Changelog

All notable changes to this project will be documented in this file.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).
This project adheres to [Semantic Versioning](https://semver.org/).

## [0.2.0] — 2026-04-20

### Added
- **Plugin system** — public `PatternPlugin` interface, `Detector.Register(...)` API,
  and three built-in plugins (`frenchssn`, `siretsiren`, `eudriverlicense`) shipped as
  examples under `pii/plugins/`. See [docs/PLUGINS.md](docs/PLUGINS.md).
- **Contextual name detection** (Phase 2 pass) — Unicode tokenizer + sliding-window
  capitalization heuristic + embedded FR/EN given-name dictionary. Off by default
  (`Config.DetectNames`); see [ADR-004](docs/ARCHITECTURE_DECISIONS.md#adr-004).
- **`StreamScanner`** — process arbitrary-size inputs with bounded memory and
  PII detection across chunk boundaries. Property-tested for parity with full-text scan.
- **WASM build** — `make wasm` produces `dist/wasm/veez-pii-guard.wasm` runnable in
  any modern browser. Demo under `examples/wasm-demo/`.
- **Optional Rust engine** — opt-in via `-tags veezrust`. cdylib under `engine-rust/`,
  thin FFI surface (3 functions, JSON over byte buffers). Pure-Go remains the default.
- **`pii-guard explain`** subcommand — for each detection, prints pattern source,
  position, confidence and originating plugin.
- **`pii-guard-lsp`** — minimal Language Server Protocol server that highlights PII
  in any LSP-capable editor (VS Code, Neovim, Helix, Emacs).
- **Pre-commit hook** — `.pre-commit-hooks.yaml` for use with [pre-commit.com](https://pre-commit.com).
- **Reproducible benchmark** — `make benchmark` runs against a vendored 1000-text
  dataset and emits `bench/results/*.json` + `bench/RESULTS.md`. Adapters for Microsoft
  Presidio (Docker) and spaCy NER (uv venv) included.
- **Property-based testing** with [`pgregory.net/rapid`](https://github.com/flyingmutant/rapid):
  `Reidentify(AnonymizeWithMap(t)) == t` is now a publicly verifiable invariant.
- **Architecture Decision Records** — eight ADRs in [docs/ARCHITECTURE_DECISIONS.md](docs/ARCHITECTURE_DECISIONS.md).
- **Strict linting** — `.golangci.yml` with 13 enabled linters; CI gate.
- **Fuzz tests** — `FuzzScan`, `FuzzAnonymize`, `FuzzReidentify`, `FuzzStream`.

### Changed
- **Breaking:** `NewDetector(cfg)` now returns `(*Detector, error)`. Use
  `MustNewDetector(cfg)` if you prefer the old panic-on-error semantics
  (acceptable for tests and main, never for libraries).
- `pii/detector.go` (200-line monolith) split into `pii/patterns/*.go` (one file per
  pattern family) plus `pii/engine.go` and `pii/overlap.go`.

### Deprecated
- (none)

### Removed
- (none)

### Fixed
- Overlap resolution edge cases (zero-length matches, equal-priority ties) now have
  dedicated unit tests in `pii/overlap_test.go`.

### Security
- All public API entry points are fuzz-tested.
- Plugin `Detect()` panics are recovered and reported as detection-skipped errors;
  they never crash the host process.

## [0.1.0] — 2026-04-19

Initial public release. Regex-based detection of 10 PII types, Luhn validation,
deterministic anonymization with `AnonymizeWithMap` / `Reidentify`, CLI with
`scan` and `anonymize` subcommands.
