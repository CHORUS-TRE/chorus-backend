package middleware

import (
	"context"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	common "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/service"

	"go.uber.org/zap"
)

type userStorageLogging struct {
	logger *logger.ContextLogger
	next   service.UserStore
}

func Logging(logger *logger.ContextLogger) func(service.UserStore) service.UserStore {
	return func(next service.UserStore) service.UserStore {
		return &userStorageLogging{
			logger: logger,
			next:   next,
		}
	}
}

func (c userStorageLogging) ListUsers(ctx context.Context, tenantID uint64, pagination *common.Pagination, filter *service.UserFilter) ([]*model.User, *common.PaginationResult, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	users, paginationRes, err := c.next.ListUsers(ctx, tenantID, pagination, filter)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, nil, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithCountField(len(users)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return users, paginationRes, nil
}

func (c userStorageLogging) GetUser(ctx context.Context, tenantID uint64, userID uint64) (*model.User, error) {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	res, err := c.next.GetUser(ctx, tenantID, userID)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			logger.WithUserIDField(userID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithUserIDField(userID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c userStorageLogging) SoftDeleteUser(ctx context.Context, tenantID, userID uint64) error {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	err := c.next.SoftDeleteUser(ctx, tenantID, userID)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			logger.WithUserIDField(userID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}
	c.logger.Debug(ctx, "request completed",
		logger.WithUserIDField(userID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c userStorageLogging) UpdateUser(ctx context.Context, tenantID uint64, user *model.User) (*model.User, error) {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	updatedUser, err := c.next.UpdateUser(ctx, tenantID, user)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			logger.WithUserIDField(user.ID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}
	c.logger.Debug(ctx, "request completed",
		logger.WithUserIDField(user.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return updatedUser, nil
}

func (c userStorageLogging) CreateUser(ctx context.Context, tenantID uint64, user *model.User) (*model.User, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	res, err := c.next.CreateUser(ctx, tenantID, user)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithUserIDField(res.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c userStorageLogging) CreateUserRoles(ctx context.Context, userID uint64, roles []model.UserRole) error {
	c.logger.Debug(ctx, "request started", logger.WithUserIDField(userID))

	now := time.Now()

	err := c.next.CreateUserRoles(ctx, userID, roles)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithUserIDField(userID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithUserIDField(userID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c userStorageLogging) RemoveUserRoles(ctx context.Context, userID uint64, userRoleIDs []uint64) error {
	c.logger.Debug(ctx, "request started", logger.WithUserIDField(userID))

	now := time.Now()

	err := c.next.RemoveUserRoles(ctx, userID, userRoleIDs)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithUserIDField(userID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithUserIDField(userID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c userStorageLogging) CreateRole(ctx context.Context, role string) error {
	c.logger.Debug(ctx, "request started", zap.String("role", role))

	now := time.Now()

	err := c.next.CreateRole(ctx, role)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}

	c.logger.Debug(ctx, "request completed",
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c userStorageLogging) GetRoles(ctx context.Context) ([]*model.Role, error) {
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

func (c userStorageLogging) GetTotpRecoveryCodes(ctx context.Context, tenantID, userID uint64) ([]*model.TotpRecoveryCode, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	res, err := c.next.GetTotpRecoveryCodes(ctx, tenantID, userID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithUserIDField(userID),
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

func (a userStorageLogging) DeleteTotpRecoveryCode(ctx context.Context, tenantID, codeID uint64) error {
	a.logger.Debug(ctx, "request started", zap.Uint64("tenant_id", tenantID), zap.Uint64("code", codeID))

	now := time.Now()

	if err := a.next.DeleteTotpRecoveryCode(ctx, tenantID, codeID); err != nil {
		a.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("tenant_id", tenantID),
			zap.Uint64("code", codeID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}

	a.logger.Debug(ctx, "request completed",
		zap.Uint64("tenant_id", tenantID),
		zap.Uint64("code", codeID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c userStorageLogging) UpdateUserWithRecoveryCodes(ctx context.Context, tenantID uint64, user *model.User, totpRecoveryCodes []string) (*model.User, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	updatedUser, err := c.next.UpdateUserWithRecoveryCodes(ctx, tenantID, user, totpRecoveryCodes)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, "request completed",
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return updatedUser, nil
}

func (c userStorageLogging) UpsertGrants(ctx context.Context, grants []model.UserGrant) error {
	c.logger.Debug(ctx, "UpsertGrants request started")

	now := time.Now()

	err := c.next.UpsertGrants(ctx, grants)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}

	c.logger.Debug(ctx, "UpsertGrants request completed",
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c userStorageLogging) DeleteGrants(ctx context.Context, tenantID uint64, userID uint64, clientID string) error {
	c.logger.Debug(ctx, "DeleteGrants request started")

	now := time.Now()

	err := c.next.DeleteGrants(ctx, tenantID, userID, clientID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}

	c.logger.Debug(ctx, "DeleteGrants request completed",
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c userStorageLogging) GetUserGrants(ctx context.Context, tenantID uint64, userID uint64, clientID string) ([]model.UserGrant, error) {
	c.logger.Debug(ctx, "GetUserGrants request started")

	now := time.Now()

	res, err := c.next.GetUserGrants(ctx, tenantID, userID, clientID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, "GetUserGrants request completed",
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}
