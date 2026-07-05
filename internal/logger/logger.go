// Package logger provides a structured Zap logger with optional file rotation via lumberjack.
// In development mode logs are emitted to both the console (human-readable) and a rotating
// daily log file under logs/YYYY-MM-DD.log.
// In production mode logs are written only to stdout as JSON for container log aggregation.
package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"devhelp/internal/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/lumberjack.v2"
)

// New constructs and returns a configured *zap.Logger based on the provided Config.
// The caller is responsible for calling logger.Sync() on shutdown.
func New(cfg *config.Config) (*zap.Logger, error) {
	level, err := zapcore.ParseLevel(cfg.Log.Level)
	if err != nil {
		level = zapcore.InfoLevel
	}

	encoderCfg := buildEncoderConfig()

	var cores []zapcore.Core

	if cfg.IsDevelopment() {
		// Human-readable console output for development.
		consoleEncoder := zapcore.NewConsoleEncoder(encoderCfg)
		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), level)
		cores = append(cores, consoleCore)

		// Daily rotating file output for development.
		if cfg.Log.FileEnabled && cfg.Log.FileDir != "" {
			fileCore, fileErr := buildFileCore(cfg.Log.FileDir, encoderCfg, level)
			if fileErr != nil {
				return nil, fileErr
			}
			cores = append(cores, fileCore)
		}
	} else {
		// JSON stdout for production — consumed by container log aggregators.
		jsonEncoder := zapcore.NewJSONEncoder(encoderCfg)
		jsonCore := zapcore.NewCore(jsonEncoder, zapcore.AddSync(os.Stdout), level)
		cores = append(cores, jsonCore)
	}

	core := zapcore.NewTee(cores...)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger, nil
}

// buildEncoderConfig returns a zapcore.EncoderConfig with sensible production defaults.
func buildEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// buildFileCore creates a zapcore.Core that writes JSON logs to a daily rotating file.
func buildFileCore(logDir string, encoderCfg zapcore.EncoderConfig, level zapcore.Level) (zapcore.Core, error) {
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return nil, fmt.Errorf("logger: creating log directory %q: %w", logDir, err)
	}

	filename := filepath.Join(logDir, time.Now().Format("2006-01-02")+".log")

	rotator := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    100, // megabytes
		MaxBackups: 30,
		MaxAge:     90, // days
		Compress:   true,
	}

	// Use JSON encoding for file output to ease log parsing.
	fileCfg := encoderCfg
	fileCfg.EncodeLevel = zapcore.LowercaseLevelEncoder
	fileEncoder := zapcore.NewJSONEncoder(fileCfg)

	return zapcore.NewCore(fileEncoder, zapcore.AddSync(rotator), level), nil
}
