package linter

import (
	"go/ast"
	"go/token"
	"go/types"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"loglint/rules"
)

const AnalyzerName = "loglint"

// Analyzer is the default analyzer instance with all loggers and default config.
// Use NewAnalyzer to create one with custom config.
var Analyzer *analysis.Analyzer

func init() {
	cfg := DefaultConfig()
	_ = cfg.Prepare()
	Analyzer = newAnalyzerFromConfig(&cfg)
}

// NewAnalyzer creates an analyzer with the given config. Config.Prepare must
// have been called before passing it in.
func NewAnalyzer(cfg *Config) *analysis.Analyzer {
	return newAnalyzerFromConfig(cfg)
}

func newAnalyzerFromConfig(cfg *Config) *analysis.Analyzer {
	a := &analysis.Analyzer{
		Name:     AnalyzerName,
		Doc:      "checks log message format and content (lowercase, English, no special chars, no sensitive data)",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}

	// Register CLI flags that mirror config fields.
	a.Flags.String("loggers", strings.Join(cfg.Loggers, ","),
		"comma-separated logger families to check (slog,zap,log)")
	a.Flags.String("sensitive-patterns", strings.Join(cfg.SensitivePatterns, ","),
		"comma-separated regexes for sensitive data detection")

	a.Run = func(pass *analysis.Pass) (interface{}, error) {
		return run(pass, cfg)
	}
	return a
}

// logMessage holds extracted info about a log call's message argument.
type logMessage struct {
	text     string        // decoded message (concatenated string parts)
	lit      *ast.BasicLit // non-nil only when the arg is a single string literal
	isConcat bool          // true when the arg is a "..." + ... expression
	hasVar   bool          // true when concatenation includes non-literal operands
	pos      token.Pos
}

// trailingRepeatedPunct matches trailing repeated punctuation like !!! or ...
var trailingRepeatedPunct = regexp.MustCompile(`[!?.]{2,}$`)

func run(pass *analysis.Pass, cfg *Config) (interface{}, error) {
	inspectResult := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	nodeFilter := []ast.Node{(*ast.CallExpr)(nil)}

	inspectResult.Preorder(nodeFilter, func(n ast.Node) {
		call := n.(*ast.CallExpr)
		lm := extractLogMessage(pass, call, cfg)
		if lm == nil {
			return
		}

		msg := lm.text
		canFix := lm.lit != nil

		// Pre-compute the combined fix so we can attach it to diagnostics.
		var fixedMsg string
		needsFix := false
		if canFix {
			fixedMsg = msg
			if rules.CheckLowercase(msg) != "" {
				fixedMsg = lowercaseFirst(fixedMsg)
				needsFix = true
			}
			if rules.CheckNoSpecialChars(msg) != "" && trailingRepeatedPunct.MatchString(strings.TrimSpace(msg)) {
				fixedMsg = strings.TrimSpace(trailingRepeatedPunct.ReplaceAllString(fixedMsg, ""))
				needsFix = true
			}
		}

		fixApplied := false
		maybeFix := func() []analysis.SuggestedFix {
			if !needsFix || !canFix || fixApplied {
				return nil
			}
			fixApplied = true
			return []analysis.SuggestedFix{{
				Message: "fix log message",
				TextEdits: []analysis.TextEdit{{
					Pos:     lm.lit.Pos(),
					End:     lm.lit.End(),
					NewText: []byte(strconv.Quote(fixedMsg)),
				}},
			}}
		}

		if r := rules.CheckLowercase(msg); r != "" {
			pass.Report(analysis.Diagnostic{
				Pos:            lm.pos,
				Message:        r,
				SuggestedFixes: maybeFix(),
			})
		}

		if r := rules.CheckEnglish(msg); r != "" {
			pass.Report(analysis.Diagnostic{
				Pos:     lm.pos,
				Message: r,
			})
		}

		if r := rules.CheckNoSpecialChars(msg); r != "" {
			pass.Report(analysis.Diagnostic{
				Pos:            lm.pos,
				Message:        r,
				SuggestedFixes: maybeFix(),
			})
		}

		if r := rules.CheckNoSensitiveData(msg, lm.hasVar, cfg.MatchesSensitive); r != "" {
			pass.Report(analysis.Diagnostic{
				Pos:     lm.pos,
				Message: r,
			})
		}
	})
	return nil, nil
}

