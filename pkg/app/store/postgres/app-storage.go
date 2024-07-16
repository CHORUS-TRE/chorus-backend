package postgres

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/utils/database"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/uuid"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app/model"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// AppStorage is the handler through which a PostgresDB backend can be queried.
type AppStorage struct {
	db *sqlx.DB
}

// NewAppStorage returns a fresh app service storage instance.
func NewAppStorage(db *sqlx.DB) *AppStorage {
	return &AppStorage{db: db}
}

func (s *AppStorage) GetApp(ctx context.Context, tenantID uint64, appID uint64) (*model.App, error) {
	const query = `
		SELECT id, tenantid, userid, name, description, status, dockerimagename, dockerimagetag, createdat, updatedat
			FROM apps
		WHERE tenantid = $1 AND id = $2;
	`

	var app model.App
	if err := s.db.GetContext(ctx, &app, query, tenantID, appID); err != nil {
		return nil, err
	}

	return &app, nil
}

func (s *AppStorage) ListApps(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.App, error) {
	const query = `
SELECT id, tenantid, userid, name, description, status, dockerimagename, dockerimagetag, createdat, updatedat
	FROM apps
WHERE tenantid = $1 AND status != 'deleted';
`
	var apps []*model.App
	if err := s.db.SelectContext(ctx, &apps, query, tenantID); err != nil {
		return nil, err
	}

	return apps, nil
}

// CreateApp saves the provided app object in the database 'apps' table.
func (s *AppStorage) CreateApp(ctx context.Context, tenantID uint64, app *model.App) (uint64, error) {
	const appQuery = `
INSERT INTO apps (tenantid, userid, name, description, status, dockerimagename, dockerimagetag, createdat, updatedat)
VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW()) RETURNING id;
	`

	var id uint64
	err := s.db.GetContext(ctx, &id, appQuery,
		tenantID, app.UserID, app.Name, app.Description, app.Status, app.DockerImageName, app.DockerImageTag,
	)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *AppStorage) UpdateApp(ctx context.Context, tenantID uint64, app *model.App) (err error) {
	const appUpdateQuery = `
		UPDATE apps
		SET name = $3, description = $4, status = $5, updatedat = NOW()
		WHERE tenantid = $1 AND id = $2;
	`

	// Update User
	rows, err := s.db.ExecContext(ctx, appUpdateQuery, tenantID, app.ID, app.Name, app.Description, app.Status)
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

func (s *AppStorage) DeleteApp(ctx context.Context, tenantID uint64, appID uint64) error {
	const query = `
		UPDATE apps	SET 
			(status, name, updatedat, deletedat) = 
			($3, concat(name, $4::TEXT), NOW(), NOW())
		WHERE tenantid = $1 AND id = $2;
	`

	rows, err := s.db.ExecContext(ctx, query, tenantID, appID, model.AppDeleted.String(), "-"+uuid.Next())
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
