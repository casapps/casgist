package utils

import (
	"log/slog"
	"os"
	"strings"
)

// NewLogger creates a new logger instance
func NewLogger() *slog.Logger {
	level := slog.LevelInfo
	
	// Check environment for log level
	if lvl := os.Getenv("CASGISTS_LOG_LEVEL"); lvl != "" {
		switch strings.ToLower(lvl) {
		case "debug":
			level = slog.LevelDebug
		case "info":
			level = slog.LevelInfo
		case "warn", "warning":
			level = slog.LevelWarn
		case "error":
			level = slog.LevelError
		}
	}
	
	// Create handler
	opts := &slog.HandlerOptions{
		Level: level,
	}
	
	handler := slog.NewTextHandler(os.Stdout, opts)
	return slog.New(handler)
}