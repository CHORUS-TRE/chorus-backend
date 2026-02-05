package middleware

import (
	"context"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	authorization_model "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	authorization_service "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"

	"go.uber.org/zap"
)

type userPermissionStorageLogging struct {
	logger *logger.ContextLogger
	next   authorization_service.UserPermissionStore
}

func UserPermissionLogging(logger *logger.ContextLogger) func(authorization_service.UserPermissionStore) authorization_service.UserPermissionStore {
	return func(next authorization_service.UserPermissionStore) authorization_service.UserPermissionStore {
		return &userPermissionStorageLogging{
			logger: logger,
			next:   next,
		}
	}
}

func (c userPermissionStorageLogging) FindUsersWithPermission(ctx context.Context, tenantID uint64, filter authorization_model.FindUsersWithPermissionFilter) ([]uint64, error) {
	c.logger.Debug(ctx, "FindUsersWithPermission started",
		zap.Uint64("tenantID", tenantID),
		zap.String("permissionName", string(filter.PermissionName)),
	)

	now := time.Now()

	res, err := c.next.FindUsersWithPermission(ctx, tenantID, filter)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, "FindUsersWithPermission completed",
		zap.Int("resultCount", len(res)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}
