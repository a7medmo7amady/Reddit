package logger

import (
	"log/slog"
	"os"
)

var Log = slog.New(slog.NewJSONHandler(os.Stdout, nil))
