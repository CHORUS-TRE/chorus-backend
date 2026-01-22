package postgres

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	common_storage "github.com/CHORUS-TRE/chorus-backend/pkg/common/storage"
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

func (s *AuditStorage) Record(ctx context.Context, entry *model.AuditEntry) (*model.AuditEntry, error) {
	const query = `
		INSERT INTO audit (tenantid, userid, username, action, resourcetype, resourceid, workspaceid, workbenchid, correlationid, method, statuscode, errormessage, description, details, createdat)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, tenantid, userid, username, action, resourcetype, resourceid, workspaceid, workbenchid, correlationid, method, statuscode, errormessage, description, details, createdat;
	`

	var createdEntry model.AuditEntry
	err := s.db.GetContext(ctx, &createdEntry, query,
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
	if err != nil {
		return nil, fmt.Errorf("unable to record audit entry: %w", err)
	}

	return &createdEntry, nil
}

func (s *AuditStorage) BulkRecord(ctx context.Context, entries []*model.AuditEntry) ([]*model.AuditEntry, error) {
	return nil, fmt.Errorf("Not implemented")
}

func (s *AuditStorage) List(ctx context.Context, pagination *common_model.Pagination, filter *model.AuditFilter) ([]*model.AuditEntry, *common_model.PaginationResult, error) {
	args := []interface{}{}

	filterClause := common_storage.BuildAuditFilterClause(filter, &args)

	// Get total count query
	countQuery := "SELECT COUNT(*) FROM audit"
	if filterClause != "" {
		countQuery += " WHERE " + filterClause
	}

	var totalCount int64
	err := s.db.GetContext(ctx, &totalCount, countQuery, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get count: %w", err)
	}

	// Get audit entries query
	query := "SELECT id, tenantid, userid, username, action, resourcetype, resourceid, workspaceid, workbenchid, correlationid, method, statuscode, errormessage, description, details, createdat FROM audit"
	if filterClause != "" {
		query += " WHERE " + filterClause
	}

	// Add pagination
	clause, validatedPagination := common_storage.BuildPaginationClause(pagination, model.AuditEntry{})
	query += clause

	// Build pagination result
	paginationRes := &common_model.PaginationResult{
		Total: uint64(totalCount),
	}

	if validatedPagination != nil {
		paginationRes.Limit = validatedPagination.Limit
		paginationRes.Offset = validatedPagination.Offset
		paginationRes.Sort = validatedPagination.Sort
	}

	var entries []*model.AuditEntry
	err = s.db.SelectContext(ctx, &entries, query, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to list audit entries: %w", err)
	}

	return entries, paginationRes, nil
}

func (s *AuditStorage) Count(ctx context.Context, filter *model.AuditFilter) (int64, error) {
	args := []interface{}{}
	filterClause := common_storage.BuildAuditFilterClause(filter, &args)

	query := "SELECT COUNT(*) FROM audit"
	if filterClause != "" {
		query += " WHERE " + filterClause
	}

	var count int64
	err := s.db.GetContext(ctx, &count, query, args...)
	if err != nil {
		return 0, err
	}

	return count, nil
}
