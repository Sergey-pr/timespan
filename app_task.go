package main

import "time"

// GetTasks returns all tasks sorted by created_at desc.
func (a *App) GetTasks() []Task {
	tasks, err := ListTasks()
	if err != nil {
		a.showError(err)
		return []Task{}
	}
	return tasks
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
	a.syncTicker()
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
	a.syncTicker()
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
	a.syncTicker()
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
	a.syncTicker()
	return true
}
