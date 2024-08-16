// Package migration contains the database initialization handlers
// as well as the schemata that is underlie the migrations.
// Currently only PostgresDB backends are supported.
//
// Note that the migrations are handled via the rubenv/sql-migrate
// library which does not use tagged releases.
package migration

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"go.uber.org/zap"

	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
)

const (
	POSTGRES = "postgres"
)

// ErrNoMigration is returned when the migration with given ID is not found.
var ErrNoMigration = errors.New("migration not found")

// Migrate executes the initalization of a postgres database instance
// with the schemata provided in postgresMigrations. It returns the
// number of performed migrations.
func Migrate(storageType string, migrations map[string]string, migrationTable string, db *sqlx.DB) (int, error) {
	memoryMigrations, err := parseMigration(migrations)
	if err != nil {
		return 0, err
	}

	migrate.SetTable(migrationTable)

	switch storageType {
	case POSTGRES:
		// Postgres uses the postgres driver.
		storageType = "postgres"
	}

	for i := 0; i < 3; i++ {
		var n int
		if n, err = migrate.Exec(db.DB, storageType, memoryMigrations, migrate.Up); err == nil {
			return n, nil
		}
		logger.TechLog.Error(context.Background(), fmt.Sprintf("unable to execute database migrations (retry: %v). Retrying in 10 seconds", i), zap.Error(err))
		time.Sleep(10 * time.Second)
	}
	return 0, fmt.Errorf("unable to execute database migrations: %w", err)
}

func parseMigration(m map[string]string) (*migrate.MemoryMigrationSource, error) {
	var migrations []*migrate.Migration
	for id, migration := range m {
		var m, err = migrate.ParseMigration(id, bytes.NewReader([]byte(migration)))
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, m)
	}

	return &migrate.MemoryMigrationSource{Migrations: migrations}, nil
}
