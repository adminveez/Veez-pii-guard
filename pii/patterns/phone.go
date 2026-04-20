package patterns

import "regexp"

// PhoneFR matches French phone numbers in domestic and +33/0033 forms.
var PhoneFR = Compiled{
	Name:       "PHONE",
	Source:     "regex/phone-fr",
	Confidence: 0.95,
	Regexp:     regexp.MustCompile(`(?:\+33[\s.-]?|0033[\s.-]?|0)[1-9][\s.-]?(?:[0-9][\s.-]?){7,12}[0-9]`),
}

// PhoneE164 matches international phone numbers in E.164 form.
var PhoneE164 = Compiled{
	Name:       "PHONE_E164",
	Source:     "regex/phone-e164",
	Confidence: 0.90,
	Regexp:     regexp.MustCompile(`\+[\s.-]?[0-9]{1,4}[\s.-]?[0-9]{1,4}[\s.-]?[0-9]{1,4}[\s.-]?[0-9]{2,4}`),
}
