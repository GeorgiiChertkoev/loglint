# loglint

A Go linter that checks log messages for consistent style and safety. Works as a standalone binary or as a **golangci-lint** plugin. Supports `log/slog`, `go.uber.org/zap`, and the standard `log` package out of the box, with configurable sensitive-data patterns and auto-fix support.

## Rules

1. **Lowercase** — Log messages must start with a lowercase letter. *(auto-fixable)*
2. **English only** — Messages must be in English (Latin script; no Cyrillic, etc.).
3. **No special characters or emojis** — No repeated punctuation (`!!!`, `...`) or emojis. *(auto-fixable for trailing repeated punctuation)*
4. **No sensitive data** — No patterns like `password:`, `api_key=`, `token:` in the message text. Configurable via regex.

### Supported loggers

- `log/slog`
- `go.uber.org/zap`
- `log` (stdlib)

## Build and run

### Prerequisites

- Go 1.22+

### Standalone

```bash
go build -o loglint .

./loglint ./...

./loglint ./cmd/...
```

### Install

```bash
go install .
# Then run: loglint ./...
```

### CLI flags

```bash
loglint -loggers=slog,zap ./...
loglint -sensitive-patterns='(?i)ssn\s*[:=]' ./...
```

## golangci-lint plugin

The linter can be loaded as a **Go plugin** by golangci-lint.

### Build the plugin

From the project root (Linux or macOS; Go plugin buildmode is not supported on Windows):

```bash
CGO_ENABLED=1 go build -buildmode=plugin -o loglint.so ./plugin
```

Requirements:

- `CGO_ENABLED=1`
- Same Go version and environment as the golangci-lint binary you use
- Build on Linux or macOS

### Configure golangci-lint

Add to `.golangci.yml`:

```yaml
version: "2"

linters:
  default: none
  enable:
    - loglint

linters-settings:
  custom:
    loglint:
      path: ./loglint.so
      description: Log message style and safety checks
      original-url: github.com/GeorgiiChertkoev/loglint
      settings:
        loggers:
          - slog
          - zap
          - log
        sensitive_patterns:
          - '(?i)(password|api_key|token|secret|credential|bearer|private_key)\s*[:=]'
```

Then run:

```bash
golangci-lint run
```

## Configuration

| Setting              | CLI flag               | Plugin setting        | Default |
|----------------------|------------------------|-----------------------|---------|
| Logger families      | `-loggers=slog,zap`    | `loggers: [slog,zap]` | `slog,zap,log` |
| Sensitive patterns   | `-sensitive-patterns=…` | `sensitive_patterns: […]` | password, api_key, token, secret, credential, bearer, private_key |

When custom `sensitive_patterns` are provided they **replace** the defaults entirely.

## Auto-fix

The linter provides `SuggestedFixes` for:

- Capitalised first letter (lowercased automatically)
- Trailing repeated punctuation like `!!!` or `...` (stripped automatically)

Apply fixes with the analysis driver or golangci-lint's `--fix` flag.

## Examples

### Incorrect

```go
slog.Info("Starting server on port 8080")        // must start with lowercase
slog.Error("ошибка подключения")                  // must be English
slog.Info("server started! 🚀")                   // no emojis
slog.Error("connection failed!!!")                // no repeated punctuation
slog.Info("user password: " + password)           // no sensitive data
log.Print("Token: leaked")                        // stdlib log is also checked
```

### Correct

```go
slog.Info("starting server on port 8080")
slog.Error("failed to connect to database")
slog.Info("server started")
slog.Error("connection failed")
slog.Info("user authenticated successfully")
log.Print("request completed")
```

## Tests

```bash
go test ./...
```

## License

MIT