func lowercaseFirst(s string) string {
	trimmed := strings.TrimLeftFunc(s, unicode.IsSpace)
	prefix := s[:len(s)-len(trimmed)]
	if trimmed == "" {
		return s
	}
	r, size := utf8.DecodeRuneInString(trimmed)
	if unicode.IsLetter(r) && !unicode.IsLower(r) {
		return prefix + string(unicode.ToLower(r)) + trimmed[size:]
	}
	return s
}

// --- Logger families ---

type loggerFamily struct {
	pkgPath string
	methods map[string]bool
}

var (
	slogFamily = loggerFamily{
		pkgPath: "log/slog",
		methods: map[string]bool{
			"Info": true, "Debug": true, "Warn": true, "Error": true,
			"InfoContext": true, "DebugContext": true, "WarnContext": true, "ErrorContext": true,
		},
	}
	zapFamily = loggerFamily{
		pkgPath: "go.uber.org/zap",
		methods: map[string]bool{
			"Info": true, "Debug": true, "Warn": true, "Error": true,
			"DPanic": true, "Panic": true, "Fatal": true,
		},
	}
	stdlogFamily = loggerFamily{
		pkgPath: "log",
		methods: map[string]bool{
			"Print": true, "Printf": true, "Println": true,
			"Fatal": true, "Fatalf": true, "Fatalln": true,
			"Panic": true, "Panicf": true, "Panicln": true,
		},
	}
)

func familiesForConfig(cfg *Config) []loggerFamily {
	var out []loggerFamily
	if cfg.LoggerEnabled("slog") {
		out = append(out, slogFamily)
	}
	if cfg.LoggerEnabled("zap") {
		out = append(out, zapFamily)
	}
	if cfg.LoggerEnabled("log") {
		out = append(out, stdlogFamily)
	}
	return out
}

// extractLogMessage checks whether call is a log call from one of the enabled
// families and extracts the message argument.
func extractLogMessage(pass *analysis.Pass, call *ast.CallExpr, cfg *Config) *logMessage {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return nil
	}
	methodName := sel.Sel.Name

	families := familiesForConfig(cfg)
	matched := false
	for _, fam := range families {
		if !fam.methods[methodName] {
			continue
		}
		if matchesFamily(pass, sel, fam.pkgPath) {
			matched = true
			break
		}
	}
	if !matched {
		return nil
	}
	if len(call.Args) == 0 {
		return nil
	}

	arg0 := call.Args[0]
	pos := arg0.Pos()

	if lit, ok := arg0.(*ast.BasicLit); ok && lit.Kind == token.STRING {
		s, err := strconv.Unquote(lit.Value)
		if err != nil {
			return nil
		}
		return &logMessage{text: s, lit: lit, pos: pos}
	}

	if bin, ok := arg0.(*ast.BinaryExpr); ok && bin.Op == token.ADD {
		text, hasVar := concatParts(bin)
		return &logMessage{text: text, isConcat: true, hasVar: hasVar, pos: pos}
	}

	return nil
}

// matchesFamily reports whether sel.X resolves to pkgPath (as package name or receiver type).
func matchesFamily(pass *analysis.Pass, sel *ast.SelectorExpr, pkgPath string) bool {
	switch x := sel.X.(type) {
	case *ast.Ident:
		obj := pass.TypesInfo.Uses[x]
		if obj == nil {
			return false
		}
		if pkg, ok := obj.(*types.PkgName); ok {
			return pkg.Imported().Path() == pkgPath
		}
		typ := pass.TypesInfo.TypeOf(x)
		return typeFromPkg(typ, pkgPath)
	default:
		typ := pass.TypesInfo.TypeOf(sel.X)
		return typeFromPkg(typ, pkgPath)
	}
}

func typeFromPkg(typ types.Type, pkgPath string) bool {
	if typ == nil {
		return false
	}
	if ptr, ok := typ.(*types.Pointer); ok {
		typ = ptr.Elem()
	}
	named, ok := typ.(*types.Named)
	if !ok {
		return false
	}
	return named.Obj().Pkg() != nil && named.Obj().Pkg().Path() == pkgPath
}

// concatParts walks a + chain and returns the joined string literal parts
// and whether any non-literal operand was found.
func concatParts(e *ast.BinaryExpr) (text string, hasVar bool) {
	if e.Op != token.ADD {
		return "", false
	}
	var parts []string
	ast.Inspect(e, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.BasicLit:
			if v.Kind == token.STRING {
				s, _ := strconv.Unquote(v.Value)
				parts = append(parts, s)
			}
		case *ast.Ident:
			hasVar = true
		case *ast.CallExpr:
			hasVar = true
		}
		return true
	})
	return strings.Join(parts, ""), hasVar
}
