package main

import (
	"context"
	"testing"
	"time"
)

func TestTickerFollowsTaskTransitions(t *testing.T) {
	app := newTestApp(t)
	id := app.CreateTask("tracked", "", 0).ID

	if app.tickerOn.Load() {
		t.Error("ticker running with no active task")
	}

	app.StartTask(id)
	if !app.tickerOn.Load() {
		t.Error("ticker not running after start")
	}

	app.PauseTask(id)
	if app.tickerOn.Load() {
		t.Error("ticker still running after pause")
	}

	app.StartTask(id)
	app.FinishTask(id)
	if app.tickerOn.Load() {
		t.Error("ticker still running after finish")
	}
}

func TestTickerStopsWhenRunningTaskIsDeleted(t *testing.T) {
	app := newTestApp(t)
	id := app.CreateTask("doomed", "", 0).ID

	app.StartTask(id)
	if !app.tickerOn.Load() {
		t.Fatal("ticker not running after start")
	}

	if !app.DeleteTask(id) {
		t.Fatal("DeleteTask returned false")
	}
	if app.tickerOn.Load() {
		t.Error("ticker still running after the active task was deleted")
	}
}

func TestTickerKeepsRunningWhenAnotherTaskTakesOver(t *testing.T) {
	app := newTestApp(t)
	first := app.CreateTask("first", "", 0).ID
	second := app.CreateTask("second", "", 0).ID

	app.StartTask(first)
	app.StartTask(second)

	if !app.tickerOn.Load() {
		t.Error("ticker stopped even though a task is still running")
	}
}

// The real runTimer must survive transitions and stop on context cancel.
func TestRunTimerRunsAndStops(t *testing.T) {
	app := newTestApp(t)
	id := app.CreateTask("tracked", "", 0).ID

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		app.runTimer(ctx)
		close(done)
	}()

	app.StartTask(id)
	time.Sleep(3 * tickInterval)
	app.PauseTask(id)
	time.Sleep(3 * tickInterval)

	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("runTimer did not return after the context was cancelled")
	}
}
