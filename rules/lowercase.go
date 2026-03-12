package rules

import (
	"strings"
	"unicode"
)

// CheckLowercase reports a diagnostic if the log message does not start with a lowercase letter.
// Empty or whitespace-only messages are not reported.
func CheckLowercase(msg string) string {
	trimmed := strings.TrimSpace(msg)
	runes := []rune(trimmed)
	if len(runes) == 0 {
		return ""
	}
	first := runes[0]
	if unicode.IsLetter(first) && !unicode.IsLower(first) {
		return "log message must start with a lowercase letter"
	}
	return ""
}
