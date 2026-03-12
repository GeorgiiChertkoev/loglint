package rules

import "testing"

func TestCheckEnglish(t *testing.T) {
	tests := []struct {
		name string
		msg  string
		want string
	}{
		{"empty", "", ""},
		{"english", "starting server on port 8080", ""},
		{"english with punctuation", "failed to connect to database.", ""},
		{"russian", "запуск сервера", "log message must be in English only"},
		{"russian error", "ошибка подключения", "log message must be in English only"},
		{"mixed", "server запущен", "log message must be in English only"},
		{"numbers and space", "port 8080", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckEnglish(tt.msg); got != tt.want {
				t.Errorf("CheckEnglish(%q) = %q, want %q", tt.msg, got, tt.want)
			}
		})
	}
}
