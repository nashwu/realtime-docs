package app

import (
	"log/slog"
	"os"
)

// NewLogger returns a slog.Logger with formatting + level based on env
// prod JSON logs at INFO level
// others Text logs at DEBUG level
func NewLogger(env string) *slog.Logger {
	var handler slog.Handler
	if env == "prod" {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	}
	return slog.New(handler)
}
