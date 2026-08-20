package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/migration"

	val "github.com/go-playground/validator/v10"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type migratableDatastore struct {
	ID      string
	Fetcher MigrationFetcher
}

var migratableDatastores = []migratableDatastore{
	{ID: "chorus", Fetcher: migration.GetMigration},
	{ID: "audit", Fetcher: migration.GetAuditMigration},
}

func lookupMigratableDatastore(id string) (migratableDatastore, bool) {
	for _, d := range migratableDatastores {
		if d.ID == id {
			return d, true
		}
	}
	return migratableDatastore{}, false
}

func migratableDatastoreIDs() []string {
	ids := make([]string, len(migratableDatastores))
	for i, d := range migratableDatastores {
		ids[i] = d.ID
	}
	return ids
}

func Migrate(datastoreID string) error {
	ctx := context.Background()

	ds, ok := lookupMigratableDatastore(datastoreID)
	if !ok {
		return fmt.Errorf("unknown datastore %q, expected one of: %s", datastoreID, strings.Join(migratableDatastoreIDs(), ", "))
	}

	resolved, err := resolveConfig()
	if err != nil {
		return err
	}

	cfg, ok := resolved.Storage.Datastores[datastoreID]
	if !ok {
		return fmt.Errorf("no storage configured for datastore %q", datastoreID)
	}

	creds, ok := resolved.Storage.Migrations[datastoreID]
	if !ok {
		return fmt.Errorf("no migration credentials configured for datastore %q", datastoreID)
	}
	cfg.Username = creds.Username
	cfg.Password = creds.Password

	if err := ProvideValidator().Struct(cfg); err != nil {
		var validationErrs val.ValidationErrors
		if !errors.As(err, &validationErrs) {
			return fmt.Errorf("datastore %q config is invalid: %w", datastoreID, err)
		}

		pathFor := func(relPath string) string { return datastoreConfigPath(datastoreID, relPath) }

		lines := make([]string, 0, len(validationErrs))
		for _, fe := range validationErrs {
			lines = append(lines, formatValidationError(cfg, pathFor, fe))
		}
		return fmt.Errorf("%s", strings.Join(lines, "\n"))
	}

	if cfg.Type != POSTGRES {
		return fmt.Errorf("unsupported database type for datastore %q: %q", datastoreID, cfg.Type)
	}

	db, err := connectPostgresForMigration(ctx, cfg)
	if err != nil {
		return fmt.Errorf("unable to connect to datastore %q: %w", datastoreID, err)
	}
	defer db.Close()

	migrations, migrationTable, err := ds.Fetcher(POSTGRES)
	if err != nil {
		return fmt.Errorf("unable to get migrations for datastore %q: %w", datastoreID, err)
	}

	n, err := migration.Migrate(POSTGRES, migrations, migrationTable, db)
	if err != nil {
		return fmt.Errorf("unable to migrate datastore %q: %w", datastoreID, err)
	}

	fmt.Printf("migrated datastore %q: %d migration(s) applied\n", datastoreID, n)
	return nil
}

func datastoreConfigPath(datastoreID, relPath string) string {
	field, _, _ := strings.Cut(relPath, ".")
	if field == "username" || field == "password" {
		return fmt.Sprintf("storage.migrations.%s.%s", datastoreID, relPath)
	}
	return fmt.Sprintf("storage.datastores.%s.%s", datastoreID, relPath)
}

func connectPostgresForMigration(_ context.Context, cfg config.Datastore) (*sqlx.DB, error) {
	var dataSourceName string
	if cfg.SSL.Enabled {
		dataSourceName = fmt.Sprintf("postgresql://%s@%s:%s/%s?sslmode=require&sslcert=%s&sslkey=%s&application_name=%s", cfg.Username, cfg.Host, cfg.Port, cfg.Database, cfg.SSL.CertificateFile, cfg.SSL.KeyFile, ProvideComponentInfo().Name)
		fmt.Printf("connecting to: postgresql://<redacted>@%s:%s/%s?sslmode=require&sslcert=<redacted>&sslkey=<redacted>&application_name=%s\n", cfg.Host, cfg.Port, cfg.Database, ProvideComponentInfo().Name)
	} else {
		dataSourceName = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable&application_name=%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database, ProvideComponentInfo().Name)
		fmt.Printf("connecting to: postgresql://<redacted>:<redacted>@%s:%s/%s?sslmode=disable&application_name=%s\n", cfg.Host, cfg.Port, cfg.Database, ProvideComponentInfo().Name)
	}

	db, err := sqlx.Connect("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(cfg.MaxConnections)
	db.SetConnMaxLifetime(cfg.MaxLifetime)
	return db, nil
}
