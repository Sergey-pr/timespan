package main

import (
	"fmt"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// OpenTimerWindow shows (or creates) the floating timer window for the given task.
func (a *App) OpenTimerWindow(id int64) {
	a.windowsMu.Lock()
	defer a.windowsMu.Unlock()

	if w, ok := a.timerWindows[id]; ok {
		w.Show()
		w.Focus()
		return
	}
	w := application.Get().Window.NewWithOptions(application.WebviewWindowOptions{
		Name:           fmt.Sprintf("timer-%d", id),
		Title:          "Timer",
		Width:          220,
		Height:         100,
		AlwaysOnTop:    true,
		Frameless:      true,
		DisableResize:  true,
		BackgroundType: application.BackgroundTypeTransparent,
		URL:            fmt.Sprintf("/timer.html?taskId=%d", id),
		HideOnEscape:   true,
	})
	// Keeps the map from holding a window the OS or shutdown already destroyed.
	w.OnWindowEvent(events.Common.WindowClosing, func(*application.WindowEvent) {
		a.forgetTimerWindow(id)
	})
	a.timerWindows[id] = w
}

// CloseTimerWindow destroys the window; hiding it would leak the WebView on reopen.
func (a *App) CloseTimerWindow(id int64) {
	a.windowsMu.Lock()
	w, ok := a.timerWindows[id]
	delete(a.timerWindows, id)
	a.windowsMu.Unlock()

	if ok {
		w.Close()
	}
}

func (a *App) forgetTimerWindow(id int64) {
	a.windowsMu.Lock()
	defer a.windowsMu.Unlock()
	delete(a.timerWindows, id)
}

// CloseErrorWindow hides the error window.
func (a *App) CloseErrorWindow() {
	if a.errorWindow != nil {
		a.errorWindow.Hide()
	}
}
