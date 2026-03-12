# loglinter

A Go linter that checks log messages for consistent style and safety. Compatible with **golangci-lint** as a custom plugin.

## Rules

1. **Lowercase** — Log messages must start with a lowercase letter.
2. **English only** — Messages must be in English (Latin script; no Cyrillic, etc.).
3. **No special characters or emojis** — No repeated punctuation (`!!!`, `...`) or emojis.
4. **No sensitive data** — No patterns like `password:`, `api_key=`, `token:` or string concatenation that may log secrets.

### Supported loggers

- `log/slog`
- `go.uber.org/zap`

## Build and run

### Prerequisites

- Go 1.22+

### Standalone

```bash
# Build
go build -o loglinter .

# Run on current directory
./loglinter ./...

# Run on a specific package
./loglinter ./cmd/...
```

### Install (standalone)

```bash
go install .
# Then run: loglinter ./...
```

## golangci-lint plugin

The linter can be loaded as a **Go plugin** by golangci-lint.

### Build the plugin

From the project root (Linux or macOS; plugin is not supported on Windows):

```bash
CGO_ENABLED=1 go build -buildmode=plugin -o loglint.so ./plugin
```

Requirements:

- `CGO_ENABLED=1`
- Same Go version and environment as the golangci-lint binary you use
- Build on Linux or macOS (Go plugin buildmode is not supported on Windows)

### Configure golangci-lint

Add to `.golangci.yml`:

```yaml
linters:
  enable:
    - loglint  # or add to default set

linters-settings:
  custom:
    loglint:
      path: /path/to/loglint.so   # or absolute path where you built loglint.so
      description: Log message style and safety checks
      original-url: github.com/yourname/loglinter
```

Then run:

```bash
golangci-lint run
```

To enable only this linter:

```bash
golangci-lint run --no-config --default=none -E loglint ./...
```

## Examples

### Incorrect

```go
log.Info("Starting server on port 8080")           // should start with lowercase
slog.Error("ошибка подключения")                   // must be English
slog.Info("server started! 🚀")                    // no emojis
slog.Error("connection failed!!!")                 // no repeated punctuation
log.Info("user password: " + password)             // no sensitive data
```

### Correct

```go
log.Info("starting server on port 8080")
slog.Error("failed to connect to database")
slog.Info("server started")
slog.Error("connection failed")
log.Info("user authenticated successfully")
```

## Tests

```bash
go test ./...
```

## License

MIT
