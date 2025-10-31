package postgres

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/pkg/devstore/model"
	"github.com/jmoiron/sqlx"
)

// DevstoreStorage is the handler through which a PostgresDB backend can be queried.
type DevstoreStorage struct {
	db *sqlx.DB
}

// NewDevstoreStorage returns a fresh devstore service storage instance.
func NewDevstoreStorage(db *sqlx.DB) *DevstoreStorage {
	return &DevstoreStorage{db: db}
}

func (s *DevstoreStorage) ListEntries(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64) ([]*model.DevstoreEntry, error) {
	query := `
		SELECT id, tenantid, scope, scopeid, "key", "value", createdat, updatedat
		FROM devstore
		WHERE tenantid = $1 AND scope = $2 AND scopeid = $3;
	`

	var entries []*model.DevstoreEntry
	if err := s.db.SelectContext(ctx, &entries, query, tenantID, scope, scopeID); err != nil {
		return nil, err
	}

	return entries, nil
}

func (s *DevstoreStorage) GetEntry(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64, key string) (*model.DevstoreEntry, error) {
	query := `
		SELECT id, tenantid, scope, scopeid, "key", "value", createdat, updatedat
		FROM devstore
		WHERE tenantid = $1 AND scope = $2 AND scopeid = $3 AND "key" = $4;
	`

	var entry model.DevstoreEntry
	if err := s.db.GetContext(ctx, &entry, query, tenantID, scope, scopeID, key); err != nil {
		return nil, err
	}

	return &entry, nil
}

func (s *DevstoreStorage) PutEntry(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64, key string, value string) (*model.DevstoreEntry, error) {
	query := `
		INSERT INTO devstore (tenantid, scope, scopeid, "key", "value", createdat, updatedat)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		ON CONFLICT (tenantid, scope, scopeid, "key") DO UPDATE SET "value" = EXCLUDED."value", updatedat = NOW()
		RETURNING id, scope, scopeid, "key", "value", createdat, updatedat;
	`

	var entry model.DevstoreEntry
	if err := s.db.GetContext(ctx, &entry, query, tenantID, scope, scopeID, key, value); err != nil {
		return nil, err
	}

	return &entry, nil
}

func (s *DevstoreStorage) DeleteEntry(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64, key string) error {
	query := `
		DELETE FROM devstore
		WHERE tenantid = $1 AND scope = $2 AND scopeid = $3 AND "key" = $4;
	`

	_, err := s.db.ExecContext(ctx, query, tenantID, scope, scopeID, key)
	return err
}
