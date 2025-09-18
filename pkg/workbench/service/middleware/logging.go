package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/service"

	"go.uber.org/zap"
)

type workbenchServiceLogging struct {
	logger *logger.ContextLogger
	next   service.Workbencher
}

func Logging(logger *logger.ContextLogger) func(service.Workbencher) service.Workbencher {
	return func(next service.Workbencher) service.Workbencher {
		return &workbenchServiceLogging{
			logger: logger,
			next:   next,
		}
	}
}

func (c workbenchServiceLogging) ListWorkbenchs(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, filter service.WorkbenchFilter) ([]*model.Workbench, *common_model.PaginationResult, error) {
	now := time.Now()

	res, paginationRes, err := c.next.ListWorkbenchs(ctx, tenantID, pagination, filter)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, nil, fmt.Errorf("unable to get workbenchs: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		zap.Int("num_workbenchs", len(res)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, paginationRes, nil
}

func (c workbenchServiceLogging) ProxyWorkbench(ctx context.Context, tenantID, workbenchID uint64, w http.ResponseWriter, r *http.Request) error {
	now := time.Now()

	err := c.next.ProxyWorkbench(ctx, tenantID, workbenchID, w, r)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return fmt.Errorf("unable to proxy workbenchs: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)

	return nil
}

func (c workbenchServiceLogging) GetWorkbench(ctx context.Context, tenantID, workbenchID uint64) (*model.Workbench, error) {
	now := time.Now()

	res, err := c.next.GetWorkbench(ctx, tenantID, workbenchID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithWorkbenchIDField(workbenchID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return res, fmt.Errorf("unable to get workbench: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithWorkbenchIDField(workbenchID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c workbenchServiceLogging) DeleteWorkbench(ctx context.Context, tenantID, workbenchID uint64) error {
	now := time.Now()

	err := c.next.DeleteWorkbench(ctx, tenantID, workbenchID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			logger.WithWorkbenchIDField(workbenchID),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return fmt.Errorf("unable to delete workbench: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithWorkbenchIDField(workbenchID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c workbenchServiceLogging) DeleteWorkbenchsInWorkspace(ctx context.Context, tenantID, workspaceID uint64) error {
	now := time.Now()

	err := c.next.DeleteWorkbenchsInWorkspace(ctx, tenantID, workspaceID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			logger.WithWorkspaceIDField(workspaceID),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return fmt.Errorf("unable to delete workbenchs in workspace %v: %w", workspaceID, err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithWorkspaceIDField(workspaceID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c workbenchServiceLogging) UpdateWorkbench(ctx context.Context, workbench *model.Workbench) (*model.Workbench, error) {
	now := time.Now()

	updatedWorkbench, err := c.next.UpdateWorkbench(ctx, workbench)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithWorkbenchIDField(workbench.ID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, fmt.Errorf("unable to update workbench: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithWorkbenchIDField(updatedWorkbench.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return updatedWorkbench, nil
}

func (c workbenchServiceLogging) CreateWorkbench(ctx context.Context, workbench *model.Workbench) (*model.Workbench, error) {
	now := time.Now()

	newWorkbench, err := c.next.CreateWorkbench(ctx, workbench)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, fmt.Errorf("unable to create workbench: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithWorkbenchIDField(newWorkbench.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return newWorkbench, nil
}

func (c workbenchServiceLogging) ListAppInstances(ctx context.Context, tenantID uint64, pagination *common_model.Pagination) ([]*model.AppInstance, *common_model.PaginationResult, error) {
	now := time.Now()

	res, paginationRes, err := c.next.ListAppInstances(ctx, tenantID, pagination)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, nil, fmt.Errorf("unable to get appInstances: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		zap.Int("num_appInstances", len(res)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, paginationRes, nil
}

func (c workbenchServiceLogging) GetAppInstance(ctx context.Context, tenantID, appInstanceID uint64) (*model.AppInstance, error) {
	now := time.Now()

	res, err := c.next.GetAppInstance(ctx, tenantID, appInstanceID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithAppInstanceIDField(appInstanceID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return res, fmt.Errorf("unable to get appInstance: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithAppInstanceIDField(appInstanceID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c workbenchServiceLogging) DeleteAppInstance(ctx context.Context, tenantID, appInstanceID uint64) error {
	now := time.Now()

	err := c.next.DeleteAppInstance(ctx, tenantID, appInstanceID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			logger.WithAppInstanceIDField(appInstanceID),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return fmt.Errorf("unable to delete appInstance: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithAppInstanceIDField(appInstanceID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c workbenchServiceLogging) UpdateAppInstance(ctx context.Context, appInstance *model.AppInstance) (*model.AppInstance, error) {
	now := time.Now()

	updatedAppInstance, err := c.next.UpdateAppInstance(ctx, appInstance)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithAppInstanceIDField(appInstance.ID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, fmt.Errorf("unable to update appInstance: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithAppInstanceIDField(updatedAppInstance.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return updatedAppInstance, nil
}

func (c workbenchServiceLogging) CreateAppInstance(ctx context.Context, appInstance *model.AppInstance) (*model.AppInstance, error) {
	now := time.Now()

	newAppInstance, err := c.next.CreateAppInstance(ctx, appInstance)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, fmt.Errorf("unable to create appInstance: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithAppInstanceIDField(newAppInstance.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return newAppInstance, nil
}

func (c workbenchServiceLogging) ManageUserRoleInWorkbench(ctx context.Context, tenantID, userID uint64, role user_model.UserRole) error {
	now := time.Now()

	err := c.next.ManageUserRoleInWorkbench(ctx, tenantID, userID, role)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			logger.WithTenantIDField(tenantID),
			logger.WithUserIDField(userID),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return fmt.Errorf("unable to manage user role in workbench: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithTenantIDField(tenantID),
		logger.WithUserIDField(userID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c workbenchServiceLogging) RemoveUserFromWorkbench(ctx context.Context, tenantID, userID, workbenchID uint64) error {
	now := time.Now()

	err := c.next.RemoveUserFromWorkbench(ctx, tenantID, userID, workbenchID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			logger.WithWorkbenchIDField(workbenchID),
			logger.WithUserIDField(userID),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return fmt.Errorf("unable to remove user from workbench: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithWorkbenchIDField(workbenchID),
		logger.WithUserIDField(userID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}
