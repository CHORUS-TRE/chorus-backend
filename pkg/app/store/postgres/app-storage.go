package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/CHORUS-TRE/chorus-backend/internal/utils/database"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/uuid"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app/model"
	common "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	common_storage "github.com/CHORUS-TRE/chorus-backend/pkg/common/storage"
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
		SELECT id, tenantid, userid, "name", "description", "status", "dockerimagename", "dockerimagetag", "dockerimageregistry", "shmsize", "kioskconfigurl", "maxcpu", "mincpu", "maxmemory", "minmemory", "maxephemeralstorage", "minephemeralstorage", "iconurl", createdat, updatedat
		FROM apps
		WHERE tenantid = $1 AND id = $2;
	`

	var app model.App
	if err := s.db.GetContext(ctx, &app, query, tenantID, appID); err != nil {
		return nil, err
	}

	return &app, nil
}

func (s *AppStorage) ListApps(ctx context.Context, tenantID uint64, pagination *common.Pagination) ([]*model.App, *common.PaginationResult, error) {
	// Get total count query
	countQuery := `SELECT COUNT(*) FROM apps WHERE tenantid = $1 AND status != 'deleted';`
	var totalCount int
	if err := s.db.GetContext(ctx, &totalCount, countQuery, tenantID); err != nil {
		return nil, nil, err
	}

	// Get apps query
	query := `
		SELECT id, tenantid, userid, "name", "description", "status", "dockerimagename", "dockerimagetag", "dockerimageregistry", "shmsize", "kioskconfigurl", "maxcpu", "mincpu", "maxmemory", "minmemory", "maxephemeralstorage", "minephemeralstorage", "iconurl", createdat, updatedat
		FROM apps
		WHERE tenantid = $1 AND status != 'deleted'
	`

	// Add pagination
	clause, validatedPagination := common_storage.BuildPaginationClause(pagination, model.App{})
	query += clause

	// Build pagination result
	paginationRes := &common.PaginationResult{
		Total: uint64(totalCount),
	}

	if validatedPagination != nil {
		paginationRes.Limit = validatedPagination.Limit
		paginationRes.Offset = validatedPagination.Offset
		paginationRes.Sort = validatedPagination.Sort
	}

	var apps []*model.App
	if err := s.db.SelectContext(ctx, &apps, query, tenantID); err != nil {
		return nil, nil, err
	}

	return apps, paginationRes, nil
}

// CreateApp saves the provided app object in the database 'apps' table.
func (s *AppStorage) CreateApp(ctx context.Context, tenantID uint64, app *model.App) (*model.App, error) {
	const appQuery = `
		INSERT INTO apps (tenantid, userid, "name", "description", "status", "dockerimagename", "dockerimagetag", "dockerimageregistry", "shmsize", "kioskconfigurl", "maxcpu", "mincpu", "maxmemory", "minmemory", "maxephemeralstorage", "minephemeralstorage", "iconurl", createdat, updatedat)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, NOW(), NOW()) 
		RETURNING id, tenantid, userid, "name", "description", "status", "dockerimagename", "dockerimagetag", "dockerimageregistry", "shmsize", "kioskconfigurl", "maxcpu", "mincpu", "maxmemory", "minmemory", "maxephemeralstorage", "minephemeralstorage", "iconurl", createdat, updatedat;
	`

	var newApp model.App
	err := s.db.GetContext(ctx, &newApp, appQuery,
		tenantID, app.UserID, app.Name, app.Description, app.Status, app.DockerImageName, app.DockerImageTag, app.DockerImageRegistry, app.ShmSize, app.KioskConfigURL, app.MaxCPU, app.MinCPU, app.MaxMemory, app.MinMemory, app.MaxEphemeralStorage, app.MinEphemeralStorage, app.IconURL,
	)
	if err != nil {
		return nil, err
	}

	return &newApp, nil
}

func (s *AppStorage) BulkCreateApps(ctx context.Context, tenantID uint64, apps []*model.App) ([]*model.App, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("unable to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			_ = tx.Rollback() // err is non-nil; don't change it
		}
	}()

	const appQuery = `
		INSERT INTO apps (tenantid, userid, name, description, status, dockerimagename, dockerimagetag, dockerimageregistry, shmsize, kioskconfigurl, maxcpu, mincpu, maxmemory, minmemory, maxephemeralstorage, minephemeralstorage, iconurl, createdat, updatedat)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, NOW(), NOW()) 
		RETURNING id, tenantid, userid, name, description, status, dockerimagename, dockerimagetag, dockerimageregistry, shmsize, kioskconfigurl, maxcpu, mincpu, maxmemory, minmemory, maxephemeralstorage, minephemeralstorage, iconurl, createdat, updatedat;
	`

	var newApps []*model.App
	for _, app := range apps {
		var newApp model.App
		err = tx.GetContext(ctx, &newApp, appQuery,
			tenantID, app.UserID, app.Name, app.Description, app.Status, app.DockerImageName, app.DockerImageTag, app.DockerImageRegistry, app.ShmSize, app.KioskConfigURL, app.MaxCPU, app.MinCPU, app.MaxMemory, app.MinMemory, app.MaxEphemeralStorage, app.MinEphemeralStorage, app.IconURL,
		)
		if err != nil {
			return nil, fmt.Errorf("unable to exec bulk insert: %w", err)
		}
		newApps = append(newApps, &newApp)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("unable to commit transaction: %w", err)
	}

	return newApps, nil
}

func (s *AppStorage) UpdateApp(ctx context.Context, tenantID uint64, app *model.App) (*model.App, error) {
	const appUpdateQuery = `
		UPDATE apps
		SET name = $3, description = $4, status = $5, dockerimagename = $6, dockerimagetag = $7, dockerimageregistry = $8, shmsize = $9, kioskconfigurl = $10, maxcpu = $11, mincpu = $12, maxmemory = $13, minmemory = $14, maxephemeralstorage = $15, minephemeralstorage = $16, iconurl = $17,
		updatedat = NOW()
		WHERE tenantid = $1 AND id = $2
		RETURNING id, tenantid, userid, "name", "description", "status", "dockerimagename", "dockerimagetag", "dockerimageregistry", "shmsize", "kioskconfigurl", "maxcpu", "mincpu", "maxmemory", "minmemory", "maxephemeralstorage", "minephemeralstorage", "iconurl", createdat, updatedat;
	`

	// Update app
	var updatedApp model.App
	err := s.db.GetContext(ctx, &updatedApp, appUpdateQuery, tenantID, app.ID, app.Name, app.Description, app.Status, app.DockerImageName, app.DockerImageTag, app.DockerImageRegistry, app.ShmSize, app.KioskConfigURL, app.MaxCPU, app.MinCPU, app.MaxMemory, app.MinMemory, app.MaxEphemeralStorage, app.MinEphemeralStorage, app.IconURL)
	if err != nil {
		return nil, err
	}

	return &updatedApp, nil
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
