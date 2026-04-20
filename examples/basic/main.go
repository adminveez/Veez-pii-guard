// Command basic is a minimal end-to-end example of using veez-pii-guard
// to scan and anonymize a string of text.
package main

import (
	"context"
	"fmt"

	"github.com/adminveez/Veez-pii-guard/pii"
)

func main() {
	d := pii.MustNewDetector(pii.DefaultConfig())
	input := "Bonjour, contactez marie.dupont@cabinet-legal.fr au 06 12 34 56 78"
	res := d.Scan(context.Background(), input)
	anonymized, mappings := pii.AnonymizeWithMap(input, res.Detections)
	restored := pii.Reidentify(anonymized, mappings)

	fmt.Println("Anonymized:", anonymized)
	fmt.Println("Restored:", restored)
}
