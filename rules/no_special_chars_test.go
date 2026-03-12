package rules

import "testing"

func TestCheckNoSpecialChars(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want string
	}{
		{"empty", "", ""},
		{"plain", "server started", ""},
		{"single punctuation", "connection failed.", ""},
		{"double bang", "error!!", "log message must not contain repeated punctuation (e.g. !!! or ...)"},
		{"triple dots", "something went wrong...", "log message must not contain repeated punctuation (e.g. !!! or ...)"},
		{"emoji", "server started \U0001F680", "log message must not contain emojis or special symbols"},
		{"warning prefix", "warning: something wrong", "log message should not use 'warning:' prefix; use plain text"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckNoSpecialChars(tt.msg); got != tt.want {
				t.Errorf("CheckNoSpecialChars(%q) = %q, want %q", tt.msg, got, tt.want)
			}
		})
	}
}
