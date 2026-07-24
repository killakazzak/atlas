// Package logger provides structured logging for Atlas services.
package logger

import (
	"log/slog"
	"os"
)

// New returns a JSON slog logger writing to stdout.
func New() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}
