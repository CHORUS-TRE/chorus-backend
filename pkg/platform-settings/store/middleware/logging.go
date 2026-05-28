package middleware

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/platform-settings/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/platform-settings/service"
)

type platformSettingsStorageLogging struct {
	logger *logger.ContextLogger
	next   service.PlatformSettingsStore
}

func Logging(logger *logger.ContextLogger) func(service.PlatformSettingsStore) service.PlatformSettingsStore {
	return func(next service.PlatformSettingsStore) service.PlatformSettingsStore {
		return &platformSettingsStorageLogging{
			logger: logger,
			next:   next,
		}
	}
}

func (c platformSettingsStorageLogging) GetPlatformSettings(ctx context.Context, tenantID uint64) (*model.PlatformSettings, error) {
	c.logger.Debug(ctx, logger.LoggerMessageRequestStarted)
	now := time.Now()

	res, err := c.next.GetPlatformSettings(ctx, tenantID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("tenant_id", tenantID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, logger.LoggerMessageRequestCompleted,
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c platformSettingsStorageLogging) UpsertPlatformSettings(ctx context.Context, settings *model.PlatformSettings) (*model.PlatformSettings, error) {
	c.logger.Debug(ctx, logger.LoggerMessageRequestStarted)
	now := time.Now()

	res, err := c.next.UpsertPlatformSettings(ctx, settings)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("tenant_id", settings.TenantID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, logger.LoggerMessageRequestCompleted,
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}
