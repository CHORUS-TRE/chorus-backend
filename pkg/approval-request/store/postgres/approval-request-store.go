package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/common/storage"

	"github.com/CHORUS-TRE/chorus-backend/pkg/approval-request/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/approval-request/service"
)

var _ service.ApprovalRequestStore = (*ApprovalRequestStorage)(nil)

type ApprovalRequestStorage struct {
	db *sqlx.DB
}

func NewApprovalRequestStorage(db *sqlx.DB) *ApprovalRequestStorage {
	return &ApprovalRequestStorage{db: db}
}

type approvalRequestRow struct {
	ID              uint64        `db:"id"`
	TenantID        uint64        `db:"tenantid"`
	RequesterID     uint64        `db:"requesterid"`
	Type            string        `db:"type"`
	Status          string        `db:"status"`
	Title           string        `db:"title"`
	Description     string        `db:"description"`
	Details         []byte        `db:"details"`
	ApproverIDs     pq.Int64Array `db:"approverids"`
	ApprovedByID    *uint64       `db:"approvedbyid"`
	AutoApproved    bool          `db:"autoapproved"`
	ApprovalMessage string        `db:"approvalmessage"`
	CreatedAt       time.Time     `db:"createdat"`
	UpdatedAt       time.Time     `db:"updatedat"`
	ApprovedAt      *time.Time    `db:"approvedat"`
}

type approvalRequestCountsRow struct {
	Total          uint64 `db:"total"`
	TotalApprover  uint64 `db:"total_approver"`
	TotalRequester uint64 `db:"total_requester"`
}

type approvalRequestGroupedCountRow struct {
	Key   string `db:"key"`
	Count uint64 `db:"count"`
}

func (r *approvalRequestRow) toModel() (*model.ApprovalRequest, error) {
	var details model.ApprovalRequestDetails
	if len(r.Details) > 0 {
		if err := json.Unmarshal(r.Details, &details); err != nil {
			return nil, fmt.Errorf("unable to unmarshal details: %w", err)
		}
	}

	return &model.ApprovalRequest{
		ID:              r.ID,
		TenantID:        r.TenantID,
		RequesterID:     r.RequesterID,
		Type:            model.ApprovalRequestType(r.Type),
		Status:          model.ApprovalRequestStatus(r.Status),
		Title:           r.Title,
		Description:     r.Description,
		Details:         details,
		ApproverIDs:     storage.PqInt64ToUint64(r.ApproverIDs),
		ApprovedByID:    r.ApprovedByID,
		AutoApproved:    r.AutoApproved,
		ApprovalMessage: r.ApprovalMessage,
		CreatedAt:       r.CreatedAt,
		UpdatedAt:       r.UpdatedAt,
		ApprovedAt:      r.ApprovedAt,
	}, nil
}

func (s *ApprovalRequestStorage) GetApprovalRequest(ctx context.Context, tenantID, requestID uint64) (*model.ApprovalRequest, error) {
	const query = `
		SELECT id, tenantid, requesterid, type, status, title, description, details, approverids, approvedbyid, autoapproved, approvalmessage, createdat, updatedat, approvedat
		FROM approval_requests
		WHERE tenantid = $1 AND id = $2 AND deletedat IS NULL
	`

	var row approvalRequestRow
	if err := s.db.GetContext(ctx, &row, query, tenantID, requestID); err != nil {
		return nil, fmt.Errorf("unable to get approval request: %w", err)
	}

	return row.toModel()
}

