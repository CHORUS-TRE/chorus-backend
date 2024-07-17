package middleware

import (
	"context"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/service"

	"go.uber.org/zap"
)

type workbenchStorageLogging struct {
	logger *logger.ContextLogger
	next   service.WorkbenchStore
}

func Logging(logger *logger.ContextLogger) func(service.WorkbenchStore) service.WorkbenchStore {
	return func(next service.WorkbenchStore) service.WorkbenchStore {
		return &workbenchStorageLogging{
			logger: logger,
			next:   next,
		}
	}
}

func (c workbenchStorageLogging) ListWorkbenchs(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.Workbench, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	res, err := c.next.ListWorkbenchs(ctx, tenantID, pagination)
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

func (c workbenchStorageLogging) GetWorkbench(ctx context.Context, tenantID uint64, workbenchID uint64) (*model.Workbench, error) {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	res, err := c.next.GetWorkbench(ctx, tenantID, workbenchID)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			logger.WithWorkbenchIDField(workbenchID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithWorkbenchIDField(workbenchID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c workbenchStorageLogging) DeleteWorkbench(ctx context.Context, tenantID, workbenchID uint64) error {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	err := c.next.DeleteWorkbench(ctx, tenantID, workbenchID)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			logger.WithWorkbenchIDField(workbenchID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}
	c.logger.Debug(ctx, "request completed",
		logger.WithWorkbenchIDField(workbenchID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c workbenchStorageLogging) UpdateWorkbench(ctx context.Context, tenantID uint64, workbench *model.Workbench) error {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	err := c.next.UpdateWorkbench(ctx, tenantID, workbench)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			logger.WithWorkbenchIDField(workbench.ID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}
	c.logger.Debug(ctx, "request completed",
		logger.WithWorkbenchIDField(workbench.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c workbenchStorageLogging) CreateWorkbench(ctx context.Context, tenantID uint64, workbench *model.Workbench) (uint64, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	workbenchId, err := c.next.CreateWorkbench(ctx, tenantID, workbench)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return 0, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithWorkbenchIDField(workbenchId),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return workbenchId, nil
}
