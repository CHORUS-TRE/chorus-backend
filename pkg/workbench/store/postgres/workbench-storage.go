package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/CHORUS-TRE/chorus-backend/internal/utils/database"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/uuid"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/common/storage"
	common_storage "github.com/CHORUS-TRE/chorus-backend/pkg/common/storage"
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
		SELECT id, tenantid, userid, workspaceid, name, shortname, description, status, serverpodstatus, k8sstatus, initialresolutionwidth, initialresolutionheight, createdat, updatedat
			FROM workbenchs
		WHERE tenantid = $1 AND id = $2 AND deletedat IS NULL;
	`

	var workbench model.Workbench
	if err := s.db.GetContext(ctx, &workbench, query, tenantID, workbenchID); err != nil {
		return nil, err
	}

	return &workbench, nil
}

func (s *WorkbenchStorage) ListWorkbenchs(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, workspaceIDsIn *[]uint64) ([]*model.Workbench, *common_model.PaginationResult, error) {
	args := []interface{}{tenantID}

	countQuery := `SELECT COUNT(*) FROM workbenchs WHERE tenantid = $1 AND status != 'deleted' AND deletedat IS NULL`
	if workspaceIDsIn != nil {
		countQuery += " AND workspaceid = ANY($2)"
		args = append(args, storage.Uint64ToPqInt64(*workspaceIDsIn))
	}
	var totalCount int64
	if err := s.db.GetContext(ctx, &totalCount, countQuery, args...); err != nil {
		return nil, nil, err
	}

	// Get workbenches query
	query := `
		SELECT id, tenantid, userid, workspaceid, name, shortname, description, status, serverpodstatus, k8sstatus, initialresolutionwidth, initialresolutionheight, createdat, updatedat
		FROM workbenchs
		WHERE tenantid = $1 AND status != 'deleted' AND deletedat IS NULL
	`
	if workspaceIDsIn != nil {
		query += " AND workspaceid = ANY($2) "
	}

	// Add pagination
	clause, validatedPagination := common_storage.BuildPaginationClause(pagination, model.Workbench{})
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

	var workbenchs []*model.Workbench
	if err := s.db.SelectContext(ctx, &workbenchs, query, args...); err != nil {
		return nil, nil, err
	}

	return workbenchs, paginationRes, nil
}

func (s *WorkbenchStorage) DeleteIdleWorkbenchs(ctx context.Context, idleTimeout time.Duration) ([]*model.Workbench, error) {
	const query = `
		UPDATE workbenchs
		SET (status, name, updatedat, deletedat) = ($1, concat(name, $2::TEXT), NOW(), NOW())
		WHERE accessedat IS NOT NULL
		  AND accessedat < NOW() - INTERVAL $3 * INTERVAL '1 second'
		  AND status != 'deleted'
		  AND deletedat IS NULL
		RETURNING id, tenantid, userid, workspaceid, name, shortname, description, status, serverpodstatus, k8sstatus, initialresolutionwidth, initialresolutionheight, createdat, updatedat;
	`

	var deletedWorkbenchs []*model.Workbench
	err := s.db.SelectContext(ctx, &deletedWorkbenchs, query, model.WorkbenchDeleted.String(), "-"+uuid.Next(), int64(idleTimeout.Seconds()))
	if err != nil {
		return nil, fmt.Errorf("unable to exec: %w", err)
	}

	return deletedWorkbenchs, nil
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
	ai.kioskconfigjwttoken,

	a.name as AppName,
    a.dockerimageregistry as AppDockerImageRegistry,
    a.dockerimagename as AppDockerImageName,
    a.dockerimagetag as AppDockerImageTag,
	a.shmsize as AppShmSize,
	a.kioskconfigurl as AppKioskConfigURL,
	a.kioskconfigjwturl as AppKioskConfigJWTUrl,
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
	AND ai.deletedat IS NULL
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
SELECT id, tenantid, userid, workspaceid, name, shortname, description, status, serverpodstatus, k8sstatus, initialresolutionwidth, initialresolutionheight, createdat, updatedat
	FROM workbenchs
WHERE deletedat IS NULL;
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
func (s *WorkbenchStorage) CreateWorkbench(ctx context.Context, tenantID uint64, workbench *model.Workbench) (*model.Workbench, error) {
	const workbenchQuery = `
		INSERT INTO workbenchs (tenantid, userid, workspaceid, name, shortname, description, initialresolutionwidth, initialresolutionheight, status, serverpodstatus, k8sstatus, createdat, updatedat)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW()) 
		RETURNING id, tenantid, userid, workspaceid, name, shortname, description, status, serverpodstatus, k8sstatus, initialresolutionwidth, initialresolutionheight, createdat, updatedat;
	`

	var newWorkbench model.Workbench
	err := s.db.GetContext(ctx, &newWorkbench, workbenchQuery,
		tenantID, workbench.UserID, workbench.WorkspaceID, workbench.Name, workbench.ShortName, workbench.Description, workbench.InitialResolutionWidth, workbench.InitialResolutionHeight, workbench.Status, workbench.ServerPodStatus, workbench.K8sStatus,
	)
	if err != nil {
		return nil, err
	}

	return &newWorkbench, nil
}

func (s *WorkbenchStorage) UpdateWorkbench(ctx context.Context, tenantID uint64, workbench *model.Workbench) (*model.Workbench, error) {
	const workbenchUpdateQuery = `
		UPDATE workbenchs
		SET status = $3, serverpodstatus = $4, k8sstatus = $5, description = $6, updatedat = NOW()
		WHERE tenantid = $1 AND id = $2
		RETURNING id, tenantid, userid, workspaceid, name, shortname, description, status, serverpodstatus, k8sstatus, initialresolutionwidth, initialresolutionheight, createdat, updatedat;
	`

	// Update workbench
	var updatedWorkbench model.Workbench
	err := s.db.GetContext(ctx, &updatedWorkbench, workbenchUpdateQuery, tenantID, workbench.ID, workbench.Status, workbench.ServerPodStatus, workbench.K8sStatus, workbench.Description)
	if err != nil {
		return nil, err
	}

	return &updatedWorkbench, nil
}

func (s *WorkbenchStorage) DeleteWorkbench(ctx context.Context, tenantID uint64, workbenchID uint64) error {
	// First delete app instances in workbench
	if err := s.DeleteAppInstancesInWorkbench(ctx, tenantID, workbenchID); err != nil {
		return fmt.Errorf("unable to delete app instances in workbench %v: %w", workbenchID, err)
	}

	const query = `
		UPDATE workbenchs
		SET (status, name, updatedat, deletedat) = ($3, concat(name, $4::TEXT), NOW(), NOW())
		WHERE tenantid = $1 AND id = $2 AND deletedat IS NULL;
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

