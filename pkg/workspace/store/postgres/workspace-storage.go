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
		       clipboard,
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
		       clipboard,
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
		                        networkpolicy, allowedfqdns, clipboard,
		                        createdat, updatedat)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		RETURNING id, tenantid, userid, name, shortname, description, status, ismain,
		          networkpolicy, allowedfqdns, networkpolicystatus, networkpolicymessage,
		          clipboard,
		          createdat, updatedat;
	`

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
	err := s.db.GetContext(ctx, &createdWorkspace, workspaceQuery,
		tenantID, workspace.UserID, workspace.Name, workspace.ShortName, workspace.Description, workspace.Status, workspace.IsMain,
		networkPolicy, pqStringArray(allowedFQDNs), clipboard,
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
		    networkpolicy = $8, allowedfqdns = $9, clipboard = $10,
		    updatedat = NOW()
		WHERE tenantid = $1 AND id = $2 AND deletedat IS NULL
		RETURNING id, tenantid, userid, name, shortname, description, status, ismain,
		          networkpolicy, allowedfqdns, networkpolicystatus, networkpolicymessage,
		          clipboard,
		          createdat, updatedat;
	`

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
	err := s.db.GetContext(ctx, &updatedWorkspace, workspaceUpdateQuery,
		tenantID, workspace.ID, workspace.Name, workspace.ShortName, workspace.Description, workspace.Status, workspace.IsMain,
		networkPolicy, pqStringArray(allowedFQDNs), clipboard,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to update workspace: %w", err)
	}

	return &updatedWorkspace, nil
}

// UpdateWorkspaceStatus updates only the workspace-level status fields (from K8s watcher).
func (s *WorkspaceStorage) UpdateWorkspaceStatus(ctx context.Context, tenantID uint64, workspaceID uint64, networkPolicyStatus, networkPolicyMessage string) error {
	const query = `
		UPDATE workspaces
		SET networkpolicystatus = $3, networkpolicymessage = $4, updatedat = NOW()
		WHERE tenantid = $1 AND id = $2 AND deletedat IS NULL;
	`

	_, err := s.db.ExecContext(ctx, query, tenantID, workspaceID, networkPolicyStatus, networkPolicyMessage)
	if err != nil {
		return fmt.Errorf("unable to update workspace status: %w", err)
	}

	return nil
}

// UpdateWorkspaceSvcStatuses batch-updates status fields for workspace services (from K8s watcher).
func (s *WorkspaceStorage) UpdateWorkspaceSvcStatuses(ctx context.Context, workspaceID uint64, statuses map[string]model.WorkspaceSvcStatusUpdate) error {
	const query = `
		UPDATE workspace_services
		SET status = $3, statusmessage = $4, connectioninfo = $5, secretname = $6, updatedat = NOW()
		WHERE workspaceid = $1 AND name = $2 AND deletedat IS NULL;
	`

	for name, st := range statuses {
		_, err := s.db.ExecContext(ctx, query, workspaceID, name, st.Status, st.StatusMessage, st.ConnectionInfo, st.SecretName)
		if err != nil {
			return fmt.Errorf("unable to update workspace service status %q: %w", name, err)
		}
	}

	return nil
}

func (s *WorkspaceStorage) GetWorkspaceSvc(ctx context.Context, tenantID, workspaceSvcID uint64) (*model.WorkspaceSvc, error) {
	const query = `
		SELECT id, tenantid, workspaceid, name,
		       state, chartregistry, chartrepository, charttag,
		       valuesoverride, credentialssecretname, credentialspaths,
		       connectioninfotemplate, computedvalues,
		       status, statusmessage, connectioninfo, secretname,
		       createdat, updatedat
		FROM workspace_services
		WHERE tenantid = $1 AND id = $2 AND deletedat IS NULL;
	`

	var svc model.WorkspaceSvc
	if err := s.db.GetContext(ctx, &svc, query, tenantID, workspaceSvcID); err != nil {
		return nil, err
	}

	return &svc, nil
}

func (s *WorkspaceStorage) ListWorkspaceSvcs(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, workspaceIDsIn *[]uint64) ([]*model.WorkspaceSvc, *common_model.PaginationResult, error) {
	args := []interface{}{tenantID}

	countQuery := `SELECT COUNT(*) FROM workspace_services WHERE tenantid = $1 AND deletedat IS NULL`
	if workspaceIDsIn != nil {
		countQuery += " AND workspaceid = ANY($2)"
		args = append(args, storage.Uint64ToPqInt64(*workspaceIDsIn))
	}

	var totalCount int
	if err := s.db.GetContext(ctx, &totalCount, countQuery, args...); err != nil {
		return nil, nil, err
	}

	query := `
		SELECT id, tenantid, workspaceid, name,
		       state, chartregistry, chartrepository, charttag,
		       valuesoverride, credentialssecretname, credentialspaths,
		       connectioninfotemplate, computedvalues,
		       status, statusmessage, connectioninfo, secretname,
		       createdat, updatedat
		FROM workspace_services
		WHERE tenantid = $1 AND deletedat IS NULL
	`
	if workspaceIDsIn != nil {
		query += " AND workspaceid = ANY($2)"
	}

	clause, validatedPagination := storage.BuildPaginationClause(pagination, model.WorkspaceSvc{})
	query += clause

	paginationRes := &common_model.PaginationResult{Total: uint64(totalCount)}
	if validatedPagination != nil {
		paginationRes.Limit = validatedPagination.Limit
		paginationRes.Offset = validatedPagination.Offset
		paginationRes.Sort = validatedPagination.Sort
	}

	var svcs []*model.WorkspaceSvc
	if err := s.db.SelectContext(ctx, &svcs, query, args...); err != nil {
		return nil, nil, err
	}

	return svcs, paginationRes, nil
}

