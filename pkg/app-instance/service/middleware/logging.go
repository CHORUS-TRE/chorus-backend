package middleware

import (
	"context"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app-instance/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app-instance/service"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"

	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type appInstanceServiceLogging struct {
	logger *logger.ContextLogger
	next   service.AppInstanceer
}

func Logging(logger *logger.ContextLogger) func(service.AppInstanceer) service.AppInstanceer {
	return func(next service.AppInstanceer) service.AppInstanceer {
		return &appInstanceServiceLogging{
			logger: logger,
			next:   next,
		}
	}
}

func (c appInstanceServiceLogging) ListAppInstances(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.AppInstance, error) {
	now := time.Now()

	res, err := c.next.ListAppInstances(ctx, tenantID, pagination)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return res, errors.Wrapf(err, "unable to get appInstances")
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		zap.Int("num_appInstances", len(res)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c appInstanceServiceLogging) GetAppInstance(ctx context.Context, tenantID, appInstanceID uint64) (*model.AppInstance, error) {
	now := time.Now()

	res, err := c.next.GetAppInstance(ctx, tenantID, appInstanceID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithAppInstanceIDField(appInstanceID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return res, errors.Wrapf(err, "unable to get appInstance")
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithAppInstanceIDField(appInstanceID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c appInstanceServiceLogging) DeleteAppInstance(ctx context.Context, tenantID, appInstanceID uint64) error {
	now := time.Now()

	err := c.next.DeleteAppInstance(ctx, tenantID, appInstanceID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			logger.WithAppInstanceIDField(appInstanceID),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return errors.Wrapf(err, "unable to delete appInstance")
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithAppInstanceIDField(appInstanceID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c appInstanceServiceLogging) UpdateAppInstance(ctx context.Context, appInstance *model.AppInstance) error {
	now := time.Now()

	err := c.next.UpdateAppInstance(ctx, appInstance)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithAppInstanceIDField(appInstance.ID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return errors.Wrapf(err, "unable to update appInstance")
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithAppInstanceIDField(appInstance.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c appInstanceServiceLogging) CreateAppInstance(ctx context.Context, appInstance *model.AppInstance) (uint64, error) {
	now := time.Now()

	appInstanceId, err := c.next.CreateAppInstance(ctx, appInstance)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return appInstanceId, errors.Wrapf(err, "unable to create appInstance")
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		logger.WithAppInstanceIDField(appInstanceId),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return appInstanceId, nil
}
