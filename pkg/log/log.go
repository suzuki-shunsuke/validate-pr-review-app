package log

import (
	"errors"
	"io"
	"log/slog"
)

func New(out io.Writer, version string, level *slog.LevelVar) *slog.Logger {
	return slog.New(slog.NewJSONHandler(out, &slog.HandlerOptions{
		Level: level,
	})).With("version", version)
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
