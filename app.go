package main

import (
	"context"
	"log/slog"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type App struct {
	// taskMu makes a read-modify-write on a task atomic; windowsMu guards the window map.
	taskMu       sync.Mutex
	windowsMu    sync.Mutex
	timerWindows map[int64]*application.WebviewWindow
	errorWindow  *application.WebviewWindow
	tickerOn     atomic.Bool
	tickerWake   chan struct{}
}

func NewApp() *App {
	return &App{
		timerWindows: make(map[int64]*application.WebviewWindow),
		tickerWake:   make(chan struct{}, 1),
	}
}

// SetErrorWindow stores the error window reference; only main.go calls it.
//
//wails:ignore
func (a *App) SetErrorWindow(w *application.WebviewWindow) {
	a.errorWindow = w
}

// showError logs first, so a failure is recorded even with no window to show it in.
func (a *App) showError(err error) {
	if err == nil {
		return
	}
	slog.Error("backend error", "error", err)
	if a.errorWindow == nil {
		return
	}
	application.Get().Event.Emit("app:error", err.Error())
	a.errorWindow.Show()
}

// emitTaskUpdated pushes a task change to the frontend, no-op without a running app.
func emitTaskUpdated(task Task) {
	if app := application.Get(); app != nil {
		app.Event.Emit("task:updated", task)
	}
}

// emitTick tells the frontend to advance elapsed time for the running task.
func emitTick() {
	if app := application.Get(); app != nil {
		app.Event.Emit("tick")
	}
}

// ServiceStartup is called by the Wails v3 service system when the app starts.
func (a *App) ServiceStartup(ctx context.Context, _ application.ServiceOptions) error {
	dir, err := configDir()
	if err != nil {
		return err
	}
	if err := initLogger(dir); err != nil {
		slog.Warn("file logging unavailable", "error", err)
	}
	slog.Info("starting", "dir", dir)

	if err := initDB(filepath.Join(dir, dbFileName)); err != nil {
		slog.Error("database init failed", "error", err)
		return err
	}
	if err := ResetRunningTasks(); err != nil {
		slog.Error("crash recovery failed", "error", err)
		return err
	}
	a.syncTicker()
	go a.runTimer(ctx)
	return nil
}
