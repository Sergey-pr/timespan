package main

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"strings"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func migrateDB(db *sql.DB) error {
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version TEXT PRIMARY KEY,
		applied_at DATETIME NOT NULL DEFAULT (datetime('now'))
	)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	for _, entry := range entries {
		version := strings.TrimSuffix(entry.Name(), ".sql")

		var exists bool
		if err = db.QueryRow(`SELECT 1 FROM schema_migrations WHERE version = ?`, version).
			Scan(&exists); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("check migration %s: %w", version, err)
		}
		if exists {
			continue
		}

		data, err := migrationsFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("read migration %s: %w", version, err)
		}

		up := extractUp(string(data))
		if up == "" {
			continue
		}

		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", version, err)
		}
		if _, err := tx.Exec(up); err != nil {
			tx.Rollback()
			return fmt.Errorf("run migration %s: %w", version, err)
		}
		if _, err := tx.Exec(`INSERT INTO schema_migrations (version) VALUES (?)`, version); err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", version, err)
		}
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", version, err)
		}
	}

	return nil
}

// extractUp returns the SQL between "-- migrate:up" and "-- migrate:down" (or EOF).
func extractUp(content string) string {
	start := strings.Index(content, "-- migrate:up")
	if start == -1 {
		return ""
	}
	content = content[start+len("-- migrate:up"):]
	if end := strings.Index(content, "-- migrate:down"); end != -1 {
		content = content[:end]
	}
	return strings.TrimSpace(content)
}