func (s *WorkbenchStorage) DeleteWorkbenchsInWorkspace(ctx context.Context, tenantID uint64, workspaceID uint64) error {
	// TODO: batch
	const query = `
		SELECT id FROM workbenchs
		WHERE tenantid = $1 AND workspaceid = $2 AND status != 'deleted' AND deletedat IS NULL;
	`

	var workbenchIDs []uint64
	if err := s.db.SelectContext(ctx, &workbenchIDs, query, tenantID, workspaceID); err != nil {
		return fmt.Errorf("unable to select workbenchs: %w", err)
	}

	// Delete workbenchs
	for _, workbenchID := range workbenchIDs {
		if err := s.DeleteWorkbench(ctx, tenantID, workbenchID); err != nil {
			return fmt.Errorf("unable to delete workbench %v: %w", workbenchID, err)
		}
	}

	return nil
}

func (s *WorkbenchStorage) GetAppInstance(ctx context.Context, tenantID uint64, appInstanceID uint64) (*model.AppInstance, error) {
	const query = `
		SELECT id, tenantid, userid, appid, workspaceid, workbenchid, status, initialresolutionwidth, initialresolutionheight, kioskconfigjwttoken, createdat, updatedat
			FROM app_instances
		WHERE tenantid = $1 AND id = $2 AND deletedat IS NULL;
	`

	var appInstance model.AppInstance
	if err := s.db.GetContext(ctx, &appInstance, query, tenantID, appInstanceID); err != nil {
		return nil, err
	}

	return &appInstance, nil
}

