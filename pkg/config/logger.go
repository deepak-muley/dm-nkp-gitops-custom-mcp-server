package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// LogLevel represents logging levels.
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// Logger provides structured logging to stderr.
type Logger struct {
	level LogLevel
}

// NewLogger creates a new logger with the specified level.
func NewLogger(level string) *Logger {
	var l LogLevel
	switch strings.ToLower(level) {
	case "debug":
		l = DebugLevel
	case "info":
		l = InfoLevel
	case "warn", "warning":
		l = WarnLevel
	case "error":
		l = ErrorLevel
	default:
		l = InfoLevel
	}
	return &Logger{level: l}
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, keysAndValues ...interface{}) {
	if l.level <= DebugLevel {
		l.log("DEBUG", msg, keysAndValues...)
	}
}

// Info logs an info message.
func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	if l.level <= InfoLevel {
		l.log("INFO", msg, keysAndValues...)
	}
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, keysAndValues ...interface{}) {
	if l.level <= WarnLevel {
		l.log("WARN", msg, keysAndValues...)
	}
}

// Error logs an error message.
func (l *Logger) Error(msg string, keysAndValues ...interface{}) {
	if l.level <= ErrorLevel {
		l.log("ERROR", msg, keysAndValues...)
	}
}

// log writes a log message to stderr.
func (l *Logger) log(level, msg string, keysAndValues ...interface{}) {
	timestamp := time.Now().Format("2006-01-02T15:04:05.000Z07:00")

	// Build key-value string
	var kvPairs []string
	for i := 0; i < len(keysAndValues)-1; i += 2 {
		key := fmt.Sprintf("%v", keysAndValues[i])
		value := fmt.Sprintf("%v", keysAndValues[i+1])
		kvPairs = append(kvPairs, fmt.Sprintf("%s=%q", key, value))
	}

	kvStr := ""
	if len(kvPairs) > 0 {
		kvStr = " " + strings.Join(kvPairs, " ")
	}

	fmt.Fprintf(os.Stderr, "%s [%s] %s%s\n", timestamp, level, msg, kvStr)
}
