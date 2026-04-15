// Package nxlog provides a global structured logger built on zap.
// It initializes a production logger at import time and exposes
// level-safe convenience functions. Use Setup to configure the
// level from a config string at startup.
package nxlog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var level = zap.NewAtomicLevelAt(zapcore.InfoLevel)

var logger *zap.Logger

func init() {
	cfg := zap.NewProductionConfig()
	cfg.Level = level
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	logger, _ = cfg.Build(zap.AddCallerSkip(1))
	zap.ReplaceGlobals(logger)
}

// Setup parses a level string (debug, info, warn, error) and
// applies it to the global logger. Safe to call multiple times.
func Setup(lvl string) {
	var l zapcore.Level
	if err := l.UnmarshalText([]byte(lvl)); err != nil {
		logger.Warn("invalid log level, keeping default", zap.String("level", lvl), zap.Error(err))
		return
	}
	level.SetLevel(l)
}

// SetLevel changes the log level at runtime (concurrent-safe).
func SetLevel(l zapcore.Level) {
	level.SetLevel(l)
}

// Logger returns the underlying *zap.Logger for advanced use.
func Logger() *zap.Logger {
	return logger
}

// Sync flushes any buffered log entries. Call on shutdown.
func Sync() {
	_ = logger.Sync()
}

// Info logs at info level.
func Info(msg string, fields ...zap.Field) { logger.Info(msg, fields...) }

// Warn logs at warn level.
func Warn(msg string, fields ...zap.Field) { logger.Warn(msg, fields...) }

// Error logs at error level.
func Error(msg string, fields ...zap.Field) { logger.Error(msg, fields...) }

// Debug logs at debug level.
func Debug(msg string, fields ...zap.Field) { logger.Debug(msg, fields...) }
