package rules

import (
	"regexp"
	"strings"
)

// CheckNoSpecialChars reports a diagnostic if the message contains special characters or emojis.
// Disallows: multiple punctuation in a row (!!!, ...), emojis, and excessive punctuation.
var (
	multiPunct = regexp.MustCompile(`[!?.]{2,}`) // !! or ... or ?? etc.
)

func isEmoji(r rune) bool {
	// Emoji and symbol ranges in Unicode
	if r >= 0x2600 && r <= 0x26FF {
		return true // Misc symbols
	}
	if r >= 0x2700 && r <= 0x27BF {
		return true // Dingbats
	}
	if r >= 0x1F300 && r <= 0x1F9FF {
		return true // Misc Symbols, Pictographs, etc.
	}
	if r >= 0x1F600 && r <= 0x1F64F {
		return true // Emoticons
	}
	if r >= 0x1F1E0 && r <= 0x1F1FF {
		return true // Flags
	}
	return false
}

// CheckNoSpecialChars reports if the message has disallowed special chars or emojis.
func CheckNoSpecialChars(msg string) string {
	if msg == "" {
		return ""
	}
	trimmed := strings.TrimSpace(msg)
	if multiPunct.MatchString(trimmed) {
		return "log message must not contain repeated punctuation (e.g. !!! or ...)"
	}
	for _, r := range msg {
		if isEmoji(r) {
			return "log message must not contain emojis or special symbols"
		}
	}
	// Disallow leading "warning:" style prefix with colon (per spec example)
	if strings.HasPrefix(strings.ToLower(trimmed), "warning:") {
		return "log message should not use 'warning:' prefix; use plain text"
	}
	return ""
}
