package middleware

import (
	"context"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/service"

	"go.uber.org/zap"
)

type workspaceStorageLogging struct {
	logger *logger.ContextLogger
	next   service.WorkspaceStore
}

func Logging(logger *logger.ContextLogger) func(service.WorkspaceStore) service.WorkspaceStore {
	return func(next service.WorkspaceStore) service.WorkspaceStore {
		return &workspaceStorageLogging{
			logger: logger,
			next:   next,
		}
	}
}

func (c workspaceStorageLogging) ListWorkspaces(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, IDIn *[]uint64, allowDeleted bool) ([]*model.Workspace, *common_model.PaginationResult, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	res, paginationRes, err := c.next.ListWorkspaces(ctx, tenantID, pagination, IDIn, allowDeleted)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, nil, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithCountField(len(res)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, paginationRes, nil
}

func (c workspaceStorageLogging) GetWorkspace(ctx context.Context, tenantID uint64, workspaceID uint64) (*model.Workspace, error) {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	res, err := c.next.GetWorkspace(ctx, tenantID, workspaceID)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			logger.WithWorkspaceIDField(workspaceID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithWorkspaceIDField(workspaceID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c workspaceStorageLogging) DeleteWorkspace(ctx context.Context, tenantID, workspaceID uint64) error {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	err := c.next.DeleteWorkspace(ctx, tenantID, workspaceID)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			logger.WithWorkspaceIDField(workspaceID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}
	c.logger.Debug(ctx, "request completed",
		logger.WithWorkspaceIDField(workspaceID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c workspaceStorageLogging) UpdateWorkspace(ctx context.Context, tenantID uint64, workspace *model.Workspace) (*model.Workspace, error) {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	updatedWorkspace, err := c.next.UpdateWorkspace(ctx, tenantID, workspace)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			logger.WithWorkspaceIDField(workspace.ID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}
	c.logger.Debug(ctx, "request completed",
		logger.WithWorkspaceIDField(workspace.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)

	return updatedWorkspace, nil
}

func (c workspaceStorageLogging) CreateWorkspace(ctx context.Context, tenantID uint64, workspace *model.Workspace) (*model.Workspace, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	newWorkspace, err := c.next.CreateWorkspace(ctx, tenantID, workspace)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithWorkspaceIDField(newWorkspace.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)

	return newWorkspace, nil
}

func (c workspaceStorageLogging) DeleteOldWorkspaces(ctx context.Context, timeout time.Duration) ([]*model.Workspace, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	deletedWorkspaces, err := c.next.DeleteOldWorkspaces(ctx, timeout)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, "request completed",
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)

	return deletedWorkspaces, nil
}

func (c workspaceStorageLogging) UpdateWorkspaceStatus(ctx context.Context, tenantID uint64, workspaceID uint64, networkPolicyStatus, networkPolicyMessage string) error {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	err := c.next.UpdateWorkspaceStatus(ctx, tenantID, workspaceID, networkPolicyStatus, networkPolicyMessage)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			logger.WithWorkspaceIDField(workspaceID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithWorkspaceIDField(workspaceID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c workspaceStorageLogging) GetWorkspaceSvc(ctx context.Context, tenantID, workspaceSvcID uint64) (*model.WorkspaceSvc, error) {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	res, err := c.next.GetWorkspaceSvc(ctx, tenantID, workspaceSvcID)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			zap.Uint64("workspace_svc_id", workspaceSvcID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, "request completed",
		zap.Uint64("workspace_svc_id", workspaceSvcID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c workspaceStorageLogging) ListWorkspaceSvcs(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, workspaceIDsIn *[]uint64) ([]*model.WorkspaceSvc, *common_model.PaginationResult, error) {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	res, paginationRes, err := c.next.ListWorkspaceSvcs(ctx, tenantID, pagination, workspaceIDsIn)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, nil, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithCountField(len(res)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, paginationRes, nil
}

func (c workspaceStorageLogging) ListWorkspaceSvcsByWorkspace(ctx context.Context, workspaceID uint64) ([]*model.WorkspaceSvc, error) {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	res, err := c.next.ListWorkspaceSvcsByWorkspace(ctx, workspaceID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithWorkspaceIDField(workspaceID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithWorkspaceIDField(workspaceID),
		logger.WithCountField(len(res)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c workspaceStorageLogging) CreateWorkspaceSvc(ctx context.Context, tenantID uint64, svc *model.WorkspaceSvc) (*model.WorkspaceSvc, error) {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	created, err := c.next.CreateWorkspaceSvc(ctx, tenantID, svc)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, "request completed",
		zap.Uint64("workspace_svc_id", created.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return created, nil
}

func (c workspaceStorageLogging) UpdateWorkspaceSvc(ctx context.Context, tenantID uint64, svc *model.WorkspaceSvc) (*model.WorkspaceSvc, error) {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	updated, err := c.next.UpdateWorkspaceSvc(ctx, tenantID, svc)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			zap.Uint64("workspace_svc_id", svc.ID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, "request completed",
		zap.Uint64("workspace_svc_id", svc.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return updated, nil
}

func (c workspaceStorageLogging) DeleteWorkspaceSvc(ctx context.Context, tenantID, workspaceSvcID uint64) error {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	err := c.next.DeleteWorkspaceSvc(ctx, tenantID, workspaceSvcID)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			zap.Uint64("workspace_svc_id", workspaceSvcID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}

	c.logger.Debug(ctx, "request completed",
		zap.Uint64("workspace_svc_id", workspaceSvcID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c workspaceStorageLogging) UpdateWorkspaceSvcStatuses(ctx context.Context, workspaceID uint64, statuses map[string]model.WorkspaceSvcStatusUpdate) error {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	err := c.next.UpdateWorkspaceSvcStatuses(ctx, workspaceID, statuses)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			logger.WithWorkspaceIDField(workspaceID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithWorkspaceIDField(workspaceID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}
