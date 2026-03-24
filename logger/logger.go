package logger

import (
	"log/slog"
	"os"
	"strings"
)

var L *slog.Logger

func levelFromEnv() slog.Level {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("LOG_LEVEL"))) {
	case "debug":
		return slog.LevelDebug
	case "info", "":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func Init() {
	L = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: levelFromEnv(),
	}))
	slog.SetDefault(L)
}
