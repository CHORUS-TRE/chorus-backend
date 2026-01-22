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
	const query = `
		INSERT INTO audit (id, tenantid, userid, username, action, resourcetype, resourceid, workspaceid, workbenchid, correlationid, method, statuscode, errormessage, description, details, createdat)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16);
	`

	_, err := s.db.ExecContext(ctx, query,
		entry.ID,
		entry.TenantID,
		entry.UserID,
		entry.Username,
		entry.Action,
		entry.ResourceType,
		entry.ResourceID,
		entry.WorkspaceID,
		entry.WorkbenchID,
		entry.CorrelationID,
		entry.Method,
		entry.StatusCode,
		entry.ErrorMessage,
		entry.Description,
		entry.Details,
		entry.CreatedAt,
	)

	return err
}

func (s *AuditStorage) BulkRecord(ctx context.Context, entries []*model.AuditEntry) error {
	return fmt.Errorf("Not implemented")
}

func (s *AuditStorage) List(ctx context.Context, pagination *common_model.Pagination, filter *model.AuditFilter) ([]*model.AuditEntry, *common_model.PaginationResult, error) {
	return nil, nil, fmt.Errorf("Not implemented")
}

func (s *AuditStorage) Count(ctx context.Context, filter *model.AuditFilter) (int64, error) {
	return 0, fmt.Errorf("Not implemented")
}
