package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/service"

	"go.uber.org/zap"
)

type userServiceLogging struct {
	logger *logger.ContextLogger
	next   service.Userer
}

func Logging(logger *logger.ContextLogger) func(service.Userer) service.Userer {
	return func(next service.Userer) service.Userer {
		return &userServiceLogging{
			logger: logger,
			next:   next,
		}
	}
}

func (c userServiceLogging) GetUsers(ctx context.Context, req service.GetUsersReq) ([]*model.User, error) {
	now := time.Now()

	res, err := c.next.GetUsers(ctx, req)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return res, fmt.Errorf("unable to get users: %w", err)
	}

	c.logger.Info(ctx, "request completed",
		zap.Int("num_users", len(res)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c userServiceLogging) GetUser(ctx context.Context, req service.GetUserReq) (*model.User, error) {
	now := time.Now()

	res, err := c.next.GetUser(ctx, req)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithUserIDField(req.ID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return res, fmt.Errorf("unable to get user: %w", err)
	}

	c.logger.Info(ctx, "request completed",
		logger.WithUserIDField(req.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c userServiceLogging) SoftDeleteUser(ctx context.Context, req service.DeleteUserReq) error {
	now := time.Now()

	err := c.next.SoftDeleteUser(ctx, req)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			logger.WithUserIDField(req.ID),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return fmt.Errorf("unable to delete user: %w", err)
	}

	c.logger.Info(ctx, "request completed",
		logger.WithUserIDField(req.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c userServiceLogging) UpdateUser(ctx context.Context, req service.UpdateUserReq) error {
	now := time.Now()

	err := c.next.UpdateUser(ctx, req)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithUserIDField(req.User.ID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return fmt.Errorf("unable to update user: %w", err)
	}

	c.logger.Info(ctx, "request completed",
		logger.WithUserIDField(req.User.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c userServiceLogging) CreateUser(ctx context.Context, req service.CreateUserReq) (uint64, error) {
	now := time.Now()

	userId, err := c.next.CreateUser(ctx, req)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return userId, fmt.Errorf("unable to create user: %w", err)
	}

	c.logger.Info(ctx, "request completed",
		logger.WithUserIDField(userId),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return userId, nil
}

func (c userServiceLogging) CreateRole(ctx context.Context, role string) error {
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

func (c userServiceLogging) GetRoles(ctx context.Context) ([]*model.Role, error) {
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

func (c userServiceLogging) UpdateUserPassword(ctx context.Context, req service.UpdateUserPasswordReq) error {
	c.logger.Debug(ctx, "request started")

	now := time.Now()

	err := c.next.UpdateUserPassword(ctx, req)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithUserIDField(req.UserID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithUserIDField(req.UserID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c userServiceLogging) EnableUserTotp(ctx context.Context, req service.EnableTotpReq) error {
	c.logger.Debug(ctx, "request started")

	now := time.Now()
	err := c.next.EnableUserTotp(ctx, req)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithUserIDField(req.UserID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithUserIDField(req.UserID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}

func (c userServiceLogging) ResetUserTotp(ctx context.Context, req service.ResetTotpReq) (string, []string, error) {
	c.logger.Debug(ctx, "request started")

	now := time.Now()
	totpSecret, totpRecoveryCodes, err := c.next.ResetUserTotp(ctx, req)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithUserIDField(req.UserID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return "", nil, err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithUserIDField(req.UserID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)

	return totpSecret, totpRecoveryCodes, nil
}

func (c userServiceLogging) ResetUserPassword(ctx context.Context, req service.ResetUserPasswordReq) error {
	c.logger.Debug(ctx, "request started")

	now := time.Now()
	err := c.next.ResetUserPassword(ctx, req)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			logger.WithUserIDField(req.UserID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}

	c.logger.Debug(ctx, "request completed",
		logger.WithUserIDField(req.UserID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)

	return nil
}

func (c userServiceLogging) GetTotpRecoveryCodes(ctx context.Context, tenantID, userID uint64) ([]*model.TotpRecoveryCode, error) {
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
		logger.WithUserIDField(userID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)

	return res, nil
}

func (c userServiceLogging) DeleteTotpRecoveryCode(ctx context.Context, req *service.DeleteTotpRecoveryCodeReq) error {
	c.logger.Debug(ctx, "request started")

	now := time.Now()
	err := c.next.DeleteTotpRecoveryCode(ctx, req)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("totp_code_id", req.CodeID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}

	c.logger.Debug(ctx, "request completed",
		zap.Uint64("totp_code_id", req.CodeID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)

	return nil
}
