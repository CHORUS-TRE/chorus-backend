package middleware

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/platform-settings/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/platform-settings/service"
)

type platformSettingsServiceLogging struct {
	logger *logger.ContextLogger
	next   service.PlatformSettingser
}

func Logging(logger *logger.ContextLogger) func(service.PlatformSettingser) service.PlatformSettingser {
	return func(next service.PlatformSettingser) service.PlatformSettingser {
		return &platformSettingsServiceLogging{
			logger: logger,
			next:   next,
		}
	}
}

func (c platformSettingsServiceLogging) GetPlatformSettings(ctx context.Context) (*model.PlatformSettings, error) {
	now := time.Now()

	res, err := c.next.GetPlatformSettings(ctx)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, fmt.Errorf("unable to get platform settings: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c platformSettingsServiceLogging) UpdatePlatformSettings(ctx context.Context, settings *model.PlatformSettings) (*model.PlatformSettings, error) {
	now := time.Now()

	res, err := c.next.UpdatePlatformSettings(ctx, settings)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("tenant_id", settings.TenantID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, fmt.Errorf("unable to update platform settings: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted,
		zap.Uint64("tenant_id", settings.TenantID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}
