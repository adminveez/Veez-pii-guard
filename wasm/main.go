//go:build js && wasm

// Command wasm exposes the pii-guard library to JavaScript via syscall/js.
// Built with `make wasm`. See examples/wasm-demo/index.html for usage.
package main

import (
	"context"
	"encoding/json"
	"syscall/js"

	"github.com/veez-ai/veez-pii-guard/pii"
)

func main() {
	js.Global().Set("veezPiiAnonymize", js.FuncOf(anonymize))
	js.Global().Set("veezPiiScan", js.FuncOf(scan))
	js.Global().Set("veezPiiVersion", js.FuncOf(versionFn))
	<-make(chan struct{}) // keep the WASM runtime alive
}

func anonymize(_ js.Value, args []js.Value) any {
	if len(args) == 0 {
		return errResponse("missing text argument")
	}
	text := args[0].String()
	cfg := pii.DefaultConfig()
	cfg.BlockOnPII = false
	cfg.BlockOnSecrets = false
	d := pii.MustNewDetector(cfg)
	res := d.Scan(context.Background(), text)
	res.AnonymizedText, res.Mappings = pii.AnonymizeWithMap(text, res.Detections)
	return marshalToJS(res)
}

func scan(_ js.Value, args []js.Value) any {
	if len(args) == 0 {
		return errResponse("missing text argument")
	}
	text := args[0].String()
	d := pii.MustNewDetector(pii.DefaultConfig())
	res := d.Scan(context.Background(), text)
	return marshalToJS(res)
}

func versionFn(_ js.Value, _ []js.Value) any {
	return "0.2.0"
}

func marshalToJS(v any) any {
	b, err := json.Marshal(v)
	if err != nil {
		return errResponse(err.Error())
	}
	return string(b)
}

func errResponse(msg string) any {
	b, _ := json.Marshal(map[string]string{"error": msg})
	return string(b)
}
