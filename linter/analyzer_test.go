package linter

import (
	"os"
	"path/filepath"
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func testdataDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(wd, "testdata")
}

func TestAnalyzer(t *testing.T) {
	analysistest.Run(t, testdataDir(t), Analyzer, "a")
}

func TestAnalyzerFixes(t *testing.T) {
	analysistest.RunWithSuggestedFixes(t, testdataDir(t), Analyzer, "a")
}
