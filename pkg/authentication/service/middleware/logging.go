package middleware

import (
	"context"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authentication/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authentication/service"

	"go.uber.org/zap"
)

var _ service.Authenticator = (*authenticationServiceLogging)(nil)

type authenticationServiceLogging struct {
	logger *logger.ContextLogger
	next   service.Authenticator
}

func Logging(logger *logger.ContextLogger) func(service.Authenticator) service.Authenticator {
	return func(next service.Authenticator) service.Authenticator {
		return &authenticationServiceLogging{
			logger: logger,
			next:   next,
		}
	}
}

func (a authenticationServiceLogging) GetAuthenticationModes() []model.AuthenticationMode {
	now := time.Now()

	res := a.next.GetAuthenticationModes()

	a.logger.Info(context.Background(), "request completed",
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res
}

func (a authenticationServiceLogging) Authenticate(ctx context.Context, username, password, totp string) (string, time.Duration, error) {
	now := time.Now()

	res, t, err := a.next.Authenticate(ctx, username, password, totp)
	if err != nil {
		a.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return res, t, err
	}

	a.logger.Info(ctx, "request completed",
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, t, err
}

func (a authenticationServiceLogging) RefreshToken(ctx context.Context) (string, time.Duration, error) {
	now := time.Now()

	res, t, err := a.next.RefreshToken(ctx)
	if err != nil {
		a.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return res, t, err
	}

	a.logger.Info(ctx, "request completed",
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)

	return res, t, nil
}

func (a authenticationServiceLogging) AuthenticateOAuth(ctx context.Context, id string) (string, error) {
	now := time.Now()

	res, err := a.next.AuthenticateOAuth(ctx, id)
	if err != nil {
		a.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return res, err
	}

	a.logger.Info(ctx, "request completed",
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, err
}

func (a authenticationServiceLogging) OAuthCallback(ctx context.Context, providerID, state, sessionState, code string) (string, time.Duration, string, error) {
	now := time.Now()

	res, t, url, err := a.next.OAuthCallback(ctx, providerID, state, sessionState, code)
	if err != nil {
		a.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return res, t, url, err
	}

	a.logger.Info(ctx, "request completed",
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, t, url, err
}

func (a authenticationServiceLogging) Logout(ctx context.Context) (string, error) {
	now := time.Now()

	res, err := a.next.Logout(ctx)
	if err != nil {
		a.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return res, err
	}

	a.logger.Info(ctx, "request completed",
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, err
}
