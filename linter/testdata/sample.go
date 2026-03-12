package a

import (
	"log"
	"log/slog"

	"go.uber.org/zap"
)

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Rule 1: lowercase - should report
	slog.Info("Starting server") // want `log message must start with a lowercase letter`

	// Rule 1: OK
	slog.Info("starting server")

	// Rule 2: Russian - should report
	slog.Error("ошибка подключения") // want `log message must be in English only`

	// Rule 2: OK
	slog.Info("failed to connect")

	// Rule 3: emoji - reports both English and emoji diagnostics
	slog.Info("server started \U0001F680") // want `log message must be in English only` `log message must not contain emojis`

	// Rule 3: repeated punctuation - should report
	slog.Error("connection failed!!!") // want `log message must not contain repeated punctuation`

	// Rule 3: OK
	slog.Warn("something went wrong")

	// Rule 4: sensitive literal - should report
	slog.Info("user password: xxx") // want `log message must not contain potentially sensitive`

	// Rule 4: literal-only concat with sensitive pattern match - should report
	slog.Debug("api_key=" + "x") // want `log message must not contain potentially sensitive`

	// Rule 4: benign literal-only concat - should NOT report
	slog.Debug("not found: " + "api_key")

	// OK via zap
	logger.Info("request completed successfully")

	// stdlib log support
	log.Print("Starting stdlib") // want `log message must start with a lowercase letter`
	log.Println("starting stdlib ok")
	log.Printf("token: leaked", "x") // want `log message must not contain potentially sensitive`
}
