package rules

import "testing"

func TestCheckLowercase(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want string
	}{
		{"empty", "", ""},
		{"whitespace", "   ", ""},
		{"starts lower", "starting server", ""},
		{"starts upper", "Starting server", "log message must start with a lowercase letter"},
		{"starts upper Error", "Failed to connect", "log message must start with a lowercase letter"},
		{"digit start", "8080 port", ""},
		{"punctuation then upper", ". Start", ""}, // trim space, first rune is .
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckLowercase(tt.msg); got != tt.want {
				t.Errorf("CheckLowercase(%q) = %q, want %q", tt.msg, got, tt.want)
			}
		})
	}
}
