package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

type LockStore struct {
	db *sqlx.DB
}

func NewLockStore(db *sqlx.DB) *LockStore {
	return &LockStore{db: db}
}

func (s *LockStore) TryAcquire(ctx context.Context, name, owner string, timeout time.Duration) (bool, error) {
	expiresAt := time.Now().UTC().Add(timeout)

	// Atomically insert a lock row or, if one exists, take it over only when
	// the current lock has expired. This guarantees that at most one backend
	// owns a given job at any point in time.
	const q = `
INSERT INTO job_locks (name, owner, lockedat, expiresat)
VALUES ($1, $2, NOW(), $3)
ON CONFLICT (name) DO UPDATE
SET owner     = EXCLUDED.owner,
    lockedat  = NOW(),
    expiresat = EXCLUDED.expiresat
WHERE job_locks.expiresat < NOW()
`

	res, err := s.db.ExecContext(ctx, q, name, owner, expiresAt)
	if err != nil {
		return false, fmt.Errorf("unable to try acquire job lock %q: %w", name, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("unable to get rows affected for job lock %q: %w", name, err)
	}

	return rows > 0, nil
}

func (s *LockStore) Release(ctx context.Context, name, owner string) error {
	const q = `DELETE FROM job_locks WHERE name = $1 AND owner = $2`

	_, err := s.db.ExecContext(ctx, q, name, owner)
	if err != nil {
		return fmt.Errorf("unable to release job lock %q: %w", name, err)
	}

	return nil
}

func (s *LockStore) ReleaseExpired(ctx context.Context) (int64, error) {
	const q = `DELETE FROM job_locks WHERE expiresat < NOW()`

	res, err := s.db.ExecContext(ctx, q)
	if err != nil {
		return 0, fmt.Errorf("unable to release expired job locks: %w", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("unable to get rows affected for expired job locks: %w", err)
	}

	return n, nil
}
