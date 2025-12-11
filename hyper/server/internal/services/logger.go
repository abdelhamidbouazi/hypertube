package services

import (
	"log/slog"
	"os"
	"path/filepath"
)

var Logger = initLogger()

func initLogger() *slog.Logger {
	logsDir := "logs"
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	logFile, err := os.OpenFile(
		filepath.Join(logsDir, "app.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return slog.New(slog.NewTextHandler(os.Stdout, nil))
	}

	return slog.New(slog.NewTextHandler(logFile, nil))
}
