package main

import (
	"database/sql"
	"io/fs"
	"testing"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = db.Close()
	})
	return db
}

func countMigrationFiles(t *testing.T) int {
	t.Helper()
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		t.Fatalf("read migrations: %v", err)
	}
	return len(entries)
}

func TestMigrateAppliesEveryMigration(t *testing.T) {
	db := newTestDB(t)

	if err := migrateDB(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	var applied int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&applied); err != nil {
		t.Fatalf("count: %v", err)
	}
	if want := countMigrationFiles(t); applied != want {
		t.Errorf("%d migrations recorded, want %d", applied, want)
	}

	for _, table := range []string{"tasks", "categories"} {
		var name string
		err := db.QueryRow(`SELECT name FROM sqlite_master WHERE type='table' AND name=?`, table).Scan(&name)
		if err != nil {
			t.Errorf("table %s missing after migrate: %v", table, err)
		}
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	db := newTestDB(t)

	if err := migrateDB(db); err != nil {
		t.Fatalf("first migrate: %v", err)
	}
	if err := migrateDB(db); err != nil {
		t.Fatalf("second migrate: %v", err)
	}

	var applied int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&applied); err != nil {
		t.Fatalf("count: %v", err)
	}
	if want := countMigrationFiles(t); applied != want {
		t.Errorf("%d migrations recorded after two runs, want %d", applied, want)
	}
}

// The pending→ready_to_start repair migration is the one that has to keep working
// on databases written by older builds.
func TestMigrateNormalisesLegacyPendingStatus(t *testing.T) {
	db := newTestDB(t)

	if _, err := db.Exec(`CREATE TABLE tasks (
		id INTEGER PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT,
		status TEXT NOT NULL DEFAULT 'pending',
		elapsed_ms INTEGER NOT NULL DEFAULT 0,
		started_at DATETIME,
		created_at DATETIME NOT NULL,
		finished_at DATETIME
	)`); err != nil {
		t.Fatalf("legacy schema: %v", err)
	}
	if _, err := db.Exec(
		`INSERT INTO tasks (title, status, created_at) VALUES ('legacy', 'pending', datetime('now'))`,
	); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if err := migrateDB(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	var status string
	if err := db.QueryRow(`SELECT status FROM tasks WHERE title = 'legacy'`).Scan(&status); err != nil {
		t.Fatalf("read back: %v", err)
	}
	if status != string(StatusReadyToStart) {
		t.Errorf("status = %q, want %q", status, StatusReadyToStart)
	}
}

func TestExtractUp(t *testing.T) {
	cases := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "up and down",
			content: "-- migrate:up\nCREATE TABLE t (id INTEGER);\n\n-- migrate:down\nDROP TABLE t;",
			want:    "CREATE TABLE t (id INTEGER);",
		},
		{
			name:    "up only",
			content: "-- migrate:up\nCREATE TABLE t (id INTEGER);",
			want:    "CREATE TABLE t (id INTEGER);",
		},
		{
			name:    "no up section",
			content: "DROP TABLE t;",
			want:    "",
		},
		{
			name:    "empty up section",
			content: "-- migrate:up\n\n-- migrate:down\nDROP TABLE t;",
			want:    "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := extractUp(tc.content); got != tc.want {
				t.Errorf("extractUp() = %q, want %q", got, tc.want)
			}
		})
	}
}
