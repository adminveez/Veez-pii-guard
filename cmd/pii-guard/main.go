package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/veez-ai/veez-pii-guard/pii"
)

func main() {
	if len(os.Args) == 1 {
		if hasPipedInput() {
			os.Exit(runScan([]string{}))
		}
		printUsage()
		os.Exit(0)
	}

	first := os.Args[1]
	if first == "--help" || first == "-h" || first == "help" {
		printUsage()
		os.Exit(0)
	}

	switch first {
	case "scan":
		os.Exit(runScan(os.Args[2:]))
	case "anonymize":
		os.Exit(runAnonymize(os.Args[2:]))
	default:
		os.Exit(runScan(os.Args[1:]))
	}
}

func runScan(args []string) int {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	text := fs.String("text", "", "Text to scan")
	file := fs.String("file", "", "Path to input file")
	format := fs.String("format", "text", "Output format: text|json")
	blockOnSecrets := fs.Bool("block-on-secrets", true, "Block when secrets are detected")
	blockOnPII := fs.Bool("block-on-pii", false, "Block when any PII is detected")
	threshold := fs.Int("block-threshold", 100, "Block when detection count reaches this threshold")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	input, err := resolveInput(*text, *file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	cfg := pii.DefaultConfig()
	cfg.BlockOnSecrets = *blockOnSecrets
	cfg.BlockOnPII = *blockOnPII
	cfg.BlockThreshold = *threshold

	detector := pii.NewDetector(cfg)
	result := detector.Scan(context.Background(), input)
	result.AnonymizedText, result.Mappings = pii.AnonymizeWithMap(input, result.Detections)
	if err := printResult(result, *format); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if result.Blocked {
		return 2
	}
	return 0
}

func runAnonymize(args []string) int {
	fs := flag.NewFlagSet("anonymize", flag.ContinueOnError)
	text := fs.String("text", "", "Text to anonymize")
	file := fs.String("file", "", "Path to input file")
	format := fs.String("format", "text", "Output format: text|json")
	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	input, err := resolveInput(*text, *file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	cfg := pii.DefaultConfig()
	cfg.BlockOnPII = false
	cfg.BlockOnSecrets = false
	detector := pii.NewDetector(cfg)
	result := detector.Scan(context.Background(), input)
	result.AnonymizedText, result.Mappings = pii.AnonymizeWithMap(input, result.Detections)

	switch *format {
	case "text":
		fmt.Println(result.AnonymizedText)
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(result); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}
	default:
		fmt.Fprintln(os.Stderr, "invalid format, expected text|json")
		return 1
	}

	return 0
}

func resolveInput(textFlag, fileFlag string) (string, error) {
	if textFlag != "" {
		return textFlag, nil
	}
	if fileFlag != "" {
		b, err := os.ReadFile(fileFlag)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	if hasPipedInput() {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(string(b)), nil
	}
	return "", errors.New("provide --text or --file")
}

func hasPipedInput() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) == 0
}

func printResult(result pii.Result, format string) error {
	switch format {
	case "text":
		fmt.Printf("detections: %d\n", result.PIICount)
		fmt.Printf("blocked: %t\n", result.Blocked)
		if result.BlockReason != "" {
			fmt.Printf("reason: %s\n", result.BlockReason)
		}
		fmt.Println("anonymized:")
		fmt.Println(result.AnonymizedText)
		return nil
	case "json":
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	default:
		return errors.New("invalid format, expected text|json")
	}
}

func printUsage() {
	fmt.Println("pii-guard <command> [flags]")
	fmt.Println("commands:")
	fmt.Println("  scan       detect pii and optionally block")
	fmt.Println("  anonymize  print anonymized text")
}