func (s *WorkbenchStorage) ListAppInstances(ctx context.Context, tenantID uint64, pagination *common_model.Pagination) ([]*model.AppInstance, *common_model.PaginationResult, error) {
	countQuery := `SELECT COUNT(*) FROM app_instances WHERE tenantid = $1 AND status != 'deleted' AND deletedat IS NULL;`
	var totalCount int64
	if err := s.db.GetContext(ctx, &totalCount, countQuery, tenantID); err != nil {
		return nil, nil, err
	}

	// Get app instances query
	query := `
		SELECT id, tenantid, userid, appid, workspaceid, workbenchid, status, initialresolutionwidth, initialresolutionheight, kioskconfigjwttoken, createdat, updatedat
		FROM app_instances
		WHERE tenantid = $1 AND status != 'deleted' AND deletedat IS NULL
	`

	// Add pagination
	clause, validatedPagination := common_storage.BuildPaginationClause(pagination, model.AppInstance{})
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

	var appInstances []*model.AppInstance
	if err := s.db.SelectContext(ctx, &appInstances, query, tenantID); err != nil {
		return nil, nil, err
	}

	return appInstances, paginationRes, nil
}

// CreateAppInstance saves the provided appInstance object in the database 'appInstances' table.
func (s *WorkbenchStorage) CreateAppInstance(ctx context.Context, tenantID uint64, appInstance *model.AppInstance) (*model.AppInstance, error) {
	const appInstanceQuery = `
		INSERT INTO app_instances (tenantid, userid, appid, workspaceid, workbenchid, status, initialresolutionwidth, initialresolutionheight, kioskconfigjwttoken, createdat, updatedat)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id, tenantid, userid, appid, workspaceid, workbenchid, status, initialresolutionwidth, initialresolutionheight, kioskconfigjwttoken, createdat, updatedat;
	`

	var newAppInstance model.AppInstance
	err := s.db.GetContext(ctx, &newAppInstance, appInstanceQuery,
		tenantID, appInstance.UserID, appInstance.AppID, appInstance.WorkspaceID, appInstance.WorkbenchID, appInstance.Status, appInstance.InitialResolutionWidth, appInstance.InitialResolutionHeight, appInstance.KioskConfigJWTToken,
	)
	if err != nil {
		return nil, err
	}

	return &newAppInstance, nil
}

func (s *WorkbenchStorage) UpdateAppInstance(ctx context.Context, tenantID uint64, appInstance *model.AppInstance) (*model.AppInstance, error) {
	const appInstanceUpdateQuery = `
		UPDATE app_instances
		SET status = $3, k8sstate = $4, k8sstatus = $5, updatedat = NOW()
		WHERE tenantid = $1 AND id = $2
		RETURNING id, tenantid, userid, appid, workspaceid, workbenchid, status, k8sstate, k8sstatus, initialresolutionwidth, initialresolutionheight, kioskconfigjwttoken, createdat, updatedat;
	`

	var updatedAppInstance model.AppInstance
	err := s.db.GetContext(ctx, &updatedAppInstance, appInstanceUpdateQuery, tenantID, appInstance.ID, appInstance.Status, appInstance.K8sState, appInstance.K8sStatus)
	if err != nil {
		return nil, err
	}

	return &updatedAppInstance, nil
}

func (s *WorkbenchStorage) UpdateAppInstances(ctx context.Context, tenantID uint64, appInstances []*model.AppInstance) (err error) {
	var errAcc []error
	for _, appInstance := range appInstances {
		if _, err := s.UpdateAppInstance(ctx, tenantID, appInstance); err != nil {
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
		WHERE tenantid = $1 AND id = $2 AND deletedat IS NULL;
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

func (s *WorkbenchStorage) DeleteAppInstances(ctx context.Context, tenantID uint64, appInstanceIDs []uint64) error {
	var errAcc []error
	for _, appInstanceID := range appInstanceIDs {
		if err := s.DeleteAppInstance(ctx, tenantID, appInstanceID); err != nil {
			errAcc = append(errAcc, fmt.Errorf("failed to delete appInstance %d: %w", appInstanceID, err))
		}
	}
	return errors.Join(errAcc...)
}

func (s *WorkbenchStorage) DeleteAppInstancesInWorkbench(ctx context.Context, tenantID uint64, workbenchID uint64) error {
	const query = `
		UPDATE app_instances
		SET (status, updatedat, deletedat) = ($3, NOW(), NOW())
		WHERE tenantid = $1 AND workbenchid = $2 AND deletedat IS NULL;
	`

	_, err := s.db.ExecContext(ctx, query, tenantID, workbenchID, model.AppInstanceDeleted.String())
	if err != nil {
		return fmt.Errorf("unable to exec: %w", err)
	}

	return nil
}
