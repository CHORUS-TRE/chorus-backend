package provider

import (
	"context"
	"fmt"
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/migration"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/database"
	_ "github.com/lib/pq"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	POSTGRES = "postgres"
)

type Database struct {
	DB   database.DB
	Type string
}

func (d *Database) GetSqlxDB() *sqlx.DB {
	return d.DB.GetSqlxDB()
}

func ProvideMainDB(opts ...Option) *Database {
	return ProvideDB("chorus", opts...)
}

var dbsOnce sync.Once
var dbs map[string]map[string]*Database

func ProvideDB(datastoreID string, opts ...Option) *Database {
	ctx := context.Background()

	dbsOnce.Do(func() {
		dbs = map[string]map[string]*Database{}
		cfg := ProvideConfig()

		for id, s := range cfg.Storage.Datastores {
			if id == "" {
				logger.TechLog.Fatal(ctx, fmt.Sprintf("invalid storage ID: '%v'", id), zap.Any("storage", s))
			}
			dbs[id] = map[string]*Database{}
		}
	})

	o := &options{}

	// Apply options.
	for _, opt := range opts {
		opt(o)
	}
	datastorEnv := viper.Get("datastore")
	if datastorEnv != nil && datastorEnv.(string) != "" {
		datastoreID = datastorEnv.(string)
	}

	m, ok := dbs[datastoreID]
	if !ok {
		logger.TechLog.Fatal(ctx, fmt.Sprintf("no storage '%v' found", datastoreID))
	}

	if _, ok = m[o.clientName]; !ok {
		cfg := ProvideConfig().Storage.Datastores[datastoreID]
		switch cfg.Type {
		case POSTGRES:
			m[o.clientName] = providePostgresDB(ctx, cfg, o)
		default:
			logger.TechLog.Fatal(ctx, fmt.Sprintf("invalid storage type. Must be 'postgres', got: '%v'", cfg.Type))
		}
	}

	return m[o.clientName]
}

type MigrationFetcher func(string) (map[string]string, string, error)

type options struct {
	clientName string
	f          MigrationFetcher
}

// Option is used to pass options to the DB provider.
type Option func(*options)

// WithClient is the option used to set the client name.
func WithClient(client string) Option {
	return func(o *options) {
		o.clientName = client
	}
}

// WithMigrations is the option used to set the migrations.
func WithMigrations(f MigrationFetcher) Option {
	return func(o *options) {
		o.f = f
	}
}

func providePostgresDB(ctx context.Context, cfg config.Datastore, opts *options) *Database {
	var dataSourceName string
	if cfg.SSL.Enabled {
		dataSourceName = fmt.Sprintf("postgresql://%s@%s:%s/%s?sslmode=require&sslcert=%s&sslkey=%s&application_name=%s", cfg.Username, cfg.Host, cfg.Port, cfg.Database, cfg.SSL.CertificateFile, cfg.SSL.KeyFile, ProvideComponentInfo().Name)
		logger.TechLog.Info(ctx, "connecting to: "+fmt.Sprintf("postgresql://<redacted>@%s:%s/%s?sslmode=require&sslcert=<redacted>&sslkey=<redacted>&application_name=%s", cfg.Host, cfg.Port, cfg.Database, ProvideComponentInfo().Name))
	} else {
		dataSourceName = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable&application_name=%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database, ProvideComponentInfo().Name)
		logger.TechLog.Info(ctx, "connecting to: "+fmt.Sprintf("postgresql://<redacted>:<redacted>@%s:%s/%s?sslmode=disable&application_name=%s", cfg.Host, cfg.Port, cfg.Database, ProvideComponentInfo().Name))
	}

	db, err := sqlx.Connect("postgres", dataSourceName)
	if err != nil {
		logger.TechLog.Fatal(ctx, "unable to connect to db", zap.Error(err))
	}
	db.SetMaxOpenConns(cfg.MaxConnections)
	db.SetConnMaxLifetime(cfg.MaxLifetime)

	// Do migrations only once.
	if opts.f != nil {
		migrations, migrationTable, err := opts.f(POSTGRES)
		if err != nil {
			logger.TechLog.Fatal(ctx, "unable to get migration", zap.Error(err))
		}

		if migrations != nil && migrationTable != "" {
			n, err := migration.Migrate(POSTGRES, migrations, migrationTable, db)
			if err != nil {
				logger.TechLog.Fatal(ctx, "unable to migrate database "+opts.clientName, zap.Error(err))
			}
			logger.TechLog.Info(ctx, "migrated database: "+opts.clientName, zap.Int("num_migrations", n))
		}
	}

	return &Database{
		DB:   database.NewDefaultDB(db),
		Type: POSTGRES,
	}
}
