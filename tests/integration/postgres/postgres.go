//go:build integration

package postgres

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	embeddedpostgres "github.com/fergusstrange/embedded-postgres"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/CHORUS-TRE/chorus-backend/internal/migration"
	"github.com/CHORUS-TRE/chorus-backend/tests/integration"
)

var (
	once     sync.Once
	globalDB *sqlx.DB
	postgres *embeddedpostgres.EmbeddedPostgres
	initErr  error
	dbPort   uint32
)

var TruncateTablesBlacklist = map[string]struct{}{
	"chorus_migrations":                  {},
	"role_definitions":                   {},
	"role_definitions_required_contexts": {},
}

const (
	DefaultDatabase = "test_chorus"
	DefaultUser     = "postgres"
	DefaultPassword = "postgres"
)

// GetDB returns the shared test database connection.
// The database is automatically stopped when the process exits (via atexit handler)
// or when interrupted (SIGINT/SIGTERM).
func GetDB() (*sqlx.DB, error) {
	integration.TestSetup()
	once.Do(func() {
		initErr = initDatabase()
	})
	if initErr != nil {
		return nil, initErr
	}
	return globalDB, nil
}

func findFreePort() (uint32, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return uint32(listener.Addr().(*net.TCPAddr).Port), nil
}

func initDatabase() error {
	port, err := findFreePort()
	if err != nil {
		return fmt.Errorf("failed to find free port: %w", err)
	}
	dbPort = port

	postgres = embeddedpostgres.NewDatabase(
		embeddedpostgres.DefaultConfig().
			Port(dbPort).
			Database(DefaultDatabase).
			Username(DefaultUser).
			Password(DefaultPassword),
	)
	if err := postgres.Start(); err != nil {
		return fmt.Errorf("failed to start embedded postgres: %w", err)
	}

	// Register signal handler for graceful shutdown on interrupt
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		Stop()
		os.Exit(1)
	}()

	dsn := fmt.Sprintf("postgresql://%s:%s@localhost:%d/%s?sslmode=disable",
		DefaultUser, DefaultPassword, dbPort, DefaultDatabase)
	globalDB, err = sqlx.Connect("postgres", dsn)
	if err != nil {
		postgres.Stop()
		return fmt.Errorf("failed to connect to embedded postgres: %w", err)
	}

	migrations, tableName, err := migration.GetMigration("postgres")
	if err != nil {
		globalDB.Close()
		postgres.Stop()
		return fmt.Errorf("failed to get migrations: %w", err)
	}

	if _, err := migration.Migrate("postgres", migrations, tableName, globalDB); err != nil {
		globalDB.Close()
		postgres.Stop()
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// Stop closes the database connection and stops the embedded postgres.
func Stop() {
	if globalDB != nil {
		globalDB.Close()
		globalDB = nil
	}
	if postgres != nil {
		postgres.Stop()
		postgres = nil
	}
}

func TruncateTables(db *sqlx.DB, tables ...string) error {
	for _, t := range tables {
		if _, err := db.Exec("TRUNCATE " + t + " CASCADE"); err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", t, err)
		}
	}
	return nil
}

func GetTables(db *sqlx.DB) ([]string, error) {
	var tables []string
	rows, err := db.Queryx("SELECT relname AS table_name FROM pg_stat_user_tables ORDER BY table_name")
	if err != nil {
		return nil, fmt.Errorf("failed to query tables: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("failed to scan table name: %w", err)
		}
		tables = append(tables, tableName)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return tables, nil
}

func CleanupTables(db *sqlx.DB) error {
	tables, err := GetTables(db)
	if err != nil {
		return fmt.Errorf("failed to get tables: %w", err)
	}

	// Exclude blacklisted tables
	var filteredTables []string
	for _, t := range tables {
		if _, ok := TruncateTablesBlacklist[t]; !ok {
			filteredTables = append(filteredTables, t)
		}
	}

	return TruncateTables(db, filteredTables...)
}
