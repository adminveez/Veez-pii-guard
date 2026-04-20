//go:build !cgo || !veezrust

// Package pii — pure-Go stub when the Rust backend is not built in.
package pii

func rustAvailable() bool { return false }

func rustScan(_ string) ([]Detection, error) { return nil, nil }
