package linter

import (
	"fmt"
	"regexp"
)

// Config controls which loggers and rules the analyzer checks.
type Config struct {
	// Logger families to check. Empty means all enabled.
	Loggers []string `json:"loggers"`

	// Regex patterns for sensitive data detection. Empty uses defaults.
	SensitivePatterns []string `json:"sensitive_patterns"`

	compiledPatterns []*regexp.Regexp
	loggerSet        map[string]bool
}

var defaultSensitivePatterns = []string{
	`(?i)(password|api_key|apikey|token|secret|credential|bearer|private_key)\s*[:=]`,
}

var defaultLoggers = []string{"slog", "zap", "log"}

// DefaultConfig returns a Config with all loggers and default sensitive patterns enabled.
func DefaultConfig() Config {
	return Config{
		Loggers:           defaultLoggers,
		SensitivePatterns: defaultSensitivePatterns,
	}
}

// Prepare compiles patterns and builds internal lookup sets. Must be called before use.
func (c *Config) Prepare() error {
	if len(c.Loggers) == 0 {
		c.Loggers = defaultLoggers
	}
	c.loggerSet = make(map[string]bool, len(c.Loggers))
	for _, l := range c.Loggers {
		c.loggerSet[l] = true
	}

	if len(c.SensitivePatterns) == 0 {
		c.SensitivePatterns = defaultSensitivePatterns
	}
	c.compiledPatterns = make([]*regexp.Regexp, 0, len(c.SensitivePatterns))
	for _, p := range c.SensitivePatterns {
		re, err := regexp.Compile(p)
		if err != nil {
			return fmt.Errorf("loglint: invalid sensitive pattern %q: %w", p, err)
		}
		c.compiledPatterns = append(c.compiledPatterns, re)
	}
	return nil
}

// LoggerEnabled reports whether the named logger family is active.
func (c *Config) LoggerEnabled(name string) bool {
	if c.loggerSet == nil {
		return true
	}
	return c.loggerSet[name]
}

// MatchesSensitive reports whether msg matches any configured sensitive pattern.
func (c *Config) MatchesSensitive(msg string) bool {
	for _, re := range c.compiledPatterns {
		if re.MatchString(msg) {
			return true
		}
	}
	return false
}

// ParsePluginConfig decodes the raw conf value from golangci-lint's New(conf any).
func ParsePluginConfig(conf any) (Config, error) {
	cfg := DefaultConfig()
	if conf == nil {
		return cfg, nil
	}
	m, ok := conf.(map[string]any)
	if !ok {
		return cfg, nil
	}
	if v, ok := m["loggers"]; ok {
		if arr, ok := v.([]any); ok {
			loggers := make([]string, 0, len(arr))
			for _, item := range arr {
				if s, ok := item.(string); ok {
					loggers = append(loggers, s)
				}
			}
			if len(loggers) > 0 {
				cfg.Loggers = loggers
			}
		}
	}
	if v, ok := m["sensitive_patterns"]; ok {
		if arr, ok := v.([]any); ok {
			patterns := make([]string, 0, len(arr))
			for _, item := range arr {
				if s, ok := item.(string); ok {
					patterns = append(patterns, s)
				}
			}
			if len(patterns) > 0 {
				cfg.SensitivePatterns = patterns
			}
		}
	}
	return cfg, nil
}
