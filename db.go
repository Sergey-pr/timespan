package main

import (
	"database/sql"
	"os"
	"path/filepath"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"
	_ "modernc.org/sqlite"
)

// goquDB is the shared database connection used by all model methods.
var goquDB *goqu.Database

// defaultDSN returns the database path in the user's config dir, creating the dir.
func defaultDSN() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(configDir, "TimeSpan")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(dir, "timespan.db"), nil
}

// initDB opens the database at dsn, migrates it and stores it in goquDB.
// Tests pass ":memory:" here.
func initDB(dsn string) error {
	sqlDB, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}
	sqlDB.SetMaxOpenConns(1)

	if err := sqlDB.Ping(); err != nil {
		return err
	}

	if err := migrateDB(sqlDB); err != nil {
		return err
	}

	goquDB = goqu.New("sqlite3", sqlDB)
	return nil
}
