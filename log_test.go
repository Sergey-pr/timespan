package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// readLog returns the contents of the log file written into dir.
func readLog(t *testing.T, dir string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(dir, logFileName))
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	return string(data)
}

func TestInitLoggerWritesToFile(t *testing.T) {
	dir := t.TempDir()

	if err := initLogger(dir); err != nil {
		t.Fatalf("initLogger: %v", err)
	}
	slog.Info("hello from the test")

	if got := readLog(t, dir); !strings.Contains(got, "hello from the test") {
		t.Errorf("log file does not contain the message:\n%s", got)
	}
}

func TestInitLoggerAppendsAcrossRuns(t *testing.T) {
	dir := t.TempDir()

	if err := initLogger(dir); err != nil {
		t.Fatalf("first initLogger: %v", err)
	}
	slog.Info("first run")

	if err := initLogger(dir); err != nil {
		t.Fatalf("second initLogger: %v", err)
	}
	slog.Info("second run")

	got := readLog(t, dir)
	if !strings.Contains(got, "first run") {
		t.Errorf("earlier run was truncated away:\n%s", got)
	}
	if !strings.Contains(got, "second run") {
		t.Errorf("second run missing:\n%s", got)
	}
}

func TestInitLoggerSurvivesUnwritableDir(t *testing.T) {
	if err := initLogger(filepath.Join(t.TempDir(), "does-not-exist")); err == nil {
		t.Error("initLogger reported success for an unusable directory")
	}
	// Logging must still work, just without the file.
	slog.Info("stderr only")
}

func TestShowErrorIsLogged(t *testing.T) {
	dir := t.TempDir()
	if err := initLogger(dir); err != nil {
		t.Fatalf("initLogger: %v", err)
	}

	app := newTestApp(t)
	app.CreateTask("   ", "", 0)

	if got := readLog(t, dir); !strings.Contains(got, "task title cannot be empty") {
		t.Errorf("swallowed error never reached the log:\n%s", got)
	}
}