func (s *WorkspaceStorage) ListWorkspaceSvcsByWorkspace(ctx context.Context, workspaceID uint64) ([]*model.WorkspaceSvc, error) {
	const query = `
		SELECT id, tenantid, workspaceid, name,
		       state, chartregistry, chartrepository, charttag,
		       valuesoverride, credentialssecretname, credentialspaths,
		       connectioninfotemplate, computedvalues,
		       status, statusmessage, connectioninfo, secretname,
		       createdat, updatedat
		FROM workspace_services
		WHERE workspaceid = $1 AND deletedat IS NULL;
	`

	var svcs []*model.WorkspaceSvc
	if err := s.db.SelectContext(ctx, &svcs, query, workspaceID); err != nil {
		return nil, err
	}

	return svcs, nil
}

func (s *WorkspaceStorage) CreateWorkspaceSvc(ctx context.Context, tenantID uint64, svc *model.WorkspaceSvc) (*model.WorkspaceSvc, error) {
	const query = `
		INSERT INTO workspace_services (tenantid, workspaceid, name,
		    state, chartregistry, chartrepository, charttag,
		    valuesoverride, credentialssecretname, credentialspaths,
		    connectioninfotemplate, computedvalues,
		    createdat, updatedat)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW(), NOW())
		RETURNING id, tenantid, workspaceid, name,
		    state, chartregistry, chartrepository, charttag,
		    valuesoverride, credentialssecretname, credentialspaths,
		    connectioninfotemplate, computedvalues,
		    status, statusmessage, connectioninfo, secretname,
		    createdat, updatedat;
	`

	valuesVal, err := svc.Values.Value()
	if err != nil {
		return nil, fmt.Errorf("unable to marshal values: %w", err)
	}
	computedVals, err := svc.ComputedValues.Value()
	if err != nil {
		return nil, fmt.Errorf("unable to marshal computed values: %w", err)
	}

	state := svc.State
	if state == "" {
		state = "Running"
	}
	credsPaths := svc.CredentialsPaths
	if credsPaths == nil {
		credsPaths = model.StringSlice{}
	}

	var created model.WorkspaceSvc
	err = s.db.GetContext(ctx, &created, query,
		tenantID, svc.WorkspaceID, svc.Name,
		state, svc.ChartRegistry, svc.ChartRepository, svc.ChartTag,
		valuesVal, svc.CredentialsSecretName, pqStringArray(credsPaths),
		svc.ConnectionInfoTemplate, computedVals,
	)
	if err != nil {
		return nil, err
	}

	return &created, nil
}

func (s *WorkspaceStorage) UpdateWorkspaceSvc(ctx context.Context, tenantID uint64, svc *model.WorkspaceSvc) (*model.WorkspaceSvc, error) {
	const query = `
		UPDATE workspace_services
		SET state = $3, chartregistry = $4, chartrepository = $5, charttag = $6,
		    valuesoverride = $7, credentialssecretname = $8, credentialspaths = $9,
		    connectioninfotemplate = $10, computedvalues = $11,
		    updatedat = NOW()
		WHERE tenantid = $1 AND id = $2 AND deletedat IS NULL
		RETURNING id, tenantid, workspaceid, name,
		    state, chartregistry, chartrepository, charttag,
		    valuesoverride, credentialssecretname, credentialspaths,
		    connectioninfotemplate, computedvalues,
		    status, statusmessage, connectioninfo, secretname,
		    createdat, updatedat;
	`

	valuesVal, err := svc.Values.Value()
	if err != nil {
		return nil, fmt.Errorf("unable to marshal values: %w", err)
	}
	computedVals, err := svc.ComputedValues.Value()
	if err != nil {
		return nil, fmt.Errorf("unable to marshal computed values: %w", err)
	}

	credsPaths := svc.CredentialsPaths
	if credsPaths == nil {
		credsPaths = model.StringSlice{}
	}

	var updated model.WorkspaceSvc
	err = s.db.GetContext(ctx, &updated, query,
		tenantID, svc.ID,
		svc.State, svc.ChartRegistry, svc.ChartRepository, svc.ChartTag,
		valuesVal, svc.CredentialsSecretName, pqStringArray(credsPaths),
		svc.ConnectionInfoTemplate, computedVals,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to update workspace service: %w", err)
	}

	return &updated, nil
}

func (s *WorkspaceStorage) DeleteWorkspaceSvc(ctx context.Context, tenantID, workspaceSvcID uint64) error {
	const query = `
		UPDATE workspace_services
		SET deletedat = NOW(), updatedat = NOW()
		WHERE tenantid = $1 AND id = $2 AND deletedat IS NULL;
	`

	rows, err := s.db.ExecContext(ctx, query, tenantID, workspaceSvcID)
	if err != nil {
		return fmt.Errorf("unable to delete workspace service: %w", err)
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
		          clipboard,
		          createdat, updatedat;
	`

	var deletedWorkspaces []*model.Workspace
	err := s.db.SelectContext(ctx, &deletedWorkspaces, query, model.WorkspaceDeleted.String(), "-"+uuid.Next(), fmt.Sprintf("%d seconds", int64(timeout.Seconds())))
	if err != nil {
		return nil, fmt.Errorf("unable to exec: %w", err)
	}

	return deletedWorkspaces, nil
}
