package middleware

import (
	"context"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
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

func (c userStorageLogging) GetUsers(ctx context.Context, tenantID uint64) ([]*model.User, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	res, err := c.next.GetUsers(ctx, tenantID)
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

func (c userStorageLogging) UpdateUser(ctx context.Context, tenantID uint64, user *model.User) error {
	c.logger.Debug(ctx, "request started")
	now := time.Now()

	err := c.next.UpdateUser(ctx, tenantID, user)
	if err != nil {
		c.logger.Error(ctx, "request completed",
			logger.WithUserIDField(user.ID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}
	c.logger.Debug(ctx, "request completed",
		logger.WithUserIDField(user.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c userStorageLogging) CreateUser(ctx context.Context, tenantID uint64, user *model.User) (uint64, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	userId, err := c.next.CreateUser(ctx, tenantID, user)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return 0, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithUserIDField(userId),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return userId, nil
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

func (c userStorageLogging) UpdateUserWithRecoveryCodes(ctx context.Context, tenantID uint64, user *model.User, totpRecoveryCodes []string) error {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	err := c.next.UpdateUserWithRecoveryCodes(ctx, tenantID, user, totpRecoveryCodes)
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
