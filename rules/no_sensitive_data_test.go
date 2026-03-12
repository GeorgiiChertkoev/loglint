package rules

import (
	"regexp"
	"testing"
)

var testSensitivePattern = regexp.MustCompile(`(?i)(password|api_key|apikey|token|secret|credential|bearer|private_key)\s*[:=]`)

func testMatcher(msg string) bool {
	return testSensitivePattern.MatchString(msg)
}

func TestCheckNoSensitiveData(t *testing.T) {
	tests := []struct {
		name   string
		msg    string
		hasVar bool
		want   string
	}{
		{"empty", "", false, ""},
		{"safe message", "user authenticated successfully", false, ""},
		{"token validated", "token validated", false, ""},
		{"concat with var and sensitive keyword", "password: ", true,
			"log message must not contain potentially sensitive data (avoid concatenating variables with sensitive keywords)"},
		{"password in message", "user password: xyz", false,
			"log message must not contain potentially sensitive data (e.g. password, token, api_key)"},
		{"api_key=", "api_key=something", false,
			"log message must not contain potentially sensitive data (e.g. password, token, api_key)"},
		{"token: ", "token: abc", false,
			"log message must not contain potentially sensitive data (e.g. password, token, api_key)"},
		{"literal only concat no sensitive", "not found: api_key", false, ""},
		{"literal only concat with sensitive pattern", "api_key=value", false,
			"log message must not contain potentially sensitive data (e.g. password, token, api_key)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckNoSensitiveData(tt.msg, tt.hasVar, testMatcher); got != tt.want {
				t.Errorf("CheckNoSensitiveData(%q, %v) = %q, want %q", tt.msg, tt.hasVar, got, tt.want)
			}
		})
	}
}
