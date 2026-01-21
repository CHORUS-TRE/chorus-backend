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
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"go.uber.org/zap"

	"github.com/jmoiron/sqlx"
	migrate "github.com/rubenv/sql-migrate"
)

const (
	POSTGRES = "postgres"
)

func readFile(migrationFS embed.FS, file string) (string, error) {
	r, err := migrationFS.Open(file)
	if err != nil {
		return "", err
	}
	defer r.Close()

	contents, err := io.ReadAll(r)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}

func filePath(storageType, fileName string) string {
	return fmt.Sprintf("%s/%s", storageType, fileName)
}

func removeFileExtension(f string) string {
	extension := filepath.Ext(f)
	return strings.TrimSuffix(f, extension)
}

func listMigrationFiles(migrationFS embed.FS, path string) ([]string, error) {
	files := []string{}

	err := fs.WalkDir(migrationFS, path, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		if !info.IsDir() {
			files = append(files, info.Name())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}

// ErrNoMigration is returned when the migration with given ID is not found.
var ErrNoMigration = errors.New("migration not found")

// Migrate executes the initalization of a postgres database instance
// with the schemata provided in postgresMigrations. It returns the
// number of performed migrations.
func Migrate(storageType string, migrations map[string]string, migrationTable string, db *sqlx.DB) (int, error) {
	switch storageType {
	case POSTGRES:
		storageType = "postgres"
	default:
		return 0, fmt.Errorf("unsupported storageType %q", storageType)
	}

	memoryMigrations, err := parseMigration(migrations)
	if err != nil {
		return 0, err
	}

	migrate.SetTable(migrationTable)

	for i := range 3 {
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
