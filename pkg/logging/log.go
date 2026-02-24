package logging

import (
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/suzuki-shunsuke/slog-error/slogerr"
)

func New(out io.Writer, version string, level *slog.LevelVar) *slog.Logger {
	return slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level: level,
	})).With("version", version)
}

func SetLevel(levelVar *slog.LevelVar, level string) error {
	if level == "" {
		return nil
	}
	lvl, err := ParseLevel(level)
	if err != nil {
		return fmt.Errorf("parse log level: %w", slogerr.With(err, "log_level", level))
	}
	levelVar.Set(lvl)
	return nil
}

// ErrUnknownLogLevel is returned when an invalid log level string is provided to ParseLevel.
var ErrUnknownLogLevel = errors.New("unknown log level")

// ParseLevel converts a string log level to slog.Level.
// Supported levels are: "debug", "info", "warn", "error".
// Returns ErrUnknownLogLevel if the level string is not recognized.
func ParseLevel(lvl string) (slog.Level, error) {
	switch lvl {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, ErrUnknownLogLevel
	}
}
