package middleware

import (
	"context"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/common"
	"github.com/luikyv/go-oidc/pkg/goidc"

	"go.uber.org/zap"
)

type grantSessionManagerLogging struct {
	logger *logger.ContextLogger
	next   goidc.GrantSessionManager
}

func GrantLogging(log *logger.ContextLogger) func(goidc.GrantSessionManager) goidc.GrantSessionManager {
	l := logger.With(log, zap.String("layer", "store"))
	return func(next goidc.GrantSessionManager) goidc.GrantSessionManager {
		return &grantSessionManagerLogging{
			logger: l,
			next:   next,
		}
	}
}

func (s *grantSessionManagerLogging) Save(ctx context.Context, session *goidc.GrantSession) error {
	log := logger.With(s.logger,
		zap.String("service", "Save"),
		zap.String("session_id", session.ID),
	)
	s.logger.Debug(ctx, logger.LoggerMessageRequestStarted)

	now := time.Now()

	err := s.next.Save(ctx, session)
	return common.LogErrorIfAny(err, ctx, now, log)
}

func (s *grantSessionManagerLogging) SessionByTokenID(ctx context.Context, tokenID string) (*goidc.GrantSession, error) {
	log := logger.With(s.logger,
		zap.String("service", "SessionByTokenID"),
		zap.String("token_id", tokenID),
	)
	s.logger.Debug(ctx, logger.LoggerMessageRequestStarted)

	now := time.Now()

	res, err := s.next.SessionByTokenID(ctx, tokenID)
	return res, common.LogErrorIfAny(err, ctx, now, logger.With(log, zap.Any("result", res)))
}

func (s *grantSessionManagerLogging) SessionByRefreshToken(ctx context.Context, refreshToken string) (*goidc.GrantSession, error) {
	log := logger.With(s.logger,
		zap.String("service", "SessionByRefreshToken"),
		zap.String("refresh_token", refreshToken),
	)
	s.logger.Debug(ctx, logger.LoggerMessageRequestStarted)

	now := time.Now()

	res, err := s.next.SessionByRefreshToken(ctx, refreshToken)
	return res, common.LogErrorIfAny(err, ctx, now, logger.With(log, zap.Any("result", res)))
}

func (s *grantSessionManagerLogging) Delete(ctx context.Context, id string) error {
	log := logger.With(s.logger,
		zap.String("service", "Delete"),
		zap.String("session_id", id),
	)
	s.logger.Debug(ctx, logger.LoggerMessageRequestStarted)

	now := time.Now()

	err := s.next.Delete(ctx, id)
	return common.LogErrorIfAny(err, ctx, now, log)
}

func (s *grantSessionManagerLogging) DeleteByAuthCode(ctx context.Context, authCode string) error {
	log := logger.With(s.logger,
		zap.String("service", "DeleteByAuthCode"),
		zap.String("auth_code", authCode),
	)
	s.logger.Debug(ctx, logger.LoggerMessageRequestStarted)

	now := time.Now()

	err := s.next.DeleteByAuthCode(ctx, authCode)
	return common.LogErrorIfAny(err, ctx, now, log)
}
