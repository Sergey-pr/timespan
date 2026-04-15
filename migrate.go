package main

import (
	"embed"
	"net/url"

	"github.com/amacneil/dbmate/v2/pkg/dbmate"
	_ "github.com/amacneil/dbmate/v2/pkg/driver/sqlite"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func migrateDB(dbPath string) error {
	u, err := url.Parse("sqlite:" + dbPath)
	if err != nil {
		return err
	}
	m := dbmate.New(u)
	m.FS = migrationsFS
	m.MigrationsDir = []string{"migrations"}
	m.AutoDumpSchema = false
	return m.CreateAndMigrate()
}
