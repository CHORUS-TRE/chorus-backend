package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/CHORUS-TRE/chorus-backend/internal/utils/database"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/uuid"
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

func (s *WorkbenchStorage) ListWorkbenchAppInstances(ctx context.Context, workbenchID uint64) ([]*model.AppInstance, error) {
	const query = `
SELECT 
    ai.id,
    ai.tenantid,
    ai.userid,
    ai.appid,
    ai.workspaceid,
    ai.workbenchid,
    ai.status,
	ai.initialresolutionwidth,
	ai.initialresolutionheight,
    ai.createdat,
    ai.updatedat,

	a.name as AppName,
    a.dockerimageregistry as AppDockerImageRegistry,
    a.dockerimagename as AppDockerImageName,
    a.dockerimagetag as AppDockerImageTag,
	a.shmsize as AppShmSize,
	a.kioskconfigurl as AppKioskConfigURL,
	a.maxcpu as AppMaxCPU,
	a.mincpu as AppMinCPU,
	a.maxmemory as AppMaxMemory,
	a.minmemory as AppMinMemory,
	a.maxephemeralstorage as AppMaxEphemeralStorage,
	a.minephemeralstorage as AppMinEphemeralStorage,
	a.iconurl as AppIconURL

FROM 
    app_instances ai
JOIN 
    apps a ON ai.appid = a.id
WHERE 
    ai.workbenchid = $1 
    AND ai.status != 'deleted'
ORDER BY ai.createdat ASC;
;
`
	var appInstances []*model.AppInstance
	if err := s.db.SelectContext(ctx, &appInstances, query, workbenchID); err != nil {
		return nil, err
	}

	return appInstances, nil
}

func (s *WorkbenchStorage) ListAllWorkbenches(ctx context.Context) ([]*model.Workbench, error) {
	const query = `
SELECT id, tenantid, userid, workspaceid, name, shortname, description, status, createdat, updatedat
	FROM workbenchs;
`
	var workbenchs []*model.Workbench
	if err := s.db.SelectContext(ctx, &workbenchs, query); err != nil {
		return nil, err
	}

	return workbenchs, nil
}

func (s *WorkbenchStorage) SaveBatchProxyHit(ctx context.Context, proxyHitCountMap map[uint64]uint64, proxyHitDateMap map[uint64]time.Time) error {
	const query = `
UPDATE public.workbenchs
SET 
    accessedat = batch_data.date,
    accessedcount = accessedcount + batch_data.count,
	updatedat = NOW()
FROM (
    SELECT UNNEST($1::BIGINT[]) AS id, UNNEST($2::TIMESTAMP[]) AS date, UNNEST($3::BIGINT[]) AS count
) AS batch_data
WHERE workbenchs.id = batch_data.id
;
`
	ids := make([]uint64, 0, len(proxyHitCountMap))
	dates := make([]string, 0, len(proxyHitCountMap))
	counts := make([]uint64, 0, len(proxyHitCountMap))

	for id, count := range proxyHitCountMap {

		ids = append(ids, id)
		dates = append(dates, proxyHitDateMap[id].Format(time.RFC3339))
		counts = append(counts, count)
	}

	_, err := s.db.ExecContext(ctx, query, pq.Array(ids), pq.Array(dates), pq.Array(counts))
	return err
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

func (s *WorkbenchStorage) GetAppInstance(ctx context.Context, tenantID uint64, appInstanceID uint64) (*model.AppInstance, error) {
	const query = `
		SELECT id, tenantid, userid, appid, workspaceid, workbenchid, status, initialresolutionwidth, initialresolutionheight, createdat, updatedat
			FROM app_instances
		WHERE tenantid = $1 AND id = $2;
	`

	var appInstance model.AppInstance
	if err := s.db.GetContext(ctx, &appInstance, query, tenantID, appInstanceID); err != nil {
		return nil, err
	}

	return &appInstance, nil
}

func (s *WorkbenchStorage) ListAppInstances(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.AppInstance, error) {
	const query = `
SELECT id, tenantid, userid, appid, workspaceid, workbenchid, status, initialresolutionwidth, initialresolutionheight, createdat, updatedat
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
func (s *WorkbenchStorage) CreateAppInstance(ctx context.Context, tenantID uint64, appInstance *model.AppInstance) (uint64, error) {
	const appInstanceQuery = `
INSERT INTO app_instances (tenantid, userid, appid, workspaceid, workbenchid, status, initialresolutionwidth, initialresolutionheight, createdat, updatedat)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW()) RETURNING id;
	`

	var id uint64
	err := s.db.GetContext(ctx, &id, appInstanceQuery,
		tenantID, appInstance.UserID, appInstance.AppID, appInstance.WorkspaceID, appInstance.WorkbenchID, appInstance.Status, appInstance.InitialResolutionWidth, appInstance.InitialResolutionHeight,
	)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *WorkbenchStorage) UpdateAppInstance(ctx context.Context, tenantID uint64, appInstance *model.AppInstance) (err error) {
	const appInstanceUpdateQuery = `
		UPDATE app_instances
		SET status = $3, k8sstate = $4, k8sstatus = $5, updatedat = NOW()
		WHERE tenantid = $1 AND id = $2;
	`

	rows, err := s.db.ExecContext(ctx, appInstanceUpdateQuery, tenantID, appInstance.ID, appInstance.Status, appInstance.K8sState, appInstance.K8sStatus)
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

func (s *WorkbenchStorage) UpdateAppInstances(ctx context.Context, tenantID uint64, appInstances []*model.AppInstance) (err error) {
	var errAcc []error
	for _, appInstance := range appInstances {
		if err := s.UpdateAppInstance(ctx, tenantID, appInstance); err != nil {
			errAcc = append(errAcc, fmt.Errorf("failed to update appInstance %d: %w", appInstance.ID, err))
		}
	}
	return errors.Join(errAcc...)
}

func (s *WorkbenchStorage) DeleteAppInstance(ctx context.Context, tenantID uint64, appInstanceID uint64) error {
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
