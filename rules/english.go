package rules

import "unicode"

// CheckEnglish reports a diagnostic if the log message contains non-English characters.
// Allows ASCII letters, digits, and spaces. Uses unicode ranges for English (Latin).
func CheckEnglish(msg string) string {
	if msg == "" {
		return ""
	}
	for _, r := range msg {
		switch {
		case r >= 'a' && r <= 'z':
			continue
		case r >= 'A' && r <= 'Z':
			continue
		case r >= '0' && r <= '9':
			continue
		case r == ' ' || r == '\t':
			continue
		case r == '.' || r == ',' || r == '-' || r == '_' || r == '\'' || r == '!' || r == '?' || r == ':' || r == '=':
			// Common punctuation in English messages (rule 3 catches excessive, rule 4 catches sensitive)
			continue
		case unicode.Is(unicode.Latin, r):
			// Latin script (covers English)
			continue
		default:
			return "log message must be in English only"
		}
	}
	return ""
}
