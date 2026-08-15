package main

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

const logFileName = "timespan.log"

// initLogger points slog at a file in dir, falling back to stderr alone if it cannot be opened.
// The file is deliberately never closed: slog writes are unbuffered, so there is nothing to flush.
func initLogger(dir string) error {
	file, err := os.OpenFile(filepath.Join(dir, logFileName), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
		return err
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(io.MultiWriter(file, os.Stderr), nil)))
	return nil
}
