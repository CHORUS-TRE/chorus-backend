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

func (c workspaceStorageLogging) ListWorkspaces(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.Workspace, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	res, err := c.next.ListWorkspaces(ctx, tenantID, pagination)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithCountField(len(res)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
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

func (c workspaceStorageLogging) UpdateWorkspace(ctx context.Context, tenantID uint64, workspace *model.Workspace) error {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	err := c.next.UpdateWorkspace(ctx, tenantID, workspace)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			logger.WithWorkspaceIDField(workspace.ID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}
	c.logger.Debug(ctx, "request completed",
		logger.WithWorkspaceIDField(workspace.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c workspaceStorageLogging) CreateWorkspace(ctx context.Context, tenantID uint64, workspace *model.Workspace) (uint64, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	workspaceId, err := c.next.CreateWorkspace(ctx, tenantID, workspace)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return 0, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithWorkspaceIDField(workspaceId),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return workspaceId, nil
}
