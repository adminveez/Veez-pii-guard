package patterns

import "regexp"

// ContractRef matches French/English contract references.
var ContractRef = Compiled{
	Name:       "CONTRACT_REF",
	Source:     "regex/contract-ref",
	Confidence: 0.95,
	Regexp:     regexp.MustCompile(`(?:Contrat\s*n[°ºo]?\s*\d{4}[-_]?\d*|Contract\s*#?\s*\d+)`),
}

// CaseRef matches case/dossier references.
var CaseRef = Compiled{
	Name:       "CASE_REF",
	Source:     "regex/case-ref",
	Confidence: 0.95,
	Regexp:     regexp.MustCompile(`(?:Dossier\s*n?[°]?\s*[\w-]+|Affaire\s*\d+/\d{4}|Case\s*#?\s*[\w-]+|Réf\.\s*[\w.-]+)`),
}

// ClientID matches "Client ID:", "N° client", etc.
var ClientID = Compiled{
	Name:       "CLIENT_ID",
	Source:     "regex/client-id",
	Confidence: 0.95,
	Regexp:     regexp.MustCompile(`(?:Client\s*ID|N°\s*client|Numéro\s*client|Customer\s*ID|Client\s*#)[\s:=-]*[\w.-]+`),
}
