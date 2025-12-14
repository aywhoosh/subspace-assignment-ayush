package logging

import (
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

type Config struct {
	Level string
	JSON  bool
}

func New(cfg Config) zerolog.Logger {
	level := parseLevel(cfg.Level)

	var out io.Writer = os.Stdout
	if cfg.JSON {
		return zerolog.New(out).
			Level(level).
			With().
			Timestamp().
			Logger()
	}

	cw := zerolog.ConsoleWriter{Out: out, TimeFormat: time.RFC3339}
	return zerolog.New(cw).
		Level(level).
		With().
		Timestamp().
		Logger()
}

func parseLevel(s string) zerolog.Level {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "debug":
		return zerolog.DebugLevel
	case "info", "":
		return zerolog.InfoLevel
	case "warn", "warning":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}
