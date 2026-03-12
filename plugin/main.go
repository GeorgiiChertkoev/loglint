package main

import (
	"loglint/linter"

	"golang.org/x/tools/go/analysis"
)

// New is the entry point used by golangci-lint when loading a Go plugin (-buildmode=plugin).
func New(conf any) ([]*analysis.Analyzer, error) {
	cfg, err := linter.ParsePluginConfig(conf)
	if err != nil {
		return nil, err
	}
	if err := cfg.Prepare(); err != nil {
		return nil, err
	}
	return []*analysis.Analyzer{linter.NewAnalyzer(&cfg)}, nil
}
