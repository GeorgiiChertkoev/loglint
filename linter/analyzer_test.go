package linter

import (
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestAnalyzer(t *testing.T) {
	testdata := filepath.Join("testdata")
	analysistest.Run(t, testdata, Analyzer, ".")
}