func (s *ApprovalRequestStorage) ListApprovalRequests(ctx context.Context, tenantID, userID uint64, pagination *common_model.Pagination, filter service.ApprovalRequestFilter) ([]*model.ApprovalRequest, *common_model.PaginationResult, error) {
	args := []interface{}{tenantID, userID}

	whereClause := "WHERE tenantid = $1 AND (requesterid = $2 OR $2 = ANY(approverids)) AND deletedat IS NULL"

	if filter.StatusesIn != nil && len(*filter.StatusesIn) > 0 {
		statuses := make([]string, len(*filter.StatusesIn))
		for i, status := range *filter.StatusesIn {
			statuses[i] = string(status)
		}
		args = append(args, pq.StringArray(statuses))
		whereClause += fmt.Sprintf(" AND status = ANY($%d)", len(args))
	}

	if filter.TypesIn != nil && len(*filter.TypesIn) > 0 {
		types := make([]string, len(*filter.TypesIn))
		for i, t := range *filter.TypesIn {
			types[i] = string(t)
		}
		args = append(args, pq.StringArray(types))
		whereClause += fmt.Sprintf(" AND type = ANY($%d)", len(args))
	}

	if filter.SourceWorkspaceID != nil {
		args = append(args, *filter.SourceWorkspaceID)
		whereClause += fmt.Sprintf(" AND (details->'data_extraction_details'->>'source_workspace_id' = $%d::TEXT OR details->'data_transfer_details'->>'source_workspace_id' = $%d::TEXT)", len(args), len(args))
	}

	if filter.PendingApproval != nil && *filter.PendingApproval {
		whereClause += " AND status = 'pending'"
	}

	if filter.ApproverID != nil {
		args = append(args, *filter.ApproverID)
		whereClause += fmt.Sprintf(" AND $%d = ANY(approverids)", len(args))
	}

	if filter.RequesterID != nil {
		args = append(args, *filter.RequesterID)
		whereClause += fmt.Sprintf(" AND requesterid = $%d", len(args))
	}

	countQuery := "SELECT COUNT(*) FROM approval_requests " + whereClause
	var totalCount int
	if err := s.db.GetContext(ctx, &totalCount, countQuery, args...); err != nil {
		return nil, nil, fmt.Errorf("unable to count approval requests: %w", err)
	}

	selectQuery := `
		SELECT id, tenantid, requesterid, type, status, title, description, details, approverids, approvedbyid, autoapproved, approvalmessage, createdat, updatedat, approvedat
		FROM approval_requests ` + whereClause

	paginationClause, validatedPagination := storage.BuildPaginationClause(pagination, model.ApprovalRequest{})
	selectQuery += paginationClause

	paginationRes := &common_model.PaginationResult{
		Total: uint64(totalCount),
	}
	if validatedPagination != nil {
		paginationRes.Limit = validatedPagination.Limit
		paginationRes.Offset = validatedPagination.Offset
		paginationRes.Sort = validatedPagination.Sort
	}

	var rows []approvalRequestRow
	if err := s.db.SelectContext(ctx, &rows, selectQuery, args...); err != nil {
		return nil, nil, fmt.Errorf("unable to list approval requests: %w", err)
	}

	requests := make([]*model.ApprovalRequest, len(rows))
	for i, row := range rows {
		req, err := row.toModel()
		if err != nil {
			return nil, nil, err
		}
		requests[i] = req
	}

	return requests, paginationRes, nil
}

func newApprovalRequestStatusCountMap() map[string]uint64 {
	counts := make(map[string]uint64, len(model.ApprovalRequestStatuses()))
	for _, status := range model.ApprovalRequestStatuses() {
		counts[string(status)] = 0
	}
	return counts
}

func newApprovalRequestTypeCountMap() map[string]uint64 {
	counts := make(map[string]uint64, len(model.ApprovalRequestTypes()))
	for _, requestType := range model.ApprovalRequestTypes() {
		counts[string(requestType)] = 0
	}
	return counts
}

func (s *ApprovalRequestStorage) CountMyApprovalRequests(ctx context.Context, tenantID, userID uint64) (*model.ApprovalRequestCounts, error) {
	const baseWhereClause = `
		FROM approval_requests
		WHERE tenantid = $1 AND deletedat IS NULL AND (requesterid = $2 OR $2 = ANY(approverids))
	`

	var summary approvalRequestCountsRow
	if err := s.db.GetContext(ctx, &summary, `
		SELECT
			COUNT(*) AS total,
			COUNT(*) FILTER (WHERE $2 = ANY(approverids)) AS total_approver,
			COUNT(*) FILTER (WHERE requesterid = $2) AS total_requester
		`+baseWhereClause, tenantID, userID); err != nil {
		return nil, fmt.Errorf("unable to count approval requests: %w", err)
	}

	countByStatus := newApprovalRequestStatusCountMap()
	var statusRows []approvalRequestGroupedCountRow
	if err := s.db.SelectContext(ctx, &statusRows, `
		SELECT status AS key, COUNT(*) AS count
		`+baseWhereClause+`
		GROUP BY status
	`, tenantID, userID); err != nil {
		return nil, fmt.Errorf("unable to count approval requests by status: %w", err)
	}
	for _, row := range statusRows {
		countByStatus[row.Key] = row.Count
	}

	countByType := newApprovalRequestTypeCountMap()
	var typeRows []approvalRequestGroupedCountRow
	if err := s.db.SelectContext(ctx, &typeRows, `
		SELECT type AS key, COUNT(*) AS count
		`+baseWhereClause+`
		GROUP BY type
	`, tenantID, userID); err != nil {
		return nil, fmt.Errorf("unable to count approval requests by type: %w", err)
	}
	for _, row := range typeRows {
		countByType[row.Key] = row.Count
	}

	return &model.ApprovalRequestCounts{
		Total:          summary.Total,
		TotalApprover:  summary.TotalApprover,
		TotalRequester: summary.TotalRequester,
		CountByStatus:  countByStatus,
		CountByType:    countByType,
	}, nil
}

