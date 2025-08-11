package log

import (
	"io"
	"log/slog"
)

func New(out io.Writer, version string) *slog.Logger {
	return slog.New(slog.NewJSONHandler(out, nil)).With("version", version)
}
