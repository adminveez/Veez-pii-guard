//go:build cgo && veezrust

// Package pii — Rust-accelerated regex pass.
//
// Active only when both `cgo` and the build tag `veezrust` are set:
//
//	CGO_ENABLED=1 go build -tags veezrust ./...
//
// Requires libveez_pii_engine.{a|so|dylib} in the linker path
// (build via `make rust`).
package pii

/*
#cgo LDFLAGS: -L${SRCDIR}/../engine-rust/target/release -lveez_pii_engine
#include <stdint.h>
#include <stdlib.h>

extern int32_t veez_pii_scan(const uint8_t* input_ptr, size_t input_len,
                             uint8_t** out_ptr, size_t* out_len);
extern void veez_pii_free(uint8_t* ptr, size_t len);
*/
import "C"

import (
	"encoding/json"
	"errors"
	"unsafe"
)

type rustDetection struct {
	Type       Type    `json:"type"`
	Text       string  `json:"text"`
	Start      int     `json:"start"`
	End        int     `json:"end"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source"`
}

// rustScan invokes the Rust FFI to scan input for high-volume PII (email,
// phone_fr, ipv4). Returns a slice of Detections compatible with the Go
// engine's overlap-resolution pipeline.
//
// rustAvailable returns true on this build; the purego stub returns false.
func rustAvailable() bool { return true }

func rustScan(input string) ([]Detection, error) {
	if input == "" {
		return nil, nil
	}
	b := []byte(input)
	var outPtr *C.uint8_t
	var outLen C.size_t
	rc := C.veez_pii_scan(
		(*C.uint8_t)(unsafe.Pointer(&b[0])),
		C.size_t(len(b)),
		&outPtr,
		&outLen,
	)
	if rc != 0 {
		return nil, errors.New("veez_pii_scan failed")
	}
	defer C.veez_pii_free(outPtr, outLen)
	jsonBytes := C.GoBytes(unsafe.Pointer(outPtr), C.int(outLen))
	var raw []rustDetection
	if err := json.Unmarshal(jsonBytes, &raw); err != nil {
		return nil, err
	}
	out := make([]Detection, 0, len(raw))
	for _, r := range raw {
		out = append(out, Detection{
			Type:       r.Type,
			Text:       r.Text,
			Start:      r.Start,
			End:        r.End,
			Confidence: r.Confidence,
			Source:     r.Source,
		})
	}
	return out, nil
}
