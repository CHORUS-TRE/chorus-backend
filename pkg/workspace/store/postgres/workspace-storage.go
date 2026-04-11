package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	cerr "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/uuid"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/common/storage"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
)

// pqStringArray converts a StringSlice to a pq.StringArray for use in SQL queries.
func pqStringArray(s model.StringSlice) pq.StringArray {
	return pq.StringArray(s)
}

// WorkspaceStorage is the handler through which a PostgresDB backend can be queried.
type WorkspaceStorage struct {
	db *sqlx.DB
}

// NewWorkspaceStorage returns a fresh workspace service storage instance.
func NewWorkspaceStorage(db *sqlx.DB) *WorkspaceStorage {
	return &WorkspaceStorage{db: db}
}

func (s *WorkspaceStorage) GetWorkspace(ctx context.Context, tenantID uint64, workspaceID uint64) (*model.Workspace, error) {
	const query = `
		SELECT id, tenantid, userid, name, shortname, description, status, ismain,
		       networkpolicy, allowedfqdns, networkpolicystatus, networkpolicymessage,
		       clipboard, services, servicestatuses,
		       createdat, updatedat
		FROM workspaces
		WHERE tenantid = $1 AND id = $2 AND deletedat IS NULL;
	`

	var workspace model.Workspace
	if err := s.db.GetContext(ctx, &workspace, query, tenantID, workspaceID); err != nil {
		return nil, err
	}

	return &workspace, nil
}

func (s *WorkspaceStorage) ListWorkspaces(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, IDin *[]uint64, allowDeleted bool) ([]*model.Workspace, *common_model.PaginationResult, error) {
	// Get total count query
	args := []interface{}{tenantID}

	countQuery := `SELECT COUNT(*) FROM workspaces WHERE tenantid = $1`
	if !allowDeleted {
		countQuery += " AND status != 'deleted' AND deletedat IS NULL"
	}
	if IDin != nil {
		countQuery += " AND id = ANY($2)"
		args = append(args, storage.Uint64ToPqInt64(*IDin))
	}

	var totalCount int
	if err := s.db.GetContext(ctx, &totalCount, countQuery, args...); err != nil {
		return nil, nil, err
	}

	// Get workspaces query
	query := `
		SELECT id, tenantid, userid, name, shortname, description, status, ismain,
		       networkpolicy, allowedfqdns, networkpolicystatus, networkpolicymessage,
		       clipboard, services, servicestatuses,
		       createdat, updatedat
		FROM workspaces
		WHERE tenantid = $1
	`

	if !allowDeleted {
		query += " AND status != 'deleted' AND deletedat IS NULL"
	}
	if IDin != nil {
		query += " AND id = ANY($2)"
	}

	// Add pagination
	clause, validatedPagination := storage.BuildPaginationClause(pagination, model.Workspace{})
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

	var workspaces []*model.Workspace
	if err := s.db.SelectContext(ctx, &workspaces, query, args...); err != nil {
		return nil, nil, err
	}

	return workspaces, paginationRes, nil
}

// CreateWorkspace saves the provided workspace object in the database 'workspaces' table.
func (s *WorkspaceStorage) CreateWorkspace(ctx context.Context, tenantID uint64, workspace *model.Workspace) (*model.Workspace, error) {
	const workspaceQuery = `
		INSERT INTO workspaces (tenantid, userid, name, shortname, description, status, ismain,
		                        networkpolicy, allowedfqdns, clipboard, services,
		                        createdat, updatedat)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		RETURNING id, tenantid, userid, name, shortname, description, status, ismain,
		          networkpolicy, allowedfqdns, networkpolicystatus, networkpolicymessage,
		          clipboard, services, servicestatuses,
		          createdat, updatedat;
	`

	servicesVal, err := workspace.Services.Value()
	if err != nil {
		return nil, fmt.Errorf("unable to marshal services: %w", err)
	}

	networkPolicy := workspace.NetworkPolicy
	if networkPolicy == "" {
		networkPolicy = "Airgapped"
	}
	clipboard := workspace.Clipboard
	if clipboard == "" {
		clipboard = "disabled"
	}
	allowedFQDNs := workspace.AllowedFQDNs
	if allowedFQDNs == nil {
		allowedFQDNs = model.StringSlice{}
	}

	var createdWorkspace model.Workspace
	err = s.db.GetContext(ctx, &createdWorkspace, workspaceQuery,
		tenantID, workspace.UserID, workspace.Name, workspace.ShortName, workspace.Description, workspace.Status, workspace.IsMain,
		networkPolicy, pqStringArray(allowedFQDNs), clipboard, servicesVal,
	)
	if err != nil {
		return nil, err
	}

	return &createdWorkspace, nil
}

