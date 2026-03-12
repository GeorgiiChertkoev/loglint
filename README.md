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

## CI/CD

The repository includes GitHub Actions workflows for continuous integration and releases.

### CI (`ci.yml`)

Runs automatically on every push to `main`, on pull requests, and on manual dispatch. Three jobs:

| Job | Runs on | What it does |
|-----|---------|-------------|
| **test** | Ubuntu, macOS, Windows | `go build .` and `go test ./...` |
| **lint** | Ubuntu | `go vet ./...` |
| **plugin-smoke** | Ubuntu | Builds the golangci-lint plugin with `-buildmode=plugin` and verifies the artifact |

The plugin smoke test runs only on Linux because Go's `-buildmode=plugin` is not supported on Windows and requires `CGO_ENABLED=1`.

### Release (`release.yml`)

Triggered when a version tag is pushed (e.g. `v0.1.0`). It builds standalone CLI binaries for five targets, packages them, and publishes a GitHub Release with checksums:

- `loglint-linux-amd64.tar.gz`
- `loglint-linux-arm64.tar.gz`
- `loglint-darwin-amd64.tar.gz`
- `loglint-darwin-arm64.tar.gz`
- `loglint-windows-amd64.zip`
- `checksums.txt`

To create a release:

```bash
git tag v0.1.0
git push origin v0.1.0
```

The plugin `.so` is **not** included in releases because it must be built in the same Go/OS environment as your `golangci-lint` binary. Build it locally instead (see [golangci-lint plugin](#golangci-lint-plugin) above).

### Using loglint in your own GitHub Actions

**Standalone** — install and run in any workflow:

```yaml
- uses: actions/setup-go@v5
  with:
    go-version: "1.22"

- run: go install github.com/GeorgiiChertkoev/loglint@latest

- run: loglint ./...
```

**Plugin mode** — build the plugin on Linux/macOS, then run golangci-lint:

```yaml
- run: |
    git clone https://github.com/GeorgiiChertkoev/loglint.git /tmp/loglint
    cd /tmp/loglint
    CGO_ENABLED=1 go build -buildmode=plugin -o loglint.so ./plugin
    cp loglint.so $GITHUB_WORKSPACE/

- run: golangci-lint run
```

Make sure your `.golangci.yml` points `path:` at `./loglint.so`.

### Troubleshooting

- **Plugin build fails on Windows** — Go's `-buildmode=plugin` is not supported on Windows. Use Linux or macOS runners.
- **Plugin load error / version mismatch** — The plugin must be compiled with the same Go version and OS/arch as the `golangci-lint` binary. Pin both in your workflow.

## License

MIT
