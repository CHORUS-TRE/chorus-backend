package middleware

import (
	"context"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/common"
	"github.com/luikyv/go-oidc/pkg/goidc"

	"go.uber.org/zap"
)

type logoutSessionManagerLogging struct {
	logger *logger.ContextLogger
	next   goidc.LogoutSessionManager
}

func LogoutLogging(log *logger.ContextLogger) func(goidc.LogoutSessionManager) goidc.LogoutSessionManager {
	l := logger.With(log, zap.String("layer", "store"))
	return func(next goidc.LogoutSessionManager) goidc.LogoutSessionManager {
		return &logoutSessionManagerLogging{
			logger: l,
			next:   next,
		}
	}
}

func (s *logoutSessionManagerLogging) Save(ctx context.Context, session *goidc.LogoutSession) error {
	log := logger.With(s.logger,
		zap.String("service", "Save"),
		zap.String("session_id", session.ID),
	)
	s.logger.Debug(ctx, logger.LoggerMessageRequestStarted)

	now := time.Now()

	err := s.next.Save(ctx, session)
	return common.LogErrorIfAny(err, ctx, now, log)
}

func (s *logoutSessionManagerLogging) SessionByCallbackID(ctx context.Context, callbackID string) (*goidc.LogoutSession, error) {
	log := logger.With(s.logger,
		zap.String("service", "SessionByCallbackID"),
		zap.String("callback_id", callbackID),
	)
	s.logger.Debug(ctx, logger.LoggerMessageRequestStarted)

	now := time.Now()

	res, err := s.next.SessionByCallbackID(ctx, callbackID)
	return res, common.LogErrorIfAny(err, ctx, now, logger.With(log, zap.Any("result", res)))
}

func (s *logoutSessionManagerLogging) Delete(ctx context.Context, id string) error {
	log := logger.With(s.logger,
		zap.String("service", "Delete"),
		zap.String("session_id", id),
	)
	s.logger.Debug(ctx, logger.LoggerMessageRequestStarted)

	now := time.Now()

	err := s.next.Delete(ctx, id)
	return common.LogErrorIfAny(err, ctx, now, log)
}
