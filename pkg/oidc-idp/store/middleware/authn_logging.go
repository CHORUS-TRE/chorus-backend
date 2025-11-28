package middleware

import (
	"context"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/common"
	"github.com/luikyv/go-oidc/pkg/goidc"

	"go.uber.org/zap"
)

type authnSessionManagerLogging struct {
	logger *logger.ContextLogger
	next   goidc.AuthnSessionManager
}

func AuthnLogging(log *logger.ContextLogger) func(goidc.AuthnSessionManager) goidc.AuthnSessionManager {
	l := logger.With(log, zap.String("layer", "store"))
	return func(next goidc.AuthnSessionManager) goidc.AuthnSessionManager {
		return &authnSessionManagerLogging{
			logger: l,
			next:   next,
		}
	}
}

func (s *authnSessionManagerLogging) Save(ctx context.Context, session *goidc.AuthnSession) error {
	log := logger.With(s.logger,
		zap.String("service", "Save"),
		zap.String("session_id", session.ID),
	)
	s.logger.Debug(ctx, logger.LoggerMessageRequestStarted)

	now := time.Now()

	err := s.next.Save(ctx, session)
	return common.LogErrorIfAny(err, ctx, now, log)
}

func (s *authnSessionManagerLogging) SessionByCallbackID(ctx context.Context, callbackID string) (*goidc.AuthnSession, error) {
	log := logger.With(s.logger,
		zap.String("service", "SessionByCallbackID"),
		zap.String("callback_id", callbackID),
	)
	s.logger.Debug(ctx, logger.LoggerMessageRequestStarted)

	now := time.Now()

	res, err := s.next.SessionByCallbackID(ctx, callbackID)
	return res, common.LogErrorIfAny(err, ctx, now, logger.With(log, zap.Any("result", res)))
}

func (s *authnSessionManagerLogging) SessionByAuthCode(ctx context.Context, authorizationCode string) (*goidc.AuthnSession, error) {
	log := logger.With(s.logger,
		zap.String("service", "SessionByAuthCode"),
		zap.String("authorization_code", authorizationCode),
	)
	s.logger.Debug(ctx, logger.LoggerMessageRequestStarted)

	now := time.Now()

	res, err := s.next.SessionByAuthCode(ctx, authorizationCode)
	return res, common.LogErrorIfAny(err, ctx, now, logger.With(log, zap.Any("result", res)))
}

func (s *authnSessionManagerLogging) SessionByPushedAuthReqID(ctx context.Context, id string) (*goidc.AuthnSession, error) {
	log := logger.With(s.logger,
		zap.String("service", "SessionByPushedAuthReqID"),
		zap.String("pushed_auth_req_id", id),
	)
	s.logger.Debug(ctx, logger.LoggerMessageRequestStarted)

	now := time.Now()

	res, err := s.next.SessionByPushedAuthReqID(ctx, id)
	return res, common.LogErrorIfAny(err, ctx, now, logger.With(log, zap.Any("result", res)))
}

func (s *authnSessionManagerLogging) SessionByCIBAAuthID(ctx context.Context, id string) (*goidc.AuthnSession, error) {
	log := logger.With(s.logger,
		zap.String("service", "SessionByCIBAAuthID"),
		zap.String("ciba_auth_id", id),
	)
	s.logger.Debug(ctx, logger.LoggerMessageRequestStarted)

	now := time.Now()

	res, err := s.next.SessionByCIBAAuthID(ctx, id)
	return res, common.LogErrorIfAny(err, ctx, now, logger.With(log, zap.Any("result", res)))
}

func (s *authnSessionManagerLogging) Delete(ctx context.Context, id string) error {
	log := logger.With(s.logger,
		zap.String("service", "Delete"),
		zap.String("session_id", id),
	)
	s.logger.Debug(ctx, logger.LoggerMessageRequestStarted)

	now := time.Now()

	err := s.next.Delete(ctx, id)
	return common.LogErrorIfAny(err, ctx, now, log)
}
