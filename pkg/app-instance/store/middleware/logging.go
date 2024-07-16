package middleware

import (
	"context"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app-instance/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app-instance/service"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"

	"go.uber.org/zap"
)

type appInstanceStorageLogging struct {
	logger *logger.ContextLogger
	next   service.AppInstanceStore
}

func Logging(logger *logger.ContextLogger) func(service.AppInstanceStore) service.AppInstanceStore {
	return func(next service.AppInstanceStore) service.AppInstanceStore {
		return &appInstanceStorageLogging{
			logger: logger,
			next:   next,
		}
	}
}

func (c appInstanceStorageLogging) ListAppInstances(ctx context.Context, tenantID uint64, pagination common_model.Pagination) ([]*model.AppInstance, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	res, err := c.next.ListAppInstances(ctx, tenantID, pagination)
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

func (c appInstanceStorageLogging) GetAppInstance(ctx context.Context, tenantID uint64, appInstanceID uint64) (*model.AppInstance, error) {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	res, err := c.next.GetAppInstance(ctx, tenantID, appInstanceID)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			logger.WithAppInstanceIDField(appInstanceID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithAppInstanceIDField(appInstanceID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c appInstanceStorageLogging) DeleteAppInstance(ctx context.Context, tenantID, appInstanceID uint64) error {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	err := c.next.DeleteAppInstance(ctx, tenantID, appInstanceID)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			logger.WithAppInstanceIDField(appInstanceID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}
	c.logger.Debug(ctx, "request completed",
		logger.WithAppInstanceIDField(appInstanceID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c appInstanceStorageLogging) UpdateAppInstance(ctx context.Context, tenantID uint64, appInstance *model.AppInstance) error {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	err := c.next.UpdateAppInstance(ctx, tenantID, appInstance)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			logger.WithAppInstanceIDField(appInstance.ID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}
	c.logger.Debug(ctx, "request completed",
		logger.WithAppInstanceIDField(appInstance.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c appInstanceStorageLogging) CreateAppInstance(ctx context.Context, tenantID uint64, appInstance *model.AppInstance) (uint64, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	appInstanceId, err := c.next.CreateAppInstance(ctx, tenantID, appInstance)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return 0, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithAppInstanceIDField(appInstanceId),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return appInstanceId, nil
}
