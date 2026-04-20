package pii

// RustAvailable reports whether the optional Rust acceleration backend
// was compiled into this binary (build tag `veezrust` + cgo).
//
// The backend is currently used only by the benchmark harness
// (`bench/cmd/run`). Future versions will route high-volume regex passes
// through it automatically when input size exceeds a threshold.
func RustAvailable() bool { return rustAvailable() }

// RustScan exposes the Rust FFI scan for benchmark and integration use.
// Returns nil when the backend is not built in.
func RustScan(input string) ([]Detection, error) { return rustScan(input) }
