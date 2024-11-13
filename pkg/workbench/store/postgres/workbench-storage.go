package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/CHORUS-TRE/chorus-backend/internal/utils/database"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/uuid"
	app_instance_model "github.com/CHORUS-TRE/chorus-backend/pkg/app-instance/model"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/model"
)

// WorkbenchStorage is the handler through which a PostgresDB backend can be queried.
type WorkbenchStorage struct {
	db *sqlx.DB
}

// NewWorkbenchStorage returns a fresh workbench service storage instance.
func NewWorkbenchStorage(db *sqlx.DB) *WorkbenchStorage {
	return &WorkbenchStorage{db: db}
}

func (s *WorkbenchStorage) GetWorkbench(ctx context.Context, tenantID uint64, workbenchID uint64) (*model.Workbench, error) {
	const query = `
		SELECT id, tenantid, userid, workspaceid, name, shortname, description, status, createdat, updatedat
			FROM workbenchs
		WHERE tenantid = $1 AND id = $2;
	`

	var workbench model.Workbench
	if err := s.db.GetContext(ctx, &workbench, query, tenantID, workbenchID); err != nil {
		return nil, err
	}

	return &workbench, nil
}

func (s *WorkbenchStorage) ListWorkbenchs(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.Workbench, error) {
	const query = `
SELECT id, tenantid, userid, workspaceid, name, shortname, description, status, createdat, updatedat
	FROM workbenchs
WHERE tenantid = $1 AND status != 'deleted';
`
	var workbenchs []*model.Workbench
	if err := s.db.SelectContext(ctx, &workbenchs, query, tenantID); err != nil {
		return nil, err
	}

	return workbenchs, nil
}

func (s *WorkbenchStorage) ListWorkbenchAppInstances(ctx context.Context, workbenchID uint64) ([]*app_instance_model.AppInstance, error) {
	const query = `
SELECT 
    ai.id,
    ai.tenantid,
    ai.userid,
    ai.appid,
    ai.workspaceid,
    ai.workbenchid,
    ai.status,
    ai.createdat,
    ai.updatedat,
	a.name as AppName,
    a.dockerimagename as AppDockerImageName,
    a.dockerimagetag as AppDockerImageTag
FROM 
    app_instances ai
JOIN 
    apps a ON ai.appid = a.id
WHERE 
    ai.workbenchid = $1 
    AND ai.status != 'deleted';
;
`
	var appInstances []*app_instance_model.AppInstance
	if err := s.db.SelectContext(ctx, &appInstances, query, workbenchID); err != nil {
		return nil, err
	}

	return appInstances, nil
}

func (s *WorkbenchStorage) ListAllActiveWorkbenchs(ctx context.Context) ([]*model.Workbench, error) {
	const query = `
SELECT id, tenantid, userid, workspaceid, name, shortname, description, status, createdat, updatedat
	FROM workbenchs
WHERE status = 'active';
`
	var workbenchs []*model.Workbench
	if err := s.db.SelectContext(ctx, &workbenchs, query); err != nil {
		return nil, err
	}

	return workbenchs, nil
}

// CreateWorkbench saves the provided workbench object in the database 'workbenchs' table.
func (s *WorkbenchStorage) CreateWorkbench(ctx context.Context, tenantID uint64, workbench *model.Workbench) (uint64, error) {
	const workbenchQuery = `
INSERT INTO workbenchs (tenantid, userid, workspaceid, name, shortname, description, status, createdat, updatedat)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW()) RETURNING id;
	`

	var id uint64
	err := s.db.GetContext(ctx, &id, workbenchQuery,
		tenantID, workbench.UserID, workbench.WorkspaceID, workbench.Name, workbench.ShortName, workbench.Description, workbench.Status,
	)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *WorkbenchStorage) UpdateWorkbench(ctx context.Context, tenantID uint64, workbench *model.Workbench) (err error) {
	const workbenchUpdateQuery = `
		UPDATE workbenchs
		SET status = $3, description = $4, updatedat = NOW()
		WHERE tenantid = $1 AND id = $2;
	`

	// Update User
	rows, err := s.db.ExecContext(ctx, workbenchUpdateQuery, tenantID, workbench.ID, workbench.Status, workbench.Description)
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

func (s *WorkbenchStorage) DeleteWorkbench(ctx context.Context, tenantID uint64, workbenchID uint64) error {
	const query = `
		UPDATE workbenchs	SET 
			(status, name, updatedat, deletedat) = 
			($3, concat(name, $4::TEXT), NOW(), NOW())
		WHERE tenantid = $1 AND id = $2;
	`

	rows, err := s.db.ExecContext(ctx, query, tenantID, workbenchID, model.WorkbenchDeleted.String(), "-"+uuid.Next())
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
