package main

import (
	"loglinter/linter"

	"golang.org/x/tools/go/analysis"
)

// New returns the list of analyzers for the loglint plugin.
// This is the entry point used by golangci-lint when loading a Go plugin (-buildmode=plugin).
func New(conf any) ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{linter.Analyzer}, nil
}
