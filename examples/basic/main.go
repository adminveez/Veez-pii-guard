package main

import (
	"context"
	"fmt"

	"github.com/veez-ai/veez-pii-guard/pii"
)

func main() {
	d := pii.NewDetector(pii.DefaultConfig())
	input := "Bonjour, contactez marie.dupont@cabinet-legal.fr au 06 12 34 56 78"
	res := d.Scan(context.Background(), input)
	anonymized, mappings := pii.AnonymizeWithMap(input, res.Detections)
	restored := pii.Reidentify(anonymized, mappings)

	fmt.Println("Anonymized:", anonymized)
	fmt.Println("Restored:", restored)
}
