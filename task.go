package main

import (
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
)

type TaskStatus string

const (
	StatusPending  TaskStatus = "pending"
	StatusRunning  TaskStatus = "running"
	StatusPaused   TaskStatus = "paused"
	StatusFinished TaskStatus = "finished"
)

var taskTable = goqu.T("tasks")

type Task struct {
	ID          int64      `db:"id"          json:"id"          goqu:"skipinsert,skipupdate"`
	Title       string     `db:"title"       json:"title"`
	Description *string    `db:"description" json:"description,omitempty"`
	Status      TaskStatus `db:"status"      json:"status"`
	ElapsedMs   int64      `db:"elapsed_ms"  json:"elapsedMs"`
	StartedAt   *time.Time `db:"started_at"  json:"startedAt"`
	CreatedAt   time.Time  `db:"created_at"  json:"createdAt"  goqu:"skipupdate"`
	FinishedAt  *time.Time `db:"finished_at" json:"finishedAt,omitempty"`
}

// GetTasks returns all tasks sorted by created_at desc.
func GetTasks() ([]Task, error) {
	var tasks []Task
	err := goquDB.From(taskTable).
		Order(goqu.I("created_at").Desc()).
		ScanStructs(&tasks)
	if tasks == nil {
		tasks = []Task{}
	}
	return tasks, err
}

// GetTaskByID returns the task with the given id, or an error if not found.
func GetTaskByID(id int64) (*Task, error) {
	var task Task
	found, err := goquDB.From(taskTable).
		Where(goqu.C("id").Eq(id)).
		ScanStruct(&task)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("task %d not found", id)
	}
	return &task, nil
}

// GetRunningTask returns the currently running task, or nil if none is running.
func GetRunningTask() (*Task, error) {
	var task Task
	found, err := goquDB.From(taskTable).
		Where(goqu.C("status").Eq(StatusRunning)).
		ScanStruct(&task)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return &task, nil
}

// Save inserts or updates the task. ID == 0 means a new record.
func (t *Task) Save() error {
	if t.ID == 0 {
		t.CreatedAt = time.Now()
		result, err := goquDB.Insert(taskTable).Rows(t).Executor().Exec()
		if err != nil {
			return err
		}
		t.ID, err = result.LastInsertId()
		return err
	}
	_, err := goquDB.Update(taskTable).
		Set(t).
		Where(goqu.C("id").Eq(t.ID)).
		Executor().Exec()
	return err
}

// Delete removes the task from the database.
func (t *Task) Delete() error {
	_, err := goquDB.Delete(taskTable).
		Where(goqu.C("id").Eq(t.ID)).
		Executor().Exec()
	return err
}
