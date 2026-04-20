# Architecture Decision Records

Each ADR follows the [Michael Nygard format](https://cognitect.com/blog/2011/11/15/documenting-architecture-decisions):
**Status / Context / Decision / Consequences / Alternatives considered**.

ADRs are immutable once accepted. New decisions get a new ADR; superseded ones are explicitly marked.

---

## ADR-001 — Pure-Go-first, Rust opt-in


**Context.** Detection runs in security-sensitive paths and is the hot loop of every call. Memory safety matters. Performance matters. But so does `go install github.com/adminveez/Veez-pii-guard/cmd/pii-guard@latest` working on every developer's machine in 2 seconds, with no toolchain dance.

**Decision.** The default build is **pure Go, zero CGo, zero non-stdlib runtime dependency**. A Rust engine lives under `engine-rust/` and is wired in *only* when the user builds with `-tags veezrust`. The two engines implement the same internal `patternMatcher` interface (see ADR-002) and are property-tested for output equivalence.

Package-level `regexp.MustCompile` on string literals is allowed: it is a build-time invariant equivalent to `static_assert`, not a runtime panic surface. Runtime errors (invalid user config, plugin registration conflicts) propagate via `error`.

**Consequences.**
- `go install` and `go get` keep working on every supported platform without a Rust toolchain.
- WASM build (ADR-008) stays trivial: same Go code compiles to `js/wasm`.
- Cross-compilation stays trivial.
- Users who want maximum throughput on long inputs opt in with one build flag.
- We carry the maintenance cost of two engines; mitigated by a thin shared interface and contract tests.

**Alternatives considered.**
- *Rust as the only engine, Go as a thin wrapper.* Rejected: kills `go install`, kills WASM, adds CGo overhead (~250ns/call) on tiny inputs which dominate real workloads.
- *Pure Go forever.* Rejected: leaves measurable performance on the table on `RegexSet`-style multi-pattern matching where the Rust `regex` crate is genuinely faster.

---

## ADR-002 — Decoupled engine behind an internal interface


**Context.** v0.1 had detection logic, anti-overlap logic, anonymization, and config validation in one 200-line `detector.go`. Adding plugins, a Rust engine, and a stream mode without an interface would have produced an unreviewable diff.

**Decision.** Introduce an internal `patternMatcher` interface (`Match(text string) []Match`). The default engine, the Rust engine, and user plugins all implement it. The `Detector` orchestrates: it owns config, dispatches to matchers, runs validation post-filters (Luhn, etc.), runs overlap resolution, and runs anonymization.

**Consequences.**
- New detection backends are plug-in points rather than forks.
- Overlap resolution is testable in isolation (`pii/overlap.go`).
- Per-matcher confidence and provenance survive end-to-end into `Detection.Method`.

**Alternatives considered.**
- *Codegen of a single mega-regex.* Rejected: opaque, slow to compile, and hostile to plugins.

---

## ADR-003 — Runtime plugins over codegen



**Context.** The `PatternPlugin` interface lets third parties register detectors without modifying the core. We could have done it via codegen (build-time registration), but that fights the Go ecosystem.

**Decision.** Plugins are Go types implementing `PatternPlugin`, registered at runtime via `Detector.Register(p)`. They run in-process, share the same `Match` value type, and report their own confidence. Registration is validated (unique name, confidence in `[0,1]`).

**Consequences.**
- Zero build complexity for plugin authors. `import` and `Register`.
- Plugins cannot crash the host: panics in `Detect()` are recovered and logged as a detection-skipped error.
- Plugins are *not* hot-loadable across processes (no `plugin` package, no shared objects); we explicitly do not support that — it would be a security footgun for a security tool.

**Alternatives considered.**
- *Go `plugin` package (.so loading).* Rejected: platform-limited, ABI-fragile, and a code-execution surface we refuse to expose for a PII tool.
- *Build-time codegen.* Rejected: hostile UX for plugin authors.

---

## ADR-004 — No neural NER, contextual heuristics + embedded dictionary



**Context.** Names (especially common French names like *Pierre*, *Rose*, *Olive*) are the long-tail of LLM-prompt PII. Real NER models exist (spaCy, Presidio, ONNX BERT-NER) and would catch more of them.

**Decision.** No neural model is shipped. Name detection is a two-pass system:
1. Regex pass for structured PII (emails, phones, IBAN, IP, dates, etc.).
2. Contextual pass: Unicode tokenizer, sliding 3-token window, capitalization heuristic, lookup against an embedded dictionary of top FR/EN given names (`go:embed`, ~80KB compressed).

Confidence is explicitly capped: `0.65` for a single match, `0.85` when followed by a capitalized token (likely surname). The flag `Config.DetectNames` is **off by default** and the README documents the false-positive trade-off.

**Consequences.**
- Binary stays small (single-digit MB), no ONNX runtime, no Python.
- Reproducible deterministic output (a property we test publicly via `rapid`).
- Lower recall on names than a NER model. This is a *positioning* trade-off, not a defect.

**Alternatives considered.**
- *Embedded ONNX BERT-NER.* Rejected: 100MB+ runtime, slow cold start, defeats `go install`.
- *Optional plugin that calls spaCy over IPC.* Rejected for v1: introduces Python in the hot path. Could be a community plugin later.

---

## ADR-005 — Stream mode: tail-overlap with hash-based dedup


**Context.** LLM contexts process documents of thousands of tokens. A naive chunked scan misses PII that straddles a chunk boundary.

**Decision.** `StreamScanner` keeps a tail buffer equal to `max(longest configured pattern length, overlap)` and prepends it to the next chunk. Detections in the overlap region are deduplicated by a hash of `(absoluteOffset, type, value)`.

The default chunk is 2KB (~512 tokens), overlap is 64 bytes. A property test asserts that scanning a 1MB document via the stream produces the same set of detections as scanning it whole.

**Consequences.**
- Constant memory for arbitrary input size.
- Detections carry an absolute `Offset` so callers can map back to source position.
- Overlap > longest pattern — IBAN can be 34 chars, so overlap defaults are conservative. Configurable for users with custom long patterns.

**Alternatives considered.**
- *No overlap, post-merge.* Rejected: needs the full text in memory anyway to merge by content, defeating the point.
- *Sentence-aware chunking.* Rejected for v0.2: depends on a sentence segmenter; revisit if real-world data shows mid-sentence false negatives.

---

## ADR-006 — Rust engine via cdylib + minimal FFI



**Context.** Phase 5 ships a Rust engine for users who opt in. We need a binding strategy that survives Linux/macOS/Windows, doesn't leak unsafe across the boundary, and is auditable.

**Decision.** A cargo workspace under `engine-rust/` builds `libveez_pii_engine.{so,dylib,dll}` as a `cdylib`. The FFI surface is intentionally tiny:

```c
const uint8_t* veez_pii_scan(const uint8_t* input, size_t len, size_t* out_len);
void veez_pii_free(const uint8_t* ptr, size_t len);
```

Rust returns a JSON-encoded buffer it owns; Go reads it, then calls `veez_pii_free`. No shared structs, no refcount, no callbacks. This is the only safe FFI for a security tool — every byte that crosses the boundary is auditable as plain data.

The Go binding lives behind `//go:build cgo && veezrust`. Without the tag, the Rust code is invisible to `go build`.

**Consequences.**
- CGo overhead (~250ns) makes the Rust path slower for sub-1KB inputs. The dispatcher routes small inputs to pure Go and only sends ≥1KB to Rust. Documented and benchmarked.
- Two engines to maintain; mitigated by contract tests (`engine_contract_test.go`) running both backends against the same corpus.
- Cross-compilation now requires the user to also build the Rust crate for the target. The default Go-only build is unaffected.

**Alternatives considered.**
- *Rich struct sharing across FFI.* Rejected: ABI fragility, alignment hazards, panic safety.
- *Exposing Rust regex VM directly.* Rejected: would leak Rust types into Go.

---

## ADR-007 — Benchmark methodology: reproducible OSS, AWS Comprehend cited but not run


**Context.** Public benchmarks are a rhetorical weapon. Done dishonestly they backfire. We claim parity or wins on specific axes and are honest about losses.

**Decision.** `make benchmark` runs against a vendored 1000-text dataset (Faker-generated FR/EN, seeded, schema-validated). Adapters for Microsoft Presidio (Docker) and spaCy NER (uv venv) are shipped and run locally. AWS Comprehend is **not** run in CI (cost, no offline mode); we cite published numbers with date and link, separated visually in `RESULTS.md`.

Metrics: precision, recall, F1, latency p50/p95, peak RSS.

Known biases are documented (`bench/RESULTS.md` § "Known biases"):
- Dataset is synthetic — overestimates structured PII recall.
- Latency includes adapter IPC for Presidio (Docker round-trip).
- spaCy uses `xx_ent_wiki_sm`; bigger models exist with different trade-offs.

**Consequences.**
- Anyone can reproduce the numbers in 5 minutes on a laptop.
- Disagreements become arguments about methodology (good), not credibility (bad).

**Alternatives considered.**
- *Curated real-world dataset.* Future work; legal review required.

---

## ADR-008 — WASM via Go stdlib (not TinyGo, not Rust→WASM)



**Context.** WASM lets a frontend redact PII before any byte leaves the browser. We had three options: Go stdlib, TinyGo, Rust→WASM.

**Decision.** Go stdlib (`GOOS=js GOARCH=wasm`).

**Consequences.**
- Single source of truth: same Go code runs natively, in WASM, and (with build tag) bridges to Rust.
- Larger `.wasm` payload than TinyGo or Rust (~3-5MB gzipped). Acceptable for v0.2; we ship a streaming loader so the page is interactive immediately.
- `wasm_exec.js` is vendored (pinned to the Go version we test against) to avoid skew.

**Alternatives considered.**
- *TinyGo.* Rejected: regex package compatibility is not guaranteed. Two regex engines diverging in a security tool = bug surface we will not introduce.
- *Rust→WASM.* Rejected: would force the Rust path to be the canonical implementation for browser users, splitting our test surface.
