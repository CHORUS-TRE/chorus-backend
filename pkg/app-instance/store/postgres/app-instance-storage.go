package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/CHORUS-TRE/chorus-backend/internal/utils/database"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app-instance/model"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
)

// AppInstanceStorage is the handler through which a PostgresDB backend can be queried.
type AppInstanceStorage struct {
	db *sqlx.DB
}

// NewAppInstanceStorage returns a fresh appInstance service storage instance.
func NewAppInstanceStorage(db *sqlx.DB) *AppInstanceStorage {
	return &AppInstanceStorage{db: db}
}

func (s *AppInstanceStorage) GetAppInstance(ctx context.Context, tenantID uint64, appInstanceID uint64) (*model.AppInstance, error) {
	const query = `
		SELECT id, tenantid, userid, appid, workspaceid, workbenchid, status, createdat, updatedat
			FROM app_instances
		WHERE tenantid = $1 AND id = $2;
	`

	var appInstance model.AppInstance
	if err := s.db.GetContext(ctx, &appInstance, query, tenantID, appInstanceID); err != nil {
		return nil, err
	}

	return &appInstance, nil
}

func (s *AppInstanceStorage) ListAppInstances(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.AppInstance, error) {
	const query = `
SELECT id, tenantid, userid, appid, workspaceid, workbenchid, status, createdat, updatedat
	FROM app_instances
WHERE tenantid = $1 AND status != 'deleted';
`
	var appInstances []*model.AppInstance
	if err := s.db.SelectContext(ctx, &appInstances, query, tenantID); err != nil {
		return nil, err
	}

	return appInstances, nil
}

// CreateAppInstance saves the provided appInstance object in the database 'appInstances' table.
func (s *AppInstanceStorage) CreateAppInstance(ctx context.Context, tenantID uint64, appInstance *model.AppInstance) (uint64, error) {
	const appInstanceQuery = `
INSERT INTO app_instances (tenantid, userid, appid, workspaceid, workbenchid, status, createdat, updatedat)
VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW()) RETURNING id;
	`

	var id uint64
	err := s.db.GetContext(ctx, &id, appInstanceQuery,
		tenantID, appInstance.UserID, appInstance.AppID, appInstance.WorkspaceID, appInstance.WorkbenchID, appInstance.Status,
	)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *AppInstanceStorage) UpdateAppInstance(ctx context.Context, tenantID uint64, appInstance *model.AppInstance) (err error) {
	const appInstanceUpdateQuery = `
		UPDATE app_instances
		SET status = $3, updatedat = NOW()
		WHERE tenantid = $1 AND id = $2;
	`

	// Update User
	rows, err := s.db.ExecContext(ctx, appInstanceUpdateQuery, tenantID, appInstance.ID, appInstance.Status)
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

func (s *AppInstanceStorage) DeleteAppInstance(ctx context.Context, tenantID uint64, appInstanceID uint64) error {
	const query = `
		UPDATE app_instances SET 
			(status, updatedat, deletedat) = 
			($3, NOW(), NOW())
		WHERE tenantid = $1 AND id = $2;
	`

	rows, err := s.db.ExecContext(ctx, query, tenantID, appInstanceID, model.AppInstanceDeleted.String())
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
