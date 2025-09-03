// Package logger provides structured logging functionality using Uber's Zap library.
package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger to provide a convenient interface
type Logger struct {
	*zap.Logger
}

// New creates a new Logger instance
func New() (*Logger, error) {
	// Check if we're in development mode
	if os.Getenv("ENV") == "development" {
		return newDevelopmentLogger()
	}
	return newProductionLogger()
}

// newDevelopmentLogger creates a logger for development environment
func newDevelopmentLogger() (*Logger, error) {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	return &Logger{logger}, nil
}

// newProductionLogger creates a logger for production environment
func newProductionLogger() (*Logger, error) {
	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, err := config.Build()
	if err != nil {
		return nil, err
	}
	return &Logger{logger}, nil
}

// WithFields adds structured fields to the logger
func (l *Logger) WithFields(fields map[string]any) *Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		zapFields = append(zapFields, zap.Any(key, value))
	}
	return &Logger{l.With(zapFields...)}
}

// Close flushes any buffered log entries
func (l *Logger) Close() error {
	return l.Sync()
}
