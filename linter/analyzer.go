package linter

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"loglinter/rules"
)

const AnalyzerName = "loglint"

var Analyzer = &analysis.Analyzer{
	Name:     AnalyzerName,
	Doc:      "checks log message format and content (lowercase, English, no special chars, no sensitive data)",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

var (
	logLevels = map[string]bool{
		"Info": true, "Debug": true, "Warn": true, "Error": true,
		"InfoContext": true, "DebugContext": true, "WarnContext": true, "ErrorContext": true,
	}
	slogPath = "log/slog"
	zapPath  = "go.uber.org/zap"
)

func run(pass *analysis.Pass) (interface{}, error) {
	inspectResult := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}
	inspectResult.Preorder(nodeFilter, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		msg, pos, isConcat := extractLogMessage(pass, call)
		if msg == nil && !isConcat {
			return
		}
		var s string
		if msg != nil {
			s = *msg
		}
		// Run rules
		if r := rules.CheckLowercase(s); r != "" {
			pass.Reportf(pos, "%s", r)
		}
		if r := rules.CheckEnglish(s); r != "" {
			pass.Reportf(pos, "%s", r)
		}
		if r := rules.CheckNoSpecialChars(s); r != "" {
			pass.Reportf(pos, "%s", r)
		}
		if r := rules.CheckNoSensitiveData(s, isConcat); r != "" {
			pass.Reportf(pos, "%s", r)
		}
	})
	return nil, nil
}

// extractLogMessage returns the log message string (if literal), position, and whether the message is a concatenation (for sensitive data rule).
func extractLogMessage(pass *analysis.Pass, call *ast.CallExpr) (msg *string, pos token.Pos, isConcat bool) {
	// Check if this call is a log call from slog or zap
	if !isLogCall(pass, call) {
		return nil, 0, false
	}
	if len(call.Args) == 0 {
		return nil, 0, false
	}
	arg0 := call.Args[0]
	pos = arg0.Pos()

	// String literal
	if lit, ok := arg0.(*ast.BasicLit); ok && lit.Kind == token.STRING {
		s, err := strconv.Unquote(lit.Value)
		if err != nil {
			return nil, 0, false
		}
		return &s, pos, false
	}

	// Binary expr: "prefix " + var
	if bin, ok := arg0.(*ast.BinaryExpr); ok && bin.Op == token.ADD {
		isConcat = true
		s := concatStringParts(bin)
		if s != "" {
			return &s, pos, true
		}
		return nil, pos, true // still report sensitive if we only have concat pattern
	}

	return nil, 0, false
}

// isLogCall reports whether call is a log-level call from log/slog or go.uber.org/zap.
func isLogCall(pass *analysis.Pass, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	name := sel.Sel.Name
	if !logLevels[name] {
		return false
	}
	switch x := sel.X.(type) {
	case *ast.Ident:
		obj := pass.TypesInfo.Uses[x]
		if obj == nil {
			return false
		}
		if pkg, ok := obj.(*types.PkgName); ok {
			path := pkg.Imported().Path()
			return path == slogPath || path == zapPath
		}
		// Method call: logger.Info(...), type of logger must be *slog.Logger or *zap.Logger
		typ := pass.TypesInfo.TypeOf(x)
		return typeFromLogPackage(typ, slogPath) || typeFromLogPackage(typ, zapPath)
	default:
		// e.g. zap.L().Info(...)
		typ := pass.TypesInfo.TypeOf(sel.X)
		return typeFromLogPackage(typ, slogPath) || typeFromLogPackage(typ, zapPath)
	}
}

func typeFromLogPackage(typ types.Type, pkgPath string) bool {
	if typ == nil {
		return false
	}
	ptr, ok := typ.(*types.Pointer)
	if ok {
		typ = ptr.Elem()
	}
	named, ok := typ.(*types.Named)
	if !ok {
		return false
	}
	return named.Obj().Pkg() != nil && named.Obj().Pkg().Path() == pkgPath
}

// concatStringParts extracts string from a chain of "a" + "b" + x (returns "ab" and caller can treat as concat).
func concatStringParts(e *ast.BinaryExpr) string {
	if e.Op != token.ADD {
		return ""
	}
	var parts []string
	ast.Inspect(e, func(n ast.Node) bool {
		if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING {
			s, _ := strconv.Unquote(lit.Value)
			parts = append(parts, s)
		}
		return true
	})
	return strings.Join(parts, "")
}