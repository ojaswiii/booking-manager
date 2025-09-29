package utils

import (
	"os"

	"github.com/sirupsen/logrus"
)

type Logger struct {
	*logrus.Logger
}

// NewLogger creates a new logger instance
func NewLogger() *Logger {
	logger := logrus.New()

	// Set log level based on environment
	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	// Set JSON formatter for production
	if os.Getenv("ENV") == "production" {
		logger.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	return &Logger{Logger: logger}
}

// Info logs an info message with fields
func (l *Logger) Info(msg string, fields ...interface{}) {
	if len(fields) > 0 {
		l.Logger.WithFields(parseFields(fields...)).Info(msg)
	} else {
		l.Logger.Info(msg)
	}
}

// Warn logs a warning message with fields
func (l *Logger) Warn(msg string, fields ...interface{}) {
	if len(fields) > 0 {
		l.Logger.WithFields(parseFields(fields...)).Warn(msg)
	} else {
		l.Logger.Warn(msg)
	}
}

// Error logs an error message with fields
func (l *Logger) Error(msg string, fields ...interface{}) {
	if len(fields) > 0 {
		l.Logger.WithFields(parseFields(fields...)).Error(msg)
	} else {
		l.Logger.Error(msg)
	}
}

// Debug logs a debug message with fields
func (l *Logger) Debug(msg string, fields ...interface{}) {
	if len(fields) > 0 {
		l.Logger.WithFields(parseFields(fields...)).Debug(msg)
	} else {
		l.Logger.Debug(msg)
	}
}

// parseFields converts variadic interface{} to logrus.Fields
func parseFields(fields ...interface{}) logrus.Fields {
	if len(fields)%2 != 0 {
		// If odd number of fields, ignore the last one
		fields = fields[:len(fields)-1]
	}

	result := make(logrus.Fields)
	for i := 0; i < len(fields); i += 2 {
		if key, ok := fields[i].(string); ok {
			result[key] = fields[i+1]
		}
	}
	return result
}
