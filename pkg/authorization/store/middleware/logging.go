package middleware

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"
)

type authorizationStorageLogging struct {
	logger *logger.ContextLogger
	next   service.AuthorizationStore
}

func Logging(logger *logger.ContextLogger) func(service.AuthorizationStore) service.AuthorizationStore {
	return func(next service.AuthorizationStore) service.AuthorizationStore {
		return &authorizationStorageLogging{
			logger: logger,
			next:   next,
		}
	}
}

func (c authorizationStorageLogging) SyncSystemRoles(ctx context.Context, roles []*model.RoleDefinition) error {
	c.logger.Debug(ctx, logger.LoggerMessageRequestStarted)
	now := time.Now()

	err := c.next.SyncSystemRoles(ctx, roles)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Int("num_roles", len(roles)),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}

	c.logger.Debug(ctx, logger.LoggerMessageRequestCompleted,
		zap.Int("num_roles", len(roles)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c authorizationStorageLogging) ListRoles(ctx context.Context) ([]*model.RoleDefinition, error) {
	c.logger.Debug(ctx, logger.LoggerMessageRequestStarted)
	now := time.Now()

	roles, err := c.next.ListRoles(ctx)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, logger.LoggerMessageRequestCompleted,
		zap.Int("num_roles", len(roles)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return roles, nil
}

func (c authorizationStorageLogging) CreateDynamicRole(ctx context.Context, role *model.RoleDefinition) error {
	c.logger.Debug(ctx, logger.LoggerMessageRequestStarted)
	now := time.Now()

	err := c.next.CreateDynamicRole(ctx, role)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.String("role", role.Name.String()),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}

	c.logger.Debug(ctx, logger.LoggerMessageRequestCompleted,
		zap.String("role", role.Name.String()),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c authorizationStorageLogging) FindUsersWithPermission(ctx context.Context, tenantID uint64, filter model.FindUsersWithPermissionFilter, rolesGranting []model.RoleName) ([]uint64, error) {
	c.logger.Debug(ctx, logger.LoggerMessageRequestStarted)
	now := time.Now()

	userIDs, err := c.next.FindUsersWithPermission(ctx, tenantID, filter, rolesGranting)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("tenant_id", tenantID),
			zap.String("permission", filter.PermissionName.String()),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, logger.LoggerMessageRequestCompleted,
		zap.Uint64("tenant_id", tenantID),
		zap.String("permission", filter.PermissionName.String()),
		zap.Int("num_users", len(userIDs)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return userIDs, nil
}
