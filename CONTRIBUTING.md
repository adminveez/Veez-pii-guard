# Contributing to veez-pii-guard

Thank you for considering a contribution. This project is part of the broader
[VEEZ](https://github.com/adminveez) sovereign-EU AI initiative — built by one
founder, in the open, with a strong bias for privacy, simplicity, and
production-grade Go.

---

## Code of Conduct

By participating you agree to uphold the [Code of Conduct](CODE_OF_CONDUCT.md).
Be respectful, assume good faith, and prefer concrete technical arguments over
opinions.

---

## How to contribute

### 1. Open an issue first (for non-trivial changes)

For anything beyond a typo or a small bug fix, please open an issue describing:

- The problem or use-case
- Your proposed approach
- Compatibility impact (public API, config, plugin contract)

This avoids duplicated work and lets us align on direction early.

### 2. Fork, branch, code

```bash
git clone https://github.com/<you>/Veez-pii-guard.git
cd Veez-pii-guard
git checkout -b feat/my-improvement
```

Branch naming:

- `feat/...`     new feature
- `fix/...`      bug fix
- `docs/...`     documentation only
- `perf/...`     performance work
- `refactor/...` internal restructuring without behavior change
- `test/...`     test-only changes

### 3. Local checks (must pass before opening a PR)

```bash
# Format
gofmt -l . | tee /tmp/gofmt && [ ! -s /tmp/gofmt ]

# Vet
GOWORK=off go vet ./...

# Tests (race detector on Linux/macOS)
GOWORK=off go test -race ./...

# Lint (golangci-lint v2.x)
GOWORK=off golangci-lint run --timeout=2m

# Optional: fuzz smoke test
GOWORK=off go test -run=^$ -fuzz=FuzzScan -fuzztime=10s ./pii
```

If you touch the Rust crate:

```bash
cd engine-rust && cargo test && cargo clippy -- -D warnings
```

If you touch the WASM target:

```bash
GOOS=js GOARCH=wasm go build -o /tmp/veez.wasm ./wasm
```

### 4. Commit messages

Use clear, conventional-style prefixes:

```
feat(pii): add detection for French driver licence numbers
fix(stream): handle CR/LF boundary across chunks
docs(readme): clarify plugin interface
perf(engine): pre-compile regexes once
refactor(types): extract Detection.Source helper
test(plugins): add fuzz corpus for IBAN
chore(ci): bump golangci-lint-action to v7
```

Keep commits focused. Squash WIP commits before requesting review.

### 5. Open the Pull Request

A good PR description includes:

- **What** changed and **why**
- **Tests** added or updated
- **Compatibility** notes (breaking? migration?)
- **Benchmarks** (if perf-related — `bench/cmd/run` is the harness)

PRs are reviewed against the [ADRs](docs/adr/) and the v0.2 quality bar:
zero-dependency core, deterministic behavior, no network in default path.

---

## What we welcome

- New detection plugins (national IDs, healthcare IDs, financial identifiers)
- Better recall/precision on existing types — with a benchmark to back it up
- Bug fixes with reproducing tests
- Documentation, examples, and tutorials
- Performance work (regex tuning, allocation reduction)
- Translations of user-facing docs

## What we are cautious about

- New runtime dependencies (the core stays zero-dep on purpose)
- Behavior changes that are not opt-in or versioned
- Network calls or telemetry
- Vendor-locked integrations

If you are unsure, open an issue first — we are happy to discuss.

---

## Plugin authors

Plugins live under `pii/plugins/<name>/`. The contract is documented in
[ADR-005](docs/adr/0005-plugin-architecture.md). Minimum requirements:

- Implement `pii.Plugin` interface
- Ship unit tests with positive **and** negative cases
- Add a fuzz target if the detection is regex-based
- Document false-positive trade-offs in the package doc comment

---

## Security

Please do **not** open public issues for security vulnerabilities.
Follow the disclosure process in [SECURITY.md](SECURITY.md).

---

## License

By contributing you agree that your contributions will be licensed under the
project's [LICENSE](LICENSE) (Apache-2.0).
# Contributing

1. Fork the repository and create a feature branch.
2. Keep changes focused and documented.
3. Run `GOWORK=off go test ./...` before opening a PR.
4. Add tests for any new behavior.
5. Keep public APIs backward compatible when possible.
6. Use clear commit messages.
7. Update README/examples when behavior changes.
8. Ensure no secrets or private infrastructure data are added.
9. Open a PR with context, test evidence, and expected impact.
10. Be respectful in review and discussion.
