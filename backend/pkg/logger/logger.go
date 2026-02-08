package logger

import (
	"iano_chat/pkg/config"
	"io"
	"log/slog"
	"os"

	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLogger(cfg *config.Config) *slog.Logger {
	var handler slog.Handler

	lj := &lumberjack.Logger{
		Filename:   cfg.Log.Path,
		MaxSize:    cfg.Log.MaxSize,
		MaxBackups: cfg.Log.MaxBackups,
		MaxAge:     cfg.Log.MaxAge,
		Compress:   cfg.Log.Compress,
	}

	multiWriter := io.MultiWriter(lj, os.Stdout)

	if cfg.Log.Format == "json" {
		handler = slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
			Level: getLogLevel(cfg.Log.Level),
		})
	} else {
		handler = slog.NewTextHandler(multiWriter, &slog.HandlerOptions{
			Level: getLogLevel(cfg.Log.Level),
		})
	}

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}

func getLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
