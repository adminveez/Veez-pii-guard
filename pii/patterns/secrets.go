package patterns

import "regexp"

// APIKey matches "api_key=..." style assignments.
var APIKey = Compiled{
	Name:       "API_KEY",
	Source:     "regex/api-key",
	Confidence: 0.99,
	Regexp:     regexp.MustCompile(`(?i)(?:api[_-]?key|apikey|api_secret|api-secret)[\s:="']+([a-zA-Z0-9_-]{20,})`),
}

// BearerToken matches HTTP Bearer tokens.
var BearerToken = Compiled{
	Name:       "BEARER_TOKEN",
	Source:     "regex/bearer-token",
	Confidence: 0.99,
	Regexp:     regexp.MustCompile(`(?i)bearer\s+[a-zA-Z0-9_.-]{20,}`),
}

// GenericSecret matches generic password/secret/token assignments.
var GenericSecret = Compiled{
	Name:       "SECRET",
	Source:     "regex/generic-secret",
	Confidence: 0.90,
	Regexp:     regexp.MustCompile(`(?i)(?:password|secret|token|passwd|pwd)[\s:="']+[^\s"']{8,}`),
}
