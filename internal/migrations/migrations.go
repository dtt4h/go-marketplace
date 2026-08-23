package migrations

import (
	"embed"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

//go:embed *.sql
var migrationFS embed.FS

// RunMigrations runs all pending up migrations from embedded SQL files.
func RunMigrations(dsn string) error {
	driver, err := iofs.New(migrationFS, ".")
	if err != nil {
		log.Fatalf("create iofs driver: %v", err)
	}

	m, err := migrate.NewWithSourceInstance("embedded", driver, dsn)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}