package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/service"

	"github.com/pkg/errors"
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

func (c workbenchServiceLogging) ListWorkbenchs(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.Workbench, error) {
	now := time.Now()

	res, err := c.next.ListWorkbenchs(ctx, tenantID, pagination)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return res, errors.Wrapf(err, "unable to get workbenchs")
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		zap.Int("num_workbenchs", len(res)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c workbenchServiceLogging) ProxyWorkbench(ctx context.Context, tenantID, workbenchID uint64, w http.ResponseWriter, r *http.Request) error {
	now := time.Now()

	err := c.next.ProxyWorkbench(ctx, tenantID, workbenchID, w, r)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return errors.Wrapf(err, "unable to proxy workbenchs")
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
		return res, errors.Wrapf(err, "unable to get workbench")
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
		return errors.Wrapf(err, "unable to delete workbench")
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithWorkbenchIDField(workbenchID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c workbenchServiceLogging) UpdateWorkbench(ctx context.Context, workbench *model.Workbench) error {
	now := time.Now()

	err := c.next.UpdateWorkbench(ctx, workbench)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithWorkbenchIDField(workbench.ID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return errors.Wrapf(err, "unable to update workbench")
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithWorkbenchIDField(workbench.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c workbenchServiceLogging) CreateWorkbench(ctx context.Context, workbench *model.Workbench) (uint64, error) {
	now := time.Now()

	workbenchId, err := c.next.CreateWorkbench(ctx, workbench)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return workbenchId, errors.Wrapf(err, "unable to create workbench")
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithWorkbenchIDField(workbenchId),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return workbenchId, nil
}
