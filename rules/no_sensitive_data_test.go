package rules

import "testing"

func TestCheckNoSensitiveData(t *testing.T) {
	tests := []struct {
		name     string
		msg      string
		isConcat bool
		want     string
	}{
		{"empty", "", false, ""},
		{"safe message", "user authenticated successfully", false, ""},
		{"token validated", "token validated", false, ""},
		{"concat", "anything", true, "log message must not contain potentially sensitive data (avoid concatenating variables with sensitive keywords)"},
		{"password in message", "user password: xyz", false, "log message must not contain potentially sensitive data (e.g. password, token, api_key)"},
		{"api_key=", "api_key=something", false, "log message must not contain potentially sensitive data (e.g. password, token, api_key)"},
		{"token: ", "token: abc", false, "log message must not contain potentially sensitive data (e.g. password, token, api_key)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckNoSensitiveData(tt.msg, tt.isConcat); got != tt.want {
				t.Errorf("CheckNoSensitiveData(%q, %v) = %q, want %q", tt.msg, tt.isConcat, got, tt.want)
			}
		})
	}
}
