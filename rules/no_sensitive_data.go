package rules

import "regexp"

var (
	// Match "key: " or "key=" in message (case-insensitive) — suggests logging a secret value
	sensitivePattern = regexp.MustCompile(`(?i)(password|api_key|apikey|token|secret|credential|bearer|private_key)\s*[:=]`)
)

// CheckNoSensitiveData reports if the message or pattern suggests logging sensitive data.
// isConcat is true when the log argument was a string concatenation (e.g. "msg " + var).
func CheckNoSensitiveData(msg string, isConcat bool) string {
	if isConcat {
		return "log message must not contain potentially sensitive data (avoid concatenating variables with sensitive keywords)"
	}
	if sensitivePattern.MatchString(msg) {
		return "log message must not contain potentially sensitive data (e.g. password, token, api_key)"
	}
	return ""
}
