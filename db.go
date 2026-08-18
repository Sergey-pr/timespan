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

const dbFileName = "timespan.db"

// configDir returns the app's data directory, creating it if needed.
func configDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	dir := filepath.Join(base, "TimeSpan")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// defaultDSN returns the database path in the user's config dir.
func defaultDSN() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, dbFileName), nil
}

// initDB opens the database at dsn, migrates it and stores it in goquDB.
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
