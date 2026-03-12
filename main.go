package main

import (
	"loglinter/linter"

	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(linter.Analyzer)
}