func (s *WorkspaceStorage) UpdateWorkspace(ctx context.Context, tenantID uint64, workspace *model.Workspace) (*model.Workspace, error) {
	const workspaceUpdateQuery = `
		UPDATE workspaces
		SET name = $3, shortname = $4, description = $5, status = $6, isMain = $7,
		    networkpolicy = $8, allowedfqdns = $9, clipboard = $10, services = $11,
		    updatedat = NOW()
		WHERE tenantid = $1 AND id = $2 AND deletedat IS NULL
		RETURNING id, tenantid, userid, name, shortname, description, status, ismain,
		          networkpolicy, allowedfqdns, networkpolicystatus, networkpolicymessage,
		          clipboard, services, servicestatuses,
		          createdat, updatedat;
	`

	servicesVal, err := workspace.Services.Value()
	if err != nil {
		return nil, fmt.Errorf("unable to marshal services: %w", err)
	}

	networkPolicy := workspace.NetworkPolicy
	if networkPolicy == "" {
		networkPolicy = "Airgapped"
	}
	clipboard := workspace.Clipboard
	if clipboard == "" {
		clipboard = "disabled"
	}
	allowedFQDNs := workspace.AllowedFQDNs
	if allowedFQDNs == nil {
		allowedFQDNs = model.StringSlice{}
	}

	var updatedWorkspace model.Workspace
	err = s.db.GetContext(ctx, &updatedWorkspace, workspaceUpdateQuery,
		tenantID, workspace.ID, workspace.Name, workspace.ShortName, workspace.Description, workspace.Status, workspace.IsMain,
		networkPolicy, pqStringArray(allowedFQDNs), clipboard, servicesVal,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to update workspace: %w", err)
	}

	return &updatedWorkspace, nil
}

// UpdateWorkspaceStatus updates only the status fields (from K8s watcher).
func (s *WorkspaceStorage) UpdateWorkspaceStatus(ctx context.Context, tenantID uint64, workspaceID uint64, networkPolicyStatus, networkPolicyMessage string, serviceStatuses model.JSONMap[model.WorkspaceServiceStatusInfo]) error {
	const query = `
		UPDATE workspaces
		SET networkpolicystatus = $3, networkpolicymessage = $4, servicestatuses = $5, updatedat = NOW()
		WHERE tenantid = $1 AND id = $2 AND deletedat IS NULL;
	`

	serviceStatusesVal, err := serviceStatuses.Value()
	if err != nil {
		return fmt.Errorf("unable to marshal service statuses: %w", err)
	}

	_, err = s.db.ExecContext(ctx, query, tenantID, workspaceID, networkPolicyStatus, networkPolicyMessage, serviceStatusesVal)
	if err != nil {
		return fmt.Errorf("unable to update workspace status: %w", err)
	}

	return nil
}

func (s *WorkspaceStorage) DeleteWorkspace(ctx context.Context, tenantID uint64, workspaceID uint64) error {
	const query = `
		UPDATE workspaces	SET 
			(status, name, updatedat, deletedat) = 
			($3, concat(name, $4::TEXT), NOW(), NOW())
		WHERE tenantid = $1 AND id = $2 AND deletedat IS NULL;
	`

	rows, err := s.db.ExecContext(ctx, query, tenantID, workspaceID, model.WorkspaceDeleted.String(), "-"+uuid.Next())
	if err != nil {
		return fmt.Errorf("unable to exec: %w", err)
	}
	affected, err := rows.RowsAffected()
	if err != nil {
		return fmt.Errorf("unable to get rows affected: %w", err)
	}

	if affected == 0 {
		return cerr.ErrNoRowsDeleted
	}

	return nil
}

func (s *WorkspaceStorage) DeleteOldWorkspaces(ctx context.Context, timeout time.Duration) ([]*model.Workspace, error) {
	const query = `
		UPDATE workspaces
		SET (status, name, updatedat, deletedat) = ($1, concat(name, $2::TEXT), NOW(), NOW())
		WHERE createdat < NOW() - $3::INTERVAL
		  AND status != 'deleted'
		  AND deletedat IS NULL
		RETURNING id, tenantid, userid, name, shortname, description, status, ismain,
		          networkpolicy, allowedfqdns, networkpolicystatus, networkpolicymessage,
		          clipboard, services, servicestatuses,
		          createdat, updatedat;
	`

	var deletedWorkspaces []*model.Workspace
	err := s.db.SelectContext(ctx, &deletedWorkspaces, query, model.WorkspaceDeleted.String(), "-"+uuid.Next(), fmt.Sprintf("%d seconds", int64(timeout.Seconds())))
	if err != nil {
		return nil, fmt.Errorf("unable to exec: %w", err)
	}

	return deletedWorkspaces, nil
}
