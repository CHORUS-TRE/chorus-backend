package middleware

import (
	"context"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	authorization_service "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/model"

	"go.uber.org/zap"
)

type userRoleStorageLogging struct {
	logger *logger.ContextLogger
	next   authorization_service.UserRoleStore
}

func Logging(logger *logger.ContextLogger) func(authorization_service.UserRoleStore) authorization_service.UserRoleStore {
	return func(next authorization_service.UserRoleStore) authorization_service.UserRoleStore {
		return &userRoleStorageLogging{
			logger: logger,
			next:   next,
		}
	}
}

func (c userRoleStorageLogging) GetRoles(ctx context.Context) ([]*model.Role, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	res, err := c.next.GetRoles(ctx)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, "request completed",
		zap.Any("result", res),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}
