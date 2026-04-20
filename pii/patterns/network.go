package patterns

import "regexp"

// IPAddress matches IPv4 with proper octet bounds.
var IPAddress = Compiled{
	Name:       "IP_ADDRESS",
	Source:     "regex/ipv4",
	Confidence: 0.95,
	Regexp:     regexp.MustCompile(`\b(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\b`),
}
