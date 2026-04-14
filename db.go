package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
	_ "modernc.org/sqlite"
)

type DB struct {
	db      *sql.DB
	dialect goqu.DialectWrapper
}

func NewDB() (*DB, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(configDir, "TimeSpan")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(dir, "timespan.db")
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	d := &DB{
		db:      db,
		dialect: goqu.Dialect("sqlite3"),
	}

	if err := d.migrate(); err != nil {
		return nil, err
	}

	return d, nil
}

func (d *DB) migrate() error {
	_, err := d.db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id          TEXT PRIMARY KEY,
			title       TEXT NOT NULL,
			description TEXT,
			status      TEXT NOT NULL DEFAULT 'pending',
			elapsed_ms  INTEGER NOT NULL DEFAULT 0,
			started_at  DATETIME,
			created_at  DATETIME NOT NULL,
			finished_at DATETIME
		)
	`)
	return err
}

const selectAllTasksSQL = `
	SELECT id, title, description, status, elapsed_ms, started_at, created_at, finished_at
	FROM tasks
	ORDER BY created_at DESC
`

func scanTask(rows *sql.Rows) (Task, error) {
	var t Task
	var description sql.NullString
	var startedAt, finishedAt sql.NullString
	var createdAtStr string

	err := rows.Scan(
		&t.ID, &t.Title, &description, &t.Status, &t.ElapsedMs,
		&startedAt, &createdAtStr, &finishedAt,
	)
	if err != nil {
		return t, err
	}

	if description.Valid {
		t.Description = &description.String
	}

	t.CreatedAt, _ = parseTimeStr(createdAtStr)

	if startedAt.Valid {
		ts, err := parseTimeStr(startedAt.String)
		if err == nil {
			t.StartedAt = &ts
		}
	}
	if finishedAt.Valid {
		ts, err := parseTimeStr(finishedAt.String)
		if err == nil {
			t.FinishedAt = &ts
		}
	}

	return t, nil
}

func parseTimeStr(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05.999999999",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, sql.ErrNoRows
}

func (d *DB) GetTasks() ([]Task, error) {
	rows, err := d.db.Query(selectAllTasksSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	if tasks == nil {
		tasks = []Task{}
	}
	return tasks, rows.Err()
}

func (d *DB) InsertTask(task Task) error {
	query, args, err := d.dialect.Insert("tasks").
		Rows(goqu.Record{
			"id":          task.ID,
			"title":       task.Title,
			"description": task.Description,
			"status":      string(task.Status),
			"elapsed_ms":  task.ElapsedMs,
			"started_at":  task.StartedAt,
			"created_at":  task.CreatedAt.Format(time.RFC3339Nano),
			"finished_at": task.FinishedAt,
		}).
		Prepared(true).
		ToSQL()
	if err != nil {
		return err
	}
	_, err = d.db.Exec(query, args...)
	return err
}

func (d *DB) UpdateTask(task Task) error {
	record := goqu.Record{
		"title":       task.Title,
		"description": task.Description,
		"status":      string(task.Status),
		"elapsed_ms":  task.ElapsedMs,
		"finished_at": task.FinishedAt,
	}
	if task.StartedAt != nil {
		record["started_at"] = task.StartedAt.Format(time.RFC3339Nano)
	} else {
		record["started_at"] = nil
	}
	if task.FinishedAt != nil {
		record["finished_at"] = task.FinishedAt.Format(time.RFC3339Nano)
	} else {
		record["finished_at"] = nil
	}

	query, args, err := d.dialect.Update("tasks").
		Set(record).
		Where(goqu.C("id").Eq(task.ID)).
		Prepared(true).
		ToSQL()
	if err != nil {
		return err
	}
	_, err = d.db.Exec(query, args...)
	return err
}

func (d *DB) DeleteTask(id string) error {
	query, args, err := d.dialect.Delete("tasks").
		Where(goqu.C("id").Eq(id)).
		Prepared(true).
		ToSQL()
	if err != nil {
		return err
	}
	_, err = d.db.Exec(query, args...)
	return err
}

func (d *DB) ResetRunningTasks() error {
	query, args, err := d.dialect.Update("tasks").
		Set(goqu.Record{
			"status":     "paused",
			"started_at": nil,
		}).
		Where(goqu.C("status").Eq("running")).
		Prepared(true).
		ToSQL()
	if err != nil {
		return err
	}
	_, err = d.db.Exec(query, args...)
	return err
}
