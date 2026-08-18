package main

import (
	"testing"
	"time"
)

// newTestApp gives every test a migrated, empty in-memory database.
func newTestApp(t *testing.T) *App {
	t.Helper()
	if err := initDB(":memory:"); err != nil {
		t.Fatalf("initDB: %v", err)
	}
	return NewApp()
}

// backdateStart rewinds started_at so elapsed accumulation is measurable without sleeping.
func backdateStart(t *testing.T, task *Task, d time.Duration) {
	t.Helper()
	started := time.Now().Add(-d)
	task.StartedAt = &started
	if err := task.Save(); err != nil {
		t.Fatalf("backdate: %v", err)
	}
}

func reload(t *testing.T, id int64) *Task {
	t.Helper()
	task, err := GetTaskByID(id)
	if err != nil {
		t.Fatalf("reload %d: %v", id, err)
	}
	return task
}

func TestCreateTaskStartsReady(t *testing.T) {
	app := newTestApp(t)

	task := app.CreateTask("write tests", "", 0)
	if task == nil {
		t.Fatal("CreateTask returned nil")
	}
	if task.Status != StatusReadyToStart {
		t.Errorf("status = %q, want %q", task.Status, StatusReadyToStart)
	}
	if task.ElapsedMs != 0 || task.StartedAt != nil {
		t.Errorf("new task already has time: elapsed=%d startedAt=%v", task.ElapsedMs, task.StartedAt)
	}
}

func TestPauseAccumulatesElapsed(t *testing.T) {
	app := newTestApp(t)

	task := app.StartTask(app.CreateTask("write tests", "", 0).ID)
	backdateStart(t, task, 2*time.Second)

	paused := app.PauseTask(task.ID)
	if paused.Status != StatusPaused {
		t.Errorf("status = %q, want %q", paused.Status, StatusPaused)
	}
	if paused.StartedAt != nil {
		t.Error("startedAt should be cleared on pause")
	}
	if paused.ElapsedMs < 2000 {
		t.Errorf("elapsed = %dms, want >= 2000", paused.ElapsedMs)
	}
}

func TestResumeKeepsPreviousElapsed(t *testing.T) {
	app := newTestApp(t)

	task := app.StartTask(app.CreateTask("write tests", "", 0).ID)
	backdateStart(t, task, 2*time.Second)
	firstSegment := app.PauseTask(task.ID).ElapsedMs

	resumed := app.StartTask(task.ID)
	backdateStart(t, resumed, 3*time.Second)
	total := app.PauseTask(task.ID).ElapsedMs

	if total < firstSegment+3000 {
		t.Errorf("elapsed = %dms, want >= %dms (both segments)", total, firstSegment+3000)
	}
}

func TestStartPausesPreviouslyRunningTask(t *testing.T) {
	app := newTestApp(t)

	first := app.StartTask(app.CreateTask("first", "", 0).ID)
	backdateStart(t, first, 2*time.Second)

	second := app.StartTask(app.CreateTask("second", "", 0).ID)

	wasRunning := reload(t, first.ID)
	if wasRunning.Status != StatusPaused {
		t.Errorf("first task status = %q, want %q", wasRunning.Status, StatusPaused)
	}
	if wasRunning.StartedAt != nil {
		t.Error("first task should have no startedAt once paused")
	}
	if wasRunning.ElapsedMs < 2000 {
		t.Errorf("first task elapsed = %dms, want >= 2000", wasRunning.ElapsedMs)
	}
	if reload(t, second.ID).Status != StatusActive {
		t.Error("second task should be active")
	}
}

func TestOnlyOneTaskRunsAtATime(t *testing.T) {
	app := newTestApp(t)

	for _, title := range []string{"a", "b", "c"} {
		app.StartTask(app.CreateTask(title, "", 0).ID)
	}

	tasks, err := ListTasks()
	if err != nil {
		t.Fatalf("ListTasks: %v", err)
	}
	active := 0
	for _, task := range tasks {
		if task.Status == StatusActive {
			active++
		}
	}
	if active != 1 {
		t.Errorf("%d active tasks, want 1", active)
	}
}

func TestFinishRecordsFinishedAt(t *testing.T) {
	app := newTestApp(t)

	task := app.StartTask(app.CreateTask("write tests", "", 0).ID)
	backdateStart(t, task, 2*time.Second)

	finished := app.FinishTask(task.ID)
	if finished.Status != StatusFinished {
		t.Errorf("status = %q, want %q", finished.Status, StatusFinished)
	}
	if finished.FinishedAt == nil {
		t.Error("finishedAt should be set")
	}
	if finished.StartedAt != nil {
		t.Error("startedAt should be cleared on finish")
	}
	if finished.ElapsedMs < 2000 {
		t.Errorf("elapsed = %dms, want >= 2000", finished.ElapsedMs)
	}
}

// A task left active by a crash must not bill the time the app was closed.
func TestResetRunningTasksDropsOrphanedSegment(t *testing.T) {
	app := newTestApp(t)

	task := app.StartTask(app.CreateTask("crashed mid-run", "", 0).ID)
	task.ElapsedMs = 5000
	backdateStart(t, task, 48*time.Hour)

	if err := ResetRunningTasks(); err != nil {
		t.Fatalf("ResetRunningTasks: %v", err)
	}

	recovered := reload(t, task.ID)
	if recovered.Status != StatusPaused {
		t.Errorf("status = %q, want %q", recovered.Status, StatusPaused)
	}
	if recovered.StartedAt != nil {
		t.Error("startedAt should be cleared, otherwise offline time keeps accruing")
	}
	if recovered.ElapsedMs != 5000 {
		t.Errorf("elapsed = %dms, want the 5000ms saved before the crash", recovered.ElapsedMs)
	}
}

func TestResetRunningTasksLeavesOtherStatuses(t *testing.T) {
	app := newTestApp(t)

	ready := app.CreateTask("ready", "", 0)
	done := app.FinishTask(app.CreateTask("done", "", 0).ID)

	if err := ResetRunningTasks(); err != nil {
		t.Fatalf("ResetRunningTasks: %v", err)
	}

	if got := reload(t, ready.ID).Status; got != StatusReadyToStart {
		t.Errorf("ready task status = %q, want %q", got, StatusReadyToStart)
	}
	if got := reload(t, done.ID).Status; got != StatusFinished {
		t.Errorf("finished task status = %q, want %q", got, StatusFinished)
	}
}
