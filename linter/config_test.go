package linter

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if err := cfg.Prepare(); err != nil {
		t.Fatal(err)
	}
	if !cfg.LoggerEnabled("slog") {
		t.Error("slog should be enabled by default")
	}
	if !cfg.LoggerEnabled("zap") {
		t.Error("zap should be enabled by default")
	}
	if !cfg.LoggerEnabled("log") {
		t.Error("log should be enabled by default")
	}
	if cfg.LoggerEnabled("custom") {
		t.Error("custom should not be enabled by default")
	}
	if !cfg.MatchesSensitive("password: x") {
		t.Error("should match default sensitive pattern")
	}
	if cfg.MatchesSensitive("hello world") {
		t.Error("should not match non-sensitive message")
	}
}

func TestCustomConfig(t *testing.T) {
	cfg := Config{
		Loggers:           []string{"slog"},
		SensitivePatterns: []string{`(?i)ssn\s*[:=]`},
	}
	if err := cfg.Prepare(); err != nil {
		t.Fatal(err)
	}
	if !cfg.LoggerEnabled("slog") {
		t.Error("slog should be enabled")
	}
	if cfg.LoggerEnabled("zap") {
		t.Error("zap should not be enabled")
	}
	if !cfg.MatchesSensitive("ssn: 123") {
		t.Error("should match custom pattern")
	}
	if cfg.MatchesSensitive("password: x") {
		t.Error("should not match default pattern when custom is set")
	}
}

func TestParsePluginConfig(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		cfg, err := ParsePluginConfig(nil)
		if err != nil {
			t.Fatal(err)
		}
		if err := cfg.Prepare(); err != nil {
			t.Fatal(err)
		}
		if !cfg.LoggerEnabled("slog") {
			t.Error("slog should be enabled by default")
		}
	})

	t.Run("with loggers", func(t *testing.T) {
		raw := map[string]any{
			"loggers": []any{"zap"},
		}
		cfg, err := ParsePluginConfig(raw)
		if err != nil {
			t.Fatal(err)
		}
		if err := cfg.Prepare(); err != nil {
			t.Fatal(err)
		}
		if cfg.LoggerEnabled("slog") {
			t.Error("slog should not be enabled")
		}
		if !cfg.LoggerEnabled("zap") {
			t.Error("zap should be enabled")
		}
	})

	t.Run("with custom patterns", func(t *testing.T) {
		raw := map[string]any{
			"sensitive_patterns": []any{`(?i)ssn\s*=`},
		}
		cfg, err := ParsePluginConfig(raw)
		if err != nil {
			t.Fatal(err)
		}
		if err := cfg.Prepare(); err != nil {
			t.Fatal(err)
		}
		if !cfg.MatchesSensitive("ssn=123") {
			t.Error("should match custom pattern")
		}
	})
}

func TestInvalidPattern(t *testing.T) {
	cfg := Config{
		SensitivePatterns: []string{`[invalid`},
	}
	if err := cfg.Prepare(); err == nil {
		t.Error("expected error for invalid regex")
	}
}
