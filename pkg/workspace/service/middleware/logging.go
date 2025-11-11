package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/service"

	"go.uber.org/zap"
)

type workspaceServiceLogging struct {
	logger *logger.ContextLogger
	next   service.Workspaceer
}

func Logging(logger *logger.ContextLogger) func(service.Workspaceer) service.Workspaceer {
	return func(next service.Workspaceer) service.Workspaceer {
		return &workspaceServiceLogging{
			logger: logger,
			next:   next,
		}
	}
}

func (c workspaceServiceLogging) ListWorkspaces(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, filter service.WorkspaceFilter) ([]*model.Workspace, *common_model.PaginationResult, error) {
	now := time.Now()

	res, paginationRes, err := c.next.ListWorkspaces(ctx, tenantID, pagination, filter)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, nil, fmt.Errorf("unable to get workspaces: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		zap.Int("num_workspaces", len(res)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)

	return res, paginationRes, nil
}

func (c workspaceServiceLogging) GetWorkspace(ctx context.Context, tenantID, workspaceID uint64) (*model.Workspace, error) {
	now := time.Now()

	res, err := c.next.GetWorkspace(ctx, tenantID, workspaceID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithWorkspaceIDField(workspaceID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return res, fmt.Errorf("unable to get workspace: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithWorkspaceIDField(workspaceID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c workspaceServiceLogging) DeleteWorkspace(ctx context.Context, tenantID, workspaceID uint64) error {
	now := time.Now()

	err := c.next.DeleteWorkspace(ctx, tenantID, workspaceID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			logger.WithWorkspaceIDField(workspaceID),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return fmt.Errorf("unable to delete workspace: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithWorkspaceIDField(workspaceID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c workspaceServiceLogging) UpdateWorkspace(ctx context.Context, workspace *model.Workspace) (*model.Workspace, error) {
	now := time.Now()

	updatedWorkspace, err := c.next.UpdateWorkspace(ctx, workspace)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithWorkspaceIDField(workspace.ID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, fmt.Errorf("unable to update workspace: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithWorkspaceIDField(workspace.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)

	return updatedWorkspace, nil
}

func (c workspaceServiceLogging) CreateWorkspace(ctx context.Context, workspace *model.Workspace) (*model.Workspace, error) {
	now := time.Now()

	newWorkspace, err := c.next.CreateWorkspace(ctx, workspace)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, fmt.Errorf("unable to create workspace: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithWorkspaceIDField(newWorkspace.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)

	return newWorkspace, nil
}

func (c workspaceServiceLogging) ManageUserRoleInWorkspace(ctx context.Context, tenantID, userID uint64, role user_model.UserRole) error {
	now := time.Now()

	err := c.next.ManageUserRoleInWorkspace(ctx, tenantID, userID, role)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return fmt.Errorf("unable to manage user role in workspace: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)

	return nil
}

func (c workspaceServiceLogging) RemoveUserFromWorkspace(ctx context.Context, tenantID, userID uint64, workspaceID uint64) error {
	now := time.Now()

	err := c.next.RemoveUserFromWorkspace(ctx, tenantID, userID, workspaceID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return fmt.Errorf("unable to remove user from workspace: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)

	return nil
}
