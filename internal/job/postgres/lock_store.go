package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/job"

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

	// Single atomic statement that handles three cases:
	//
	// 1. No active lock exists: expired function is no-op, INSERT succeeds -> acquired.
	// 2. Active but expired lock: expired function marks it completed (preserving history),
	//    INSERT succeeds because the expired row is excluded via NOT IN -> acquired.
	// 3. Active non-expired lock: expired function is no-op, NOT EXISTS finds the active
	//    row and blocks the INSERT -> not acquired.
	//
	// The CTE and main query share a snapshot, so the CTE's UPDATE is not
	// visible to NOT EXISTS; we exclude just-expired IDs explicitly. The
	// partial unique index on (name) WHERE completedat IS NULL prevents
	// concurrent duplicates.
	const q = `
WITH expired AS (
    UPDATE job_locks
    SET completedat = NOW(),
        status      = 'expired',
        message     = 'lock expired before completion'
    WHERE name = $1 AND completedat IS NULL AND expiresat < NOW()
    RETURNING id
)
INSERT INTO job_locks (name, owner, lockedat, expiresat)
SELECT $1, $2, NOW(), $3
WHERE NOT EXISTS (
    SELECT 1 FROM job_locks
    WHERE name = $1 AND completedat IS NULL
    AND id NOT IN (SELECT id FROM expired)
)
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

func (s *LockStore) Release(ctx context.Context, name, owner string, status job.Status, msg string) error {
	const q = `
UPDATE job_locks
SET completedat = NOW(),
    status      = $3,
    message     = $4
WHERE name = $1 AND owner = $2 AND completedat IS NULL
`

	_, err := s.db.ExecContext(ctx, q, name, owner, status.String(), msg)
	if err != nil {
		return fmt.Errorf("unable to release job lock %q: %w", name, err)
	}

	return nil
}

func (s *LockStore) ReleaseExpired(ctx context.Context) (int64, error) {
	const q = `
UPDATE job_locks
SET completedat = NOW(),
    status      = 'expired',
    message     = 'lock expired before completion'
WHERE expiresat < NOW() AND completedat IS NULL
`

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
