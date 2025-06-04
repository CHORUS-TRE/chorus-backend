package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"

	"github.com/CHORUS-TRE/chorus-backend/internal/utils/database"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/uuid"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
)

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
		SELECT id, tenantid, userid, name, shortname, description, status, createdat, updatedat
			FROM workspaces
		WHERE tenantid = $1 AND id = $2 AND deletedat IS NULL;
	`

	var workspace model.Workspace
	if err := s.db.GetContext(ctx, &workspace, query, tenantID, workspaceID); err != nil {
		return nil, err
	}

	return &workspace, nil
}

func (s *WorkspaceStorage) ListWorkspaces(ctx context.Context, tenantID uint64, pagination common_model.Pagination, allowDeleted bool) ([]*model.Workspace, error) {
	query := `
SELECT id, tenantid, userid, name, shortname, description, status, createdat, updatedat
	FROM workspaces
`

	conditions := []string{}
	arguments := []interface{}{}

	if tenantID != 0 {
		conditions = append(conditions, "tenantid = $1")
		arguments = append(arguments, tenantID)
	}

	if !allowDeleted {
		conditions = append(conditions, "status != 'deleted'", "deletedat IS NULL")
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var workspaces []*model.Workspace
	if err := s.db.SelectContext(ctx, &workspaces, query, arguments...); err != nil {
		return nil, err
	}

	return workspaces, nil
}

// CreateWorkspace saves the provided workspace object in the database 'workspaces' table.
func (s *WorkspaceStorage) CreateWorkspace(ctx context.Context, tenantID uint64, workspace *model.Workspace) (uint64, error) {
	const workspaceQuery = `
INSERT INTO workspaces (tenantid, userid, name, shortname, description, status, createdat, updatedat)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW()) RETURNING id;
	`

	var id uint64
	err := s.db.GetContext(ctx, &id, workspaceQuery,
		tenantID, workspace.UserID, workspace.Name, workspace.ShortName, workspace.Description, workspace.Status,
	)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *WorkspaceStorage) UpdateWorkspace(ctx context.Context, tenantID uint64, workspace *model.Workspace) (err error) {
	const workspaceUpdateQuery = `
		UPDATE workspaces
		SET name = $3, shortname = $4, description = $5, status = $6, updatedat = NOW()
		WHERE tenantid = $1 AND id = $2 AND deletedat IS NULL;
	`

	// Update User
	rows, err := s.db.ExecContext(ctx, workspaceUpdateQuery, tenantID, workspace.ID, workspace.Name, workspace.ShortName, workspace.Description, workspace.Status)
	if err != nil {
		return err
	}
	affected, err := rows.RowsAffected()
	if err != nil {
		return err
	}

	if affected == 0 {
		return database.ErrNoRowsUpdated
	}

	return err
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
		return database.ErrNoRowsDeleted
	}

	return nil
}
