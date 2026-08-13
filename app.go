package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// Binding calls arrive concurrently, so the two pieces of shared state each get a lock:
// taskMu makes a read-modify-write on a task atomic, windowsMu guards the window map.

type App struct {
	taskMu       sync.Mutex
	windowsMu    sync.Mutex
	timerWindows map[int64]*application.WebviewWindow
	errorWindow  *application.WebviewWindow
}

func NewApp() *App {
	return &App{timerWindows: make(map[int64]*application.WebviewWindow)}
}

// SetErrorWindow stores the error window reference (called from main before Run).
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

// emitTaskUpdated pushes a task change to the frontend.
// No-op when no application is running, which is the case in tests.
func emitTaskUpdated(task Task) {
	if app := application.Get(); app != nil {
		app.Event.Emit("task:updated", task)
	}
}

// ServiceStartup is called by the Wails v3 service system when the app starts.
func (a *App) ServiceStartup(ctx context.Context, _ application.ServiceOptions) error {
	dsn, err := defaultDSN()
	if err != nil {
		return err
	}
	if err := initDB(dsn); err != nil {
		return err
	}
	if err := ResetRunningTasks(); err != nil {
		return err
	}
	go a.runTimer(ctx)
	return nil
}

func (a *App) runTimer(ctx context.Context) {
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
	tasks, err := ListTasks()
	if err != nil {
		a.showError(err)
		return []Task{}
	}
	return tasks
}

// ExportReport prompts for a save location and writes an xlsx report of all tasks.
func (a *App) ExportReport() ExportResult {
	path, err := application.Get().Dialog.SaveFile().
		SetFilename(fmt.Sprintf("timespan-report-%s.xlsx", time.Now().Format("2006-01-02"))).
		AddFilter("Excel Workbook", "*.xlsx").
		CanCreateDirectories(true).
		PromptForSingleSelection()
	if err != nil {
		a.showError(err)
		return ExportResult{Status: ExportFailed}
	}
	if path == "" {
		return ExportResult{Status: ExportCancelled}
	}
	if err := buildReport(path); err != nil {
		a.showError(err)
		return ExportResult{Status: ExportFailed}
	}
	return ExportResult{Status: ExportSaved, Path: path}
}

// GetCategories returns all categories sorted alphabetically.
func (a *App) GetCategories() []Category {
	cats, err := ListCategories()
	if err != nil {
		a.showError(err)
		return []Category{}
	}
	return cats
}

// CreateCategory creates a new category with the given name.
func (a *App) CreateCategory(name string) *Category {
	cat := &Category{Name: name}
	if err := cat.Save(); err != nil {
		a.showError(err)
		return nil
	}
	return cat
}

// RenameCategory updates the name of an existing category.
func (a *App) RenameCategory(id int64, name string) *Category {
	cat, err := GetCategoryByID(id)
	if err != nil {
		a.showError(err)
		return nil
	}
	cat.Name = name
	if err := cat.Save(); err != nil {
		a.showError(err)
		return nil
	}
	return cat
}

// DeleteCategory removes a category. Returns false if any tasks still reference it.
func (a *App) DeleteCategory(id int64) bool {
	count, err := CountTasksInCategory(id)
	if err != nil {
		a.showError(err)
		return false
	}
	if count > 0 {
		a.showError(fmt.Errorf("cannot delete category: %d task(s) still use it", count))
		return false
	}
	cat, err := GetCategoryByID(id)
	if err != nil {
		a.showError(err)
		return false
	}
	if err := cat.Delete(); err != nil {
		a.showError(err)
		return false
	}
	return true
}

// CreateTask creates a new pending task. categoryID == 0 means no category.
func (a *App) CreateTask(title string, description string, categoryID int64) *Task {
	var desc *string
	if description != "" {
		desc = &description
	}
	var catID *int64
	if categoryID != 0 {
		catID = &categoryID
	}
	task := &Task{
		Title:       title,
		Description: desc,
		CategoryID:  catID,
		Status:      StatusReadyToStart,
	}
	if err := task.Save(); err != nil {
		a.showError(err)
		return nil
	}
	return task
}

// StartTask pauses any currently running task then starts the given task.
func (a *App) StartTask(id int64) *Task {
	a.taskMu.Lock()
	defer a.taskMu.Unlock()

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
		emitTaskUpdated(*running)
	}

	task, err := GetTaskByID(id)
	if err != nil {
		a.showError(err)
		return nil
	}
	task.Status = StatusActive
	task.StartedAt = &now
	task.FinishedAt = nil
	if err = task.Save(); err != nil {
		a.showError(err)
		return nil
	}
	emitTaskUpdated(*task)
	return task
}

// PauseTask accumulates elapsed time and pauses the task.
func (a *App) PauseTask(id int64) *Task {
	a.taskMu.Lock()
	defer a.taskMu.Unlock()

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
	emitTaskUpdated(*task)
	return task
}

// FinishTask accumulates final elapsed time and marks the task done.
func (a *App) FinishTask(id int64) *Task {
	a.taskMu.Lock()
	defer a.taskMu.Unlock()

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
	emitTaskUpdated(*task)
	return task
}

// EditTask updates the title, description and category of a task. categoryID == 0 clears the category.
func (a *App) EditTask(id int64, title string, description string, categoryID int64) *Task {
	a.taskMu.Lock()
	defer a.taskMu.Unlock()

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
	if categoryID != 0 {
		task.CategoryID = &categoryID
	} else {
		task.CategoryID = nil
	}
	if err = task.Save(); err != nil {
		a.showError(err)
		return nil
	}
	emitTaskUpdated(*task)
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
	// A window destroyed by the OS or at shutdown reports back here, so the map
	// never holds on to a window that is already gone.
	w.OnWindowEvent(events.Common.WindowClosing, func(*application.WindowEvent) {
		a.forgetTimerWindow(id)
	})
	a.timerWindows[id] = w
}

// CloseTimerWindow destroys the floating timer window for the given task.
// Hiding it instead would leak the WebView: reopening always builds a new window.
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
