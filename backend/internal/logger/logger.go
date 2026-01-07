package logger

import (
	"context"
	"log/slog"
	"os"
)

// Logger is the global structured logger
var Logger *slog.Logger

// InitLogger initializes the global structured logger
// This is OTel-compatible and outputs JSON for production environments
func InitLogger(environment string) {
	var handler slog.Handler

	if environment == "production" {
		// JSON handler for production (OTel-compatible)
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelInfo,
			AddSource: true,
		})
	} else {
		// Text handler for development (human-readable)
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level:     slog.LevelDebug,
			AddSource: false,
		})
	}

	Logger = slog.New(handler)
	slog.SetDefault(Logger)
}

// Info logs an info level message with optional attributes
func Info(msg string, args ...any) {
	Logger.Info(msg, args...)
}

// InfoCtx logs an info level message with context
func InfoCtx(ctx context.Context, msg string, args ...any) {
	Logger.InfoContext(ctx, msg, args...)
}

// Error logs an error level message with optional attributes
func Error(msg string, args ...any) {
	Logger.Error(msg, args...)
}

// ErrorCtx logs an error level message with context
func ErrorCtx(ctx context.Context, msg string, args ...any) {
	Logger.ErrorContext(ctx, msg, args...)
}

// Debug logs a debug level message with optional attributes
func Debug(msg string, args ...any) {
	Logger.Debug(msg, args...)
}

// DebugCtx logs a debug level message with context
func DebugCtx(ctx context.Context, msg string, args ...any) {
	Logger.DebugContext(ctx, msg, args...)
}

// Warn logs a warning level message with optional attributes
func Warn(msg string, args ...any) {
	Logger.Warn(msg, args...)
}

// WarnCtx logs a warning level message with context
func WarnCtx(ctx context.Context, msg string, args ...any) {
	Logger.WarnContext(ctx, msg, args...)
}

// Fatal logs an error and exits the program
func Fatal(msg string, args ...any) {
	Logger.Error(msg, args...)
	os.Exit(1)
}

// With returns a new logger with the given attributes
func With(args ...any) *slog.Logger {
	return Logger.With(args...)
}
