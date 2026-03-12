// Package main demonstrates correct and incorrect log usage for loglint.
package main

import (
	"log/slog"
)

func main() {
	slog.Info("starting server on port 8080")
	slog.Error("failed to connect to database")
	slog.Debug("request completed")

	// These would be reported by loglint:
	// slog.Info("Starting with capital")
	// slog.Info("запуск сервера")
	// slog.Info("done!!!")
	// slog.Info("password: " + pwd)
}
