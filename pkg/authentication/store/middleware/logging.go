package middleware

import (
	"context"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authentication/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authentication/service"

	"go.uber.org/zap"
)

type authenticationStorageLogging struct {
	logger *logger.ContextLogger
	next   service.AuthenticationStore
}

func (a authenticationStorageLogging) GetActiveUser(ctx context.Context, username string) (*model.User, error) {
	a.logger.Debug(ctx, "request started")

	now := time.Now()

	res, err := a.next.GetActiveUser(ctx, username)
	if err != nil {
		a.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return res, err
	}

	a.logger.Debug(ctx, "request completed",
		logger.WithUserIDField(res.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)

	return res, nil
}

func Logging(logger *logger.ContextLogger) func(service.AuthenticationStore) service.AuthenticationStore {
	return func(next service.AuthenticationStore) service.AuthenticationStore {
		return &authenticationStorageLogging{
			logger: logger,
			next:   next,
		}
	}
}
