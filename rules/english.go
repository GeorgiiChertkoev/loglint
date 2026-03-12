package rules

import "unicode"

// CheckEnglish reports a diagnostic if the log message contains non-English characters.
// Allows ASCII letters, digits, spaces, and common English punctuation.
func CheckEnglish(msg string) string {
	if msg == "" {
		return ""
	}
	for _, r := range msg {
		if unicode.IsLetter(r) {
			// Accept only Latin script letters for English
			if !unicode.In(r, unicode.Latin) {
				return "log message must be in English only"
			}
			continue
		}
		if unicode.IsDigit(r) || unicode.IsSpace(r) {
			continue
		}
		switch r {
		case '.', ',', '-', '_', '\'', '!', '?', ':', '=':
			// Accept common English punctuation
			continue
		}
		// Any other rune is considered non-English for this rule.
		return "log message must be in English only"
	}
	return ""
}
