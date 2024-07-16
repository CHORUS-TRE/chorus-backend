package middleware

import (
	"context"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app/service"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"

	"go.uber.org/zap"
)

type appStorageLogging struct {
	logger *logger.ContextLogger
	next   service.AppStore
}

func Logging(logger *logger.ContextLogger) func(service.AppStore) service.AppStore {
	return func(next service.AppStore) service.AppStore {
		return &appStorageLogging{
			logger: logger,
			next:   next,
		}
	}
}

func (c appStorageLogging) ListApps(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.App, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	res, err := c.next.ListApps(ctx, tenantID, pagination)
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

func (c appStorageLogging) GetApp(ctx context.Context, tenantID uint64, appID uint64) (*model.App, error) {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	res, err := c.next.GetApp(ctx, tenantID, appID)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			logger.WithAppIDField(appID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithAppIDField(appID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c appStorageLogging) DeleteApp(ctx context.Context, tenantID, appID uint64) error {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	err := c.next.DeleteApp(ctx, tenantID, appID)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			logger.WithAppIDField(appID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}
	c.logger.Debug(ctx, "request completed",
		logger.WithAppIDField(appID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c appStorageLogging) UpdateApp(ctx context.Context, tenantID uint64, app *model.App) error {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	err := c.next.UpdateApp(ctx, tenantID, app)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			logger.WithAppIDField(app.ID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}
	c.logger.Debug(ctx, "request completed",
		logger.WithAppIDField(app.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c appStorageLogging) CreateApp(ctx context.Context, tenantID uint64, app *model.App) (uint64, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	appId, err := c.next.CreateApp(ctx, tenantID, app)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return 0, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithAppIDField(appId),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return appId, nil
}
