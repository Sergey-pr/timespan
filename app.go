package main

import (
	"context"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type TaskStatus string

const (
	StatusPending  TaskStatus = "pending"
	StatusRunning  TaskStatus = "running"
	StatusPaused   TaskStatus = "paused"
	StatusFinished TaskStatus = "finished"
)

type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	Status      TaskStatus `json:"status"`
	ElapsedMs   int64      `json:"elapsedMs"`
	StartedAt   *time.Time `json:"startedAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	FinishedAt  *time.Time `json:"finishedAt"`
}

type App struct {
	db          *DB
	timerWindow *application.WebviewWindow
}

func NewApp() *App {
	return &App{}
}

// SetTimerWindow stores the floating timer window reference (called from main before Run).
//
//wails:internal
func (a *App) SetTimerWindow(w *application.WebviewWindow) {
	a.timerWindow = w
}

// ServiceStartup is called by the Wails v3 service system when the app starts.
func (a *App) ServiceStartup(ctx context.Context, _ application.ServiceOptions) error {
	db, err := NewDB()
	if err != nil {
		return err
	}
	a.db = db

	if err := a.db.ResetRunningTasks(); err != nil {
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
	tasks, err := a.db.GetTasks()
	if err != nil {
		return []Task{}
	}
	return tasks
}

// CreateTask creates a new pending task.
func (a *App) CreateTask(title string, description string) Task {
	now := time.Now()
	var desc *string
	if description != "" {
		desc = &description
	}
	task := Task{
		ID:          generateID(),
		Title:       title,
		Description: desc,
		Status:      StatusPending,
		ElapsedMs:   0,
		CreatedAt:   now,
	}
	_ = a.db.InsertTask(task)
	return task
}

// StartTask pauses any currently running task then starts the given task.
func (a *App) StartTask(id string) *Task {
	tasks, err := a.db.GetTasks()
	if err != nil {
		return nil
	}
	now := time.Now()

	for i := range tasks {
		if tasks[i].Status == StatusRunning && tasks[i].ID != id {
			if tasks[i].StartedAt != nil {
				tasks[i].ElapsedMs += now.Sub(*tasks[i].StartedAt).Milliseconds()
			}
			tasks[i].Status = StatusPaused
			tasks[i].StartedAt = nil
			_ = a.db.UpdateTask(tasks[i])
			application.Get().Event.Emit("task:updated", tasks[i])
		}
	}

	for i := range tasks {
		if tasks[i].ID == id {
			tasks[i].Status = StatusRunning
			tasks[i].StartedAt = &now
			tasks[i].FinishedAt = nil
			_ = a.db.UpdateTask(tasks[i])
			application.Get().Event.Emit("task:updated", tasks[i])
			return &tasks[i]
		}
	}
	return nil
}

// EditTask updates the title and description of a task.
func (a *App) EditTask(id string, title string, description string) *Task {
	tasks, err := a.db.GetTasks()
	if err != nil {
		return nil
	}
	for i := range tasks {
		if tasks[i].ID == id {
			tasks[i].Title = title
			if description != "" {
				tasks[i].Description = &description
			} else {
				tasks[i].Description = nil
			}
			_ = a.db.UpdateTask(tasks[i])
			application.Get().Event.Emit("task:updated", tasks[i])
			return &tasks[i]
		}
	}
	return nil
}

// PauseTask accumulates elapsed time and pauses the task.
func (a *App) PauseTask(id string) *Task {
	tasks, err := a.db.GetTasks()
	if err != nil {
		return nil
	}
	now := time.Now()
	for i := range tasks {
		if tasks[i].ID == id {
			if tasks[i].StartedAt != nil {
				tasks[i].ElapsedMs += now.Sub(*tasks[i].StartedAt).Milliseconds()
			}
			tasks[i].Status = StatusPaused
			tasks[i].StartedAt = nil
			_ = a.db.UpdateTask(tasks[i])
			application.Get().Event.Emit("task:updated", tasks[i])
			return &tasks[i]
		}
	}
	return nil
}

// FinishTask accumulates final elapsed time and marks the task done.
func (a *App) FinishTask(id string) *Task {
	tasks, err := a.db.GetTasks()
	if err != nil {
		return nil
	}
	now := time.Now()
	for i := range tasks {
		if tasks[i].ID == id {
			if tasks[i].StartedAt != nil {
				tasks[i].ElapsedMs += now.Sub(*tasks[i].StartedAt).Milliseconds()
			}
			tasks[i].Status = StatusFinished
			tasks[i].StartedAt = nil
			tasks[i].FinishedAt = &now
			_ = a.db.UpdateTask(tasks[i])
			application.Get().Event.Emit("task:updated", tasks[i])
			return &tasks[i]
		}
	}
	return nil
}

// DeleteTask removes a task by id.
func (a *App) DeleteTask(id string) bool {
	return a.db.DeleteTask(id) == nil
}

// GetCurrentElapsed returns live elapsed ms including the active running segment.
func (a *App) GetCurrentElapsed(id string) int64 {
	tasks, err := a.db.GetTasks()
	if err != nil {
		return 0
	}
	for _, t := range tasks {
		if t.ID == id {
			if t.Status == StatusRunning && t.StartedAt != nil {
				return t.ElapsedMs + time.Since(*t.StartedAt).Milliseconds()
			}
			return t.ElapsedMs
		}
	}
	return 0
}

// OpenTimerWindow shows the floating timer OS window for the given task.
func (a *App) OpenTimerWindow(id string) {
	if a.timerWindow == nil {
		return
	}
	// Emit the task id so the timer window knows what to display.
	application.Get().Event.Emit("timer:open", id)
	a.timerWindow.Show()
	a.timerWindow.SetAlwaysOnTop(true)
}

// CloseTimerWindow hides the floating timer OS window.
func (a *App) CloseTimerWindow() {
	if a.timerWindow != nil {
		a.timerWindow.Hide()
	}
}
