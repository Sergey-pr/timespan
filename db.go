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

func initDB() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	dir := filepath.Join(configDir, "TimeSpan")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	dbPath := filepath.Join(dir, "timespan.db")

	sqlDB, err := sql.Open("sqlite", dbPath)
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

func ResetRunningTasks() error {
	_, err := goquDB.Update(taskTable).
		Set(goqu.Record{"status": "paused", "started_at": nil}).
		Where(goqu.C("status").Eq("running")).
		Executor().Exec()
	return err
}
