package logger

import (
	"fmt"
	"io"
	"log/slog"
	"strings"
)

type Config struct {
	Level  string
	Format string
}

func New(out io.Writer, cfg Config) (*slog.Logger, error) {
	level, err := parseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}
	opts := &slog.HandlerOptions{Level: level}

	switch strings.ToLower(strings.TrimSpace(cfg.Format)) {
	case "", "json":
		return slog.New(slog.NewJSONHandler(out, opts)), nil
	case "text":
		return slog.New(slog.NewTextHandler(out, opts)), nil
	default:
		return nil, fmt.Errorf("unsupported log format %q", cfg.Format)
	}
}

func parseLevel(value string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "", "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unsupported log level %q", value)
	}
}