func (s *ApprovalRequestStorage) CreateApprovalRequest(ctx context.Context, tenantID uint64, request *model.ApprovalRequest) (*model.ApprovalRequest, error) {
	detailsJSON, err := json.Marshal(request.Details)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal details: %w", err)
	}

	const query = `
		INSERT INTO approval_requests (tenantid, requesterid, type, status, title, description, details, approverids, autoapproved, approvalmessage, createdat, updatedat)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING id, tenantid, requesterid, type, status, title, description, details, approverids, approvedbyid, autoapproved, approvalmessage, createdat, updatedat, approvedat
	`

	var row approvalRequestRow
	err = s.db.GetContext(ctx, &row, query,
		tenantID,
		request.RequesterID,
		string(request.Type),
		string(request.Status),
		request.Title,
		request.Description,
		detailsJSON,
		storage.Uint64ToPqInt64(request.ApproverIDs),
		request.AutoApproved,
		request.ApprovalMessage,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create approval request: %w", err)
	}

	return row.toModel()
}

func (s *ApprovalRequestStorage) UpdateApprovalRequest(ctx context.Context, tenantID uint64, request *model.ApprovalRequest) (*model.ApprovalRequest, error) {
	detailsJSON, err := json.Marshal(request.Details)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal details: %w", err)
	}

	var query string
	var args []interface{}

	if request.Status == model.ApprovalRequestStatusApproved || request.Status == model.ApprovalRequestStatusRejected {
		query = `
			UPDATE approval_requests
			SET type = $3, status = $4, title = $5, description = $6, details = $7, approverids = $8, approvedbyid = $9, autoapproved = $10, approvalmessage = $11, approvedat = NOW(), updatedat = NOW()
			WHERE tenantid = $1 AND id = $2 AND deletedat IS NULL
			RETURNING id, tenantid, requesterid, type, status, title, description, details, approverids, approvedbyid, autoapproved, approvalmessage, createdat, updatedat, approvedat
		`
		args = []interface{}{
			tenantID,
			request.ID,
			string(request.Type),
			string(request.Status),
			request.Title,
			request.Description,
			detailsJSON,
			storage.Uint64ToPqInt64(request.ApproverIDs),
			request.ApprovedByID,
			request.AutoApproved,
			request.ApprovalMessage,
		}
	} else {
		query = `
			UPDATE approval_requests
			SET type = $3, status = $4, title = $5, description = $6, details = $7, approverids = $8, autoapproved = $9, approvalmessage = $10, updatedat = NOW()
			WHERE tenantid = $1 AND id = $2 AND deletedat IS NULL
			RETURNING id, tenantid, requesterid, type, status, title, description, details, approverids, approvedbyid, autoapproved, approvalmessage, createdat, updatedat, approvedat
		`
		args = []interface{}{
			tenantID,
			request.ID,
			string(request.Type),
			string(request.Status),
			request.Title,
			request.Description,
			detailsJSON,
			storage.Uint64ToPqInt64(request.ApproverIDs),
			request.AutoApproved,
			request.ApprovalMessage,
		}
	}

	var row approvalRequestRow
	err = s.db.GetContext(ctx, &row, query, args...)
	if err != nil {
		return nil, fmt.Errorf("unable to update approval request: %w", err)
	}

	return row.toModel()
}

func (s *ApprovalRequestStorage) DeleteApprovalRequest(ctx context.Context, tenantID, requestID uint64) error {
	const query = `
		UPDATE approval_requests
		SET deletedat = NOW(), updatedat = NOW()
		WHERE tenantid = $1 AND id = $2 AND deletedat IS NULL
	`

	result, err := s.db.ExecContext(ctx, query, tenantID, requestID)
	if err != nil {
		return fmt.Errorf("unable to delete approval request: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("unable to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("approval request not found")
	}

	return nil
}
