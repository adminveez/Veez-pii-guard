.PHONY: help test test-race fuzz fuzz-long bench benchmark lint cover wasm wasm-demo rust install clean

GO ?= go
GOLANGCI_LINT ?= golangci-lint
COVER_THRESHOLD ?= 85
WASM_DIR := dist/wasm
RUST_DIR := engine-rust

help: ## Print this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

test: ## Run unit tests
	$(GO) test ./...

test-race: ## Run tests with the race detector
	$(GO) test -race -count=1 ./...

cover: ## Compute coverage and enforce threshold ($(COVER_THRESHOLD)%)
	$(GO) test -coverprofile=coverage.out -covermode=atomic ./pii/... ./cmd/... ./examples/...
	@total=$$($(GO) tool cover -func=coverage.out | awk '/total:/ {print $$3}' | tr -d '%'); \
	echo "Total coverage: $$total%"; \
	awk -v t=$$total -v th=$(COVER_THRESHOLD) 'BEGIN { if (t+0 < th+0) { printf "FAIL: coverage %s%% < %s%%\n", t, th; exit 1 } }'

fuzz: ## Run fuzzers for 30s each (CI default)
	$(GO) test -run=- -fuzz=FuzzScan -fuzztime=30s ./pii
	$(GO) test -run=- -fuzz=FuzzAnonymize -fuzztime=30s ./pii
	$(GO) test -run=- -fuzz=FuzzReidentify -fuzztime=30s ./pii
	$(GO) test -run=- -fuzz=FuzzStream -fuzztime=30s ./pii

fuzz-long: ## Run fuzzers for 5min each (nightly)
	$(GO) test -run=- -fuzz=FuzzScan -fuzztime=5m ./pii

bench: ## Run Go benchmarks
	$(GO) test -bench=. -benchmem -run=^$$ ./pii

benchmark: ## Run reproducible cross-tool benchmark vs Presidio + spaCy
	$(GO) run ./bench/cmd/run -out bench/results

lint: ## Run golangci-lint with the strict config
	$(GOLANGCI_LINT) run ./...

install: ## Install the CLI locally
	$(GO) install ./cmd/pii-guard

wasm: ## Build the WASM artifact + JS wrapper
	@mkdir -p $(WASM_DIR)
	GOOS=js GOARCH=wasm $(GO) build -trimpath -ldflags="-s -w" -o $(WASM_DIR)/veez-pii-guard.wasm ./wasm
	cp "$$($(GO) env GOROOT)/lib/wasm/wasm_exec.js" $(WASM_DIR)/wasm_exec.js 2>/dev/null || \
	cp "$$($(GO) env GOROOT)/misc/wasm/wasm_exec.js" $(WASM_DIR)/wasm_exec.js
	@echo "WASM built at $(WASM_DIR)/veez-pii-guard.wasm ($$(du -h $(WASM_DIR)/veez-pii-guard.wasm | cut -f1))"

wasm-demo: wasm ## Build WASM and serve the demo on :8080
	@cp $(WASM_DIR)/veez-pii-guard.wasm examples/wasm-demo/
	@cp $(WASM_DIR)/wasm_exec.js examples/wasm-demo/
	@echo "Open http://localhost:8080"
	@cd examples/wasm-demo && python3 -m http.server 8080

rust: ## Build the optional Rust engine cdylib
	cd $(RUST_DIR) && cargo build --release
	@echo "Built $(RUST_DIR)/target/release/libveez_pii_engine.* — run tests with: go test -tags veezrust ./pii"

clean: ## Clean build artifacts
	rm -rf dist coverage.out bench/results/*.json
	-cd $(RUST_DIR) && cargo clean 2>/dev/null
