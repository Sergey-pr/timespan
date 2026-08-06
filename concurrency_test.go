package main

import (
	"sync"
	"testing"
)

func countActive(t *testing.T) int {
	t.Helper()
	tasks, err := GetTasks()
	if err != nil {
		t.Fatalf("GetTasks: %v", err)
	}
	active := 0
	for _, task := range tasks {
		if task.Status == StatusActive {
			active++
		}
	}
	return active
}

// Two clicks landing in the same instant must not leave two tasks running.
func TestConcurrentStartsLeaveOneActiveTask(t *testing.T) {
	app := newTestApp(t)

	var ids []int64
	for _, title := range []string{"a", "b", "c", "d", "e"} {
		ids = append(ids, app.CreateTask(title, "", 0).ID)
	}

	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			app.StartTask(id)
		}(ids[i%len(ids)])
	}
	wg.Wait()

	if active := countActive(t); active != 1 {
		t.Errorf("%d active tasks after concurrent starts, want 1", active)
	}
}

// Start and pause racing on one task must not lose the accumulated time.
func TestConcurrentTransitionsOnOneTask(t *testing.T) {
	app := newTestApp(t)
	id := app.CreateTask("contended", "", 0).ID

	var wg sync.WaitGroup
	for i := 0; i < 40; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			switch i % 3 {
			case 0:
				app.StartTask(id)
			case 1:
				app.PauseTask(id)
			default:
				app.EditTask(id, "contended", "", 0)
			}
		}(i)
	}
	wg.Wait()

	task := reload(t, id)
	if task.Title != "contended" {
		t.Errorf("title = %q, want %q", task.Title, "contended")
	}
	if task.ElapsedMs < 0 {
		t.Errorf("elapsed = %dms, want >= 0", task.ElapsedMs)
	}
}
