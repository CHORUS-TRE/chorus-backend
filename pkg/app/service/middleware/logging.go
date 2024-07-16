package middleware

import (
	"context"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app/service"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type appServiceLogging struct {
	logger *logger.ContextLogger
	next   service.Apper
}

func Logging(logger *logger.ContextLogger) func(service.Apper) service.Apper {
	return func(next service.Apper) service.Apper {
		return &appServiceLogging{
			logger: logger,
			next:   next,
		}
	}
}

func (c appServiceLogging) ListApps(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.App, error) {
	now := time.Now()

	res, err := c.next.ListApps(ctx, tenantID, pagination)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return res, errors.Wrapf(err, "unable to get apps")
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		zap.Int("num_apps", len(res)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c appServiceLogging) GetApp(ctx context.Context, tenantID, appID uint64) (*model.App, error) {
	now := time.Now()

	res, err := c.next.GetApp(ctx, tenantID, appID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithAppIDField(appID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return res, errors.Wrapf(err, "unable to get app")
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithAppIDField(appID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c appServiceLogging) DeleteApp(ctx context.Context, tenantID, appID uint64) error {
	now := time.Now()

	err := c.next.DeleteApp(ctx, tenantID, appID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			logger.WithAppIDField(appID),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return errors.Wrapf(err, "unable to delete app")
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithAppIDField(appID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c appServiceLogging) UpdateApp(ctx context.Context, app *model.App) error {
	now := time.Now()

	err := c.next.UpdateApp(ctx, app)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithAppIDField(app.ID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return errors.Wrapf(err, "unable to update app")
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithAppIDField(app.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c appServiceLogging) CreateApp(ctx context.Context, app *model.App) (uint64, error) {
	now := time.Now()

	appId, err := c.next.CreateApp(ctx, app)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return appId, errors.Wrapf(err, "unable to create app")
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithAppIDField(appId),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return appId, nil
}
