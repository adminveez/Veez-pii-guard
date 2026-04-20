# Benchmark harness — veez-pii-guard

This directory contains the **reproducible benchmark harness** used to
compare veez-pii-guard against external PII engines on a curated synthetic
corpus.

## Methodology

See [`docs/ARCHITECTURE_DECISIONS.md` ADR-007](../docs/ARCHITECTURE_DECISIONS.md).

- **Dataset**: 1000 deterministic synthetic samples across 5 categories
  (chat, email, log, ticket, doc) with known ground-truth PII spans.
- **Reproducibility**: fixed RNG seed (default `42`).
- **Metrics**: precision / recall / F1 per PII type, plus p50/p95/p99
  latency and throughput (chars/sec).

## Running

```bash
# Pure-Go default engine
go run ./bench/cmd/run --engine=veez --out=bench/results/veez.json

# Opt-in Rust backend
go run -tags veezrust ./bench/cmd/run --engine=veez-rust --out=bench/results/veez-rust.json

# External engines (require local install — see bench/adapters/README.md)
go run ./bench/cmd/run --engine=presidio --out=bench/results/presidio.json
go run ./bench/cmd/run --engine=spacy    --out=bench/results/spacy.json
```

Or use the Makefile shortcut:

```bash
make benchmark
```

## Results

After running, see [`bench/RESULTS.md`](RESULTS.md) for the latest comparison
table. Re-generate with:

```bash
go run ./bench/cmd/render < bench/results/*.json > bench/RESULTS.md
```

## Notes

- The Presidio and spaCy adapters are intentionally external (not vendored)
  to avoid Python dependencies in the Go module. They are invoked via
  `exec` and skipped when the binary is not on `$PATH`.
- The dataset is generated in-memory — no fixtures are committed to keep
  the repo small.
