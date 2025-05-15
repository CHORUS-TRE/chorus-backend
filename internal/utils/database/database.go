package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

var (
	ErrNoRowsUpdated = errors.New("database: no rows updated")
	ErrNoRowsDeleted = errors.New("database: no rows deleted")
)

// Meant to be a constant, but cannot declare a nil const.
// This exists only to make store calls more readable when no transaction is needed
var NoTransaction Queryable

// Queryable represents an object on which a SQL query can be executed.
// It's used to abstract both a DB connection and a sqlx transaction
type Queryable interface {
	Get(dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	Exec(query string, args ...interface{}) (sql.Result, error)
	NamedExecContext(ctx context.Context, query string, arg interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// DB is the interface of the storages. Each storage (postgres,...) must
// implement this interface.
type DB interface {
	Queryable
	GetDB() *sql.DB
	Beginx() (Tx, error)
	Close() error
	// GetSqlxDB returns a sqlx.DB instance, which is needed as the Job library needs a
	// .Beginx() *sqlx.Tx method
	GetSqlxDB() *sqlx.DB
}

// Tx is the interface for a DB transaction. It is queryable and can be "closed"
type Tx interface {
	Queryable
	Commit() error
	Rollback() error
}

type Database interface {
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type TxDB interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

type wrappedDB struct {
	*sqlx.DB
}

func (db *wrappedDB) GetDB() *sql.DB {
	return db.DB.DB
}

func (db *wrappedDB) GetSqlxDB() *sqlx.DB {
	return db.DB
}

func NewDefaultDB(db *sqlx.DB) DB {
	return &wrappedDB{DB: db}
}

func (db *wrappedDB) Beginx() (Tx, error) {
	return db.DB.Beginx()
}

func (db *wrappedDB) Select(dest interface{}, query string, args ...interface{}) error {
	logger.TechLog.Debug(context.Background(), "executing query", zap.String("query", query))
	return db.DB.Select(dest, query, args...)
}

func (db *wrappedDB) Get(dest interface{}, query string, args ...interface{}) error {
	logger.TechLog.Debug(context.Background(), "executing query", zap.String("query", query))
	return db.DB.Get(dest, query, args...)
}

func (db *wrappedDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	logger.TechLog.Debug(ctx, "executing query", zap.String("query", query))
	return db.DB.ExecContext(ctx, query, args...)
}

func (db *wrappedDB) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	logger.TechLog.Debug(ctx, "executing query", zap.String("query", query))
	return db.DB.SelectContext(ctx, dest, query, args...)
}

func GetQueryParam[T any](param string, query map[string][]string, target *[]T) error {
	if param == "" {
		return fmt.Errorf("param cannot be empty")
	}

	if query == nil {
		return fmt.Errorf("query cannot be nil")
	}

	if target == nil {
		return fmt.Errorf("target cannot be nil")
	}

	if _, ok := query[param]; !ok {
		return fmt.Errorf("param %s not found in query", param)
	}

	*target = make([]T, len(query[param]))
	for i, value := range query[param] {
		var t T
		_, err := fmt.Sscanf(value, "%v", &t)
		if err != nil {
			return fmt.Errorf("failed to parse value %s: %w", value, err)
		}
		(*target)[i] = t
	}

	return nil
}
