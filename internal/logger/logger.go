package logger

import (
	"log/slog"
	"os"
)

func New(env string) *slog.Logger {
	var handler slog.Handler

	level := slog.LevelInfo
	if env == "development" {
		level = slog.LevelDebug
	}

	// Always use JSON handler for structured logging
	// This works well with log rotation, Docker logging, and log aggregators
	handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	log := slog.New(handler)
	slog.SetDefault(log)
	return log
}
