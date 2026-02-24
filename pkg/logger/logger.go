package logger

import (
	"context"
	"log/slog"
	"os"
)

// Config holds logger configuration
type Config struct {
	Level      slog.Level
	Module     string // Module identifier (e.g., "client", "stream")
	AddSource  bool   // Include source file and line number
	JSONFormat bool   // Use JSON format instead of text
}

// NewLogger creates a new structured logger with slog
func NewLogger(config Config) *slog.Logger {
	opts := &slog.HandlerOptions{
		Level:     config.Level,
		AddSource: config.AddSource,
	}

	var handler slog.Handler
	if config.JSONFormat {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	// Create logger with module identifier
	logger := slog.New(handler)

	// Add module identifier to all log entries
	if config.Module != "" {
		logger = logger.With("module", config.Module)
	}

	return logger
}

// Helper functions for context-aware logging with module support

// InfoCtx logs an info message with context
func InfoCtx(ctx context.Context, logger *slog.Logger, msg string, args ...any) {
	logger.InfoContext(ctx, msg, args...)
}

// DebugCtx logs a debug message with context
func DebugCtx(ctx context.Context, logger *slog.Logger, msg string, args ...any) {
	logger.DebugContext(ctx, msg, args...)
}

// WarnCtx logs a warning message with context
func WarnCtx(ctx context.Context, logger *slog.Logger, msg string, args ...any) {
	logger.WarnContext(ctx, msg, args...)
}

// ErrorCtx logs an error message with context
func ErrorCtx(ctx context.Context, logger *slog.Logger, msg string, args ...any) {
	logger.ErrorContext(ctx, msg, args...)
}

// Info logs an info message without context
func Info(logger *slog.Logger, msg string, args ...any) {
	logger.Info(msg, args...)
}

// Debug logs a debug message without context
func Debug(logger *slog.Logger, msg string, args ...any) {
	logger.Debug(msg, args...)
}

// Warn logs a warning message without context
func Warn(logger *slog.Logger, msg string, args ...any) {
	logger.Warn(msg, args...)
}

// Error logs an error message without context
func Error(logger *slog.Logger, msg string, args ...any) {
	logger.Error(msg, args...)
}
