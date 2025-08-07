package log

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
)

var (
	defaultLogger *slog.Logger
	logLevel      = new(slog.LevelVar)
	logWriter     = os.Stdout
)

func init() {
	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewTextHandler(logWriter, opts)
	defaultLogger = slog.New(handler)
	slog.SetDefault(defaultLogger)
}

// LogTo configures where logs are written and at what level
func LogTo(target string, levelName string) {
	// Configure output target
	switch target {
	case "stdout":
		logWriter = os.Stdout
	case "stderr":
		logWriter = os.Stderr
	case "none":
		// Create a no-op writer
		logWriter = nil
	default:
		// File logging
		file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			panic(fmt.Sprintf("Failed to open log file %s: %v", target, err))
		}
		logWriter = file
	}

	// Configure log level
	switch strings.ToUpper(levelName) {
	case "DEBUG", "FINE", "FINEST", "TRACE":
		logLevel.Set(slog.LevelDebug)
	case "INFO":
		logLevel.Set(slog.LevelInfo)
	case "WARNING", "WARN":
		logLevel.Set(slog.LevelWarn)
	case "ERROR", "CRITICAL":
		logLevel.Set(slog.LevelError)
	default:
		logLevel.Set(slog.LevelInfo)
	}

	// Recreate logger with new settings
	if logWriter != nil {
		opts := &slog.HandlerOptions{
			Level: logLevel,
		}
		handler := slog.NewTextHandler(logWriter, opts)
		defaultLogger = slog.New(handler)
		slog.SetDefault(defaultLogger)
	}
}

// Logger interface for compatibility
type Logger interface {
	AddLogPrefix(string)
	ClearLogPrefixes()
	Debug(string, ...interface{})
	Info(string, ...interface{})
	Warn(string, ...interface{}) error
	Error(string, ...interface{}) error
}

// PrefixLogger adds prefixes to log messages
type PrefixLogger struct {
	logger   *slog.Logger
	prefix   string
	prefixes []string
}

// NewPrefixLogger creates a logger with prefixes
func NewPrefixLogger(prefixes ...string) Logger {
	prefix := strings.Join(prefixes, ".")
	logger := defaultLogger.With("component", prefix)
	return &PrefixLogger{
		logger:   logger,
		prefix:   "[" + prefix + "]",
		prefixes: prefixes,
	}
}

func (pl *PrefixLogger) AddLogPrefix(prefix string) {
	pl.prefixes = append(pl.prefixes, prefix)
	pl.prefix = "[" + strings.Join(pl.prefixes, "] [") + "]"
	pl.logger = defaultLogger.With("component", strings.Join(pl.prefixes, "."))
}

func (pl *PrefixLogger) ClearLogPrefixes() {
	pl.prefixes = nil
	pl.prefix = ""
	pl.logger = defaultLogger
}

func (pl *PrefixLogger) Debug(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	pl.logger.Debug(pl.prefix + " " + msg)
}

func (pl *PrefixLogger) Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	pl.logger.Info(pl.prefix + " " + msg)
}

func (pl *PrefixLogger) Warn(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	pl.logger.Warn(pl.prefix + " " + msg)
	return nil
}

func (pl *PrefixLogger) Error(format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	pl.logger.Error(pl.prefix + " " + msg)
	return nil
}

// Global logging functions
func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(fmt.Sprintf(format, args...))
}

func Info(format string, args ...interface{}) {
	defaultLogger.Info(fmt.Sprintf(format, args...))
}

func Warn(format string, args ...interface{}) error {
	defaultLogger.Warn(fmt.Sprintf(format, args...))
	return nil
}

func Error(format string, args ...interface{}) error {
	defaultLogger.Error(fmt.Sprintf(format, args...))
	return nil
}
