package postgres

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/utils/database"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/uuid"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
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
		WHERE tenantid = $1 AND id = $2;
	`

	var workspace model.Workspace
	if err := s.db.GetContext(ctx, &workspace, query, tenantID, workspaceID); err != nil {
		return nil, err
	}

	return &workspace, nil
}

func (s *WorkspaceStorage) ListWorkspaces(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.Workspace, error) {
	const query = `
SELECT id, tenantid, userid, name, shortname, description, status, createdat, updatedat
	FROM workspaces
WHERE tenantid = $1 AND status != 'deleted';
`
	var workspaces []*model.Workspace
	if err := s.db.SelectContext(ctx, &workspaces, query, tenantID); err != nil {
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
		SET status = $3, description = $4, updatedat = NOW()
		WHERE tenantid = $1 AND id = $2;
	`

	// Update User
	rows, err := s.db.ExecContext(ctx, workspaceUpdateQuery, tenantID, workspace.ID, workspace.Status, workspace.Description)
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
		WHERE tenantid = $1 AND id = $2;
	`

	rows, err := s.db.ExecContext(ctx, query, tenantID, workspaceID, model.WorkspaceDeleted.String(), "-"+uuid.Next())
	if err != nil {
		return errors.Wrap(err, "unable to exec")
	}
	affected, err := rows.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "unable to get rows affected")
	}

	if affected == 0 {
		return database.ErrNoRowsDeleted
	}

	return nil
}
