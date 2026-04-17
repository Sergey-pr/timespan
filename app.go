package main

import (
	"context"
	"fmt"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type App struct {
	timerWindows map[int64]*application.WebviewWindow
	errorWindow  *application.WebviewWindow
}

func NewApp() *App {
	return &App{timerWindows: make(map[int64]*application.WebviewWindow)}
}

// SetErrorWindow stores the error window reference (called from main before Run).
//
//wails:internal
func (a *App) SetErrorWindow(w *application.WebviewWindow) {
	a.errorWindow = w
}

func (a *App) showError(err error) {
	if err == nil || a.errorWindow == nil {
		return
	}
	application.Get().Event.Emit("app:error", err.Error())
	a.errorWindow.Show()
}

// ServiceStartup is called by the Wails v3 service system when the app starts.
func (a *App) ServiceStartup(ctx context.Context, _ application.ServiceOptions) error {
	if err := initDB(); err != nil {
		return err
	}
	if err := ResetRunningTasks(); err != nil {
		return err
	}
	go a.runTicker(ctx)
	return nil
}

func (a *App) runTicker(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			application.Get().Event.Emit("tick")
		case <-ctx.Done():
			return
		}
	}
}

// GetTasks returns all tasks sorted by created_at desc.
func (a *App) GetTasks() []Task {
	tasks, err := GetTasks()
	if err != nil {
		a.showError(err)
		return []Task{}
	}
	return tasks
}

// CreateTask creates a new pending task.
func (a *App) CreateTask(title string, description string) *Task {
	var desc *string
	if description != "" {
		desc = &description
	}
	task := &Task{
		Title:       title,
		Description: desc,
		Status:      StatusPending,
	}
	if err := task.Save(); err != nil {
		a.showError(err)
		return nil
	}
	return task
}

// StartTask pauses any currently running task then starts the given task.
func (a *App) StartTask(id int64) *Task {
	now := time.Now()

	running, err := GetRunningTask()
	if err != nil {
		a.showError(err)
		return nil
	}
	if running != nil && running.ID != id {
		if running.StartedAt != nil {
			running.ElapsedMs += now.Sub(*running.StartedAt).Milliseconds()
		}
		running.Status = StatusPaused
		running.StartedAt = nil
		if err := running.Save(); err != nil {
			a.showError(err)
			return nil
		}
		application.Get().Event.Emit("task:updated", *running)
	}

	task, err := GetTaskByID(id)
	if err != nil {
		a.showError(err)
		return nil
	}
	task.Status = StatusRunning
	task.StartedAt = &now
	task.FinishedAt = nil
	if err = task.Save(); err != nil {
		a.showError(err)
		return nil
	}
	application.Get().Event.Emit("task:updated", *task)
	return task
}

// PauseTask accumulates elapsed time and pauses the task.
func (a *App) PauseTask(id int64) *Task {
	task, err := GetTaskByID(id)
	if err != nil {
		a.showError(err)
		return nil
	}
	now := time.Now()
	if task.StartedAt != nil {
		task.ElapsedMs += now.Sub(*task.StartedAt).Milliseconds()
	}
	task.Status = StatusPaused
	task.StartedAt = nil
	if err = task.Save(); err != nil {
		a.showError(err)
		return nil
	}
	application.Get().Event.Emit("task:updated", *task)
	return task
}

// FinishTask accumulates final elapsed time and marks the task done.
func (a *App) FinishTask(id int64) *Task {
	task, err := GetTaskByID(id)
	if err != nil {
		a.showError(err)
		return nil
	}
	now := time.Now()
	if task.StartedAt != nil {
		task.ElapsedMs += now.Sub(*task.StartedAt).Milliseconds()
	}
	task.Status = StatusFinished
	task.StartedAt = nil
	task.FinishedAt = &now
	if err = task.Save(); err != nil {
		a.showError(err)
		return nil
	}
	application.Get().Event.Emit("task:updated", *task)
	return task
}

// EditTask updates the title and description of a task.
func (a *App) EditTask(id int64, title string, description string) *Task {
	task, err := GetTaskByID(id)
	if err != nil {
		a.showError(err)
		return nil
	}
	task.Title = title
	if description != "" {
		task.Description = &description
	} else {
		task.Description = nil
	}
	if err = task.Save(); err != nil {
		a.showError(err)
		return nil
	}
	application.Get().Event.Emit("task:updated", *task)
	return task
}

// DeleteTask removes a task by id.
func (a *App) DeleteTask(id int64) bool {
	task, err := GetTaskByID(id)
	if err != nil {
		a.showError(err)
		return false
	}
	if err = task.Delete(); err != nil {
		a.showError(err)
		return false
	}
	return true
}

// GetCurrentElapsed returns live elapsed ms including the active running segment.
func (a *App) GetCurrentElapsed(id int64) int64 {
	task, err := GetTaskByID(id)
	if err != nil {
		a.showError(err)
		return 0
	}
	if task.Status == StatusRunning && task.StartedAt != nil {
		return task.ElapsedMs + time.Since(*task.StartedAt).Milliseconds()
	}
	return task.ElapsedMs
}

// OpenTimerWindow shows (or creates) the floating timer window for the given task.
func (a *App) OpenTimerWindow(id int64) {
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
	a.timerWindows[id] = w
}

// CloseTimerWindow hides the floating timer window for the given task.
func (a *App) CloseTimerWindow(id int64) {
	if w, ok := a.timerWindows[id]; ok {
		w.Hide()
		delete(a.timerWindows, id)
	}
}

// CloseErrorWindow hides the error window.
func (a *App) CloseErrorWindow() {
	if a.errorWindow != nil {
		a.errorWindow.Hide()
	}
}
