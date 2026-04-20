// Package dataset generates a deterministic synthetic corpus of texts
// containing known PII spans, suitable for benchmarking PII detectors.
//
// The corpus is reproducible via a fixed RNG seed so that successive
// benchmark runs (and CI runs across machines) compare apples to apples.
//
// Categories:
//
//   - chat:   short conversational messages with phones/emails/IPs
//   - email:  structured From/To/Subject/Body blocks
//   - log:    JSON log lines with API keys and IPs
//   - ticket: support tickets with IBAN, SSN, names
//   - doc:    long-form documents with mixed PII
package dataset

import (
	"fmt"
	"math/rand"
)

// Type mirrors pii.Type but is duplicated here to keep the dataset
// package zero-dependency on the engine being benchmarked.
type Type string

// Span is a ground-truth PII annotation: byte offsets [Start, End)
// and the PII Type expected at that span.
type Span struct {
	Start int  `json:"start"`
	End   int  `json:"end"`
	Type  Type `json:"type"`
}

// Sample is one synthetic input plus its ground-truth annotations.
type Sample struct {
	Category string `json:"category"`
	Text     string `json:"text"`
	Truth    []Span `json:"truth"`
}

// Generate builds n synthetic samples deterministically from `seed`.
// Distribution: 30% chat, 20% email, 20% log, 15% ticket, 15% doc.
func Generate(seed int64, n int) []Sample {
	r := rand.New(rand.NewSource(seed))
	out := make([]Sample, 0, n)
	for i := 0; i < n; i++ {
		switch {
		case i%10 < 3:
			out = append(out, genChat(r))
		case i%10 < 5:
			out = append(out, genEmail(r))
		case i%10 < 7:
			out = append(out, genLog(r))
		case i%10 < 8:
			out = append(out, genTicket(r))
		default:
			out = append(out, genDoc(r))
		}
	}
	return out
}

func genChat(r *rand.Rand) Sample {
	email := fakeEmail(r)
	prefix := "hey, can you reach me at "
	text := prefix + email + " ? thanks"
	return Sample{
		Category: "chat",
		Text:     text,
		Truth:    []Span{{Start: len(prefix), End: len(prefix) + len(email), Type: "EMAIL"}},
	}
}

func genEmail(r *rand.Rand) Sample {
	from := fakeEmail(r)
	to := fakeEmail(r)
	body := "Hi, please confirm receipt at " + from + ". Best."
	header := fmt.Sprintf("From: %s\nTo: %s\nSubject: hi\n\n", from, to)
	text := header + body
	truth := []Span{
		{Start: 6, End: 6 + len(from), Type: "EMAIL"},
		{Start: len("From: ") + len(from) + len("\nTo: "), End: len("From: ") + len(from) + len("\nTo: ") + len(to), Type: "EMAIL"},
	}
	bodyOff := len(header) + len("Hi, please confirm receipt at ")
	truth = append(truth, Span{Start: bodyOff, End: bodyOff + len(from), Type: "EMAIL"})
	return Sample{Category: "email", Text: text, Truth: truth}
}

func genLog(r *rand.Rand) Sample {
	ip := fakeIP(r)
	key := fakeAPIKey(r)
	text := fmt.Sprintf(`{"ts":"2025-04-20T01:00:00Z","src":"%s","auth":"%s","msg":"ok"}`, ip, key)
	srcOff := len(`{"ts":"2025-04-20T01:00:00Z","src":"`)
	authOff := srcOff + len(ip) + len(`","auth":"`)
	return Sample{
		Category: "log",
		Text:     text,
		Truth: []Span{
			{Start: srcOff, End: srcOff + len(ip), Type: "IP"},
			{Start: authOff, End: authOff + len(key), Type: "API_KEY"},
		},
	}
}

func genTicket(r *rand.Rand) Sample {
	phone := fakePhoneFR(r)
	prefix := "Customer reports issue. Reach them at "
	text := prefix + phone + " between 9-17h."
	return Sample{
		Category: "ticket",
		Text:     text,
		Truth:    []Span{{Start: len(prefix), End: len(prefix) + len(phone), Type: "PHONE"}},
	}
}

func genDoc(r *rand.Rand) Sample {
	email := fakeEmail(r)
	ip := fakeIP(r)
	body := "This is a long document. Contact: " + email + ". Server IP: " + ip + ". End."
	emailOff := len("This is a long document. Contact: ")
	ipOff := emailOff + len(email) + len(". Server IP: ")
	return Sample{
		Category: "doc",
		Text:     body,
		Truth: []Span{
			{Start: emailOff, End: emailOff + len(email), Type: "EMAIL"},
			{Start: ipOff, End: ipOff + len(ip), Type: "IP"},
		},
	}
}

func fakeEmail(r *rand.Rand) string {
	first := []string{"alice", "bob", "carol", "dave", "eve", "marie", "jean"}[r.Intn(7)]
	last := []string{"durand", "smith", "patel", "kim", "garcia", "ono", "ivanov"}[r.Intn(7)]
	dom := []string{"example.com", "veez.io", "test.org", "mail.fr"}[r.Intn(4)]
	return fmt.Sprintf("%s.%s@%s", first, last, dom)
}

func fakePhoneFR(r *rand.Rand) string {
	return fmt.Sprintf("+33 6 %02d %02d %02d %02d", r.Intn(100), r.Intn(100), r.Intn(100), r.Intn(100))
}

func fakeIP(r *rand.Rand) string {
	return fmt.Sprintf("%d.%d.%d.%d", r.Intn(223)+1, r.Intn(256), r.Intn(256), r.Intn(254)+1)
}

func fakeAPIKey(r *rand.Rand) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 40)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return "sk_live_" + string(b)
}
