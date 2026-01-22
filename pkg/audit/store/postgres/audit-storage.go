package postgres

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/jmoiron/sqlx"
)

// AuditStorage is the handler through which a PostgresDB backend can be queried.
type AuditStorage struct {
	db *sqlx.DB
}

// NewAuditStorage returns a fresh audit service storage instance.
func NewAuditStorage(db *sqlx.DB) *AuditStorage {
	return &AuditStorage{db: db}
}

func (s *AuditStorage) Record(ctx context.Context, entry *model.AuditEntry) error {
	return fmt.Errorf("Not implemented")
}

func (s *AuditStorage) RecordBatch(ctx context.Context, entries []*model.AuditEntry) error {
	return fmt.Errorf("Not implemented")
}

func (s *AuditStorage) List(ctx context.Context, pagination *common_model.Pagination, filter *model.AuditFilter) ([]*model.AuditEntry, *common_model.PaginationResult, error) {
	return nil, nil, fmt.Errorf("Not implemented")
}

func (s *AuditStorage) Count(ctx context.Context, filter *model.AuditFilter) (int64, error) {
	return 0, fmt.Errorf("Not implemented")
}
