package middleware

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/terms-of-use/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/terms-of-use/service"
)

type termsOfUseStorageLogging struct {
	logger *logger.ContextLogger
	next   service.TermsOfUseStore
}

func Logging(log *logger.ContextLogger) func(service.TermsOfUseStore) service.TermsOfUseStore {
	return func(next service.TermsOfUseStore) service.TermsOfUseStore {
		return &termsOfUseStorageLogging{logger: log, next: next}
	}
}

func (s *termsOfUseStorageLogging) CreateTermsOfUseVersion(ctx context.Context, tenantID uint64, content string) (*model.TermsOfUseVersion, error) {
	s.logger.Debug(ctx, logger.LoggerMessageRequestStarted)
	now := time.Now()

	res, err := s.next.CreateTermsOfUseVersion(ctx, tenantID, content)
	if err != nil {
		s.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("tenant_id", tenantID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, ms(now)),
		)
		return nil, err
	}

	s.logger.Debug(ctx, logger.LoggerMessageRequestCompleted, zap.Float64(logger.LoggerKeyElapsedMs, ms(now)))
	return res, nil
}

func (s *termsOfUseStorageLogging) UpdateTermsOfUseVersion(ctx context.Context, tenantID, id uint64, content string) (*model.TermsOfUseVersion, error) {
	s.logger.Debug(ctx, logger.LoggerMessageRequestStarted)
	now := time.Now()

	res, err := s.next.UpdateTermsOfUseVersion(ctx, tenantID, id, content)
	if err != nil {
		s.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("tenant_id", tenantID),
			zap.Uint64("id", id),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, ms(now)),
		)
		return nil, err
	}

	s.logger.Debug(ctx, logger.LoggerMessageRequestCompleted, zap.Float64(logger.LoggerKeyElapsedMs, ms(now)))
	return res, nil
}

func (s *termsOfUseStorageLogging) PublishTermsOfUseVersion(ctx context.Context, tenantID, id uint64) (*model.TermsOfUseVersion, error) {
	s.logger.Debug(ctx, logger.LoggerMessageRequestStarted)
	now := time.Now()

	res, err := s.next.PublishTermsOfUseVersion(ctx, tenantID, id)
	if err != nil {
		s.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("tenant_id", tenantID),
			zap.Uint64("id", id),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, ms(now)),
		)
		return nil, err
	}

	s.logger.Debug(ctx, logger.LoggerMessageRequestCompleted, zap.Float64(logger.LoggerKeyElapsedMs, ms(now)))
	return res, nil
}

func (s *termsOfUseStorageLogging) GetTermsOfUseVersion(ctx context.Context, tenantID, versionID uint64) (*model.TermsOfUseVersion, error) {
	s.logger.Debug(ctx, logger.LoggerMessageRequestStarted)
	now := time.Now()

	res, err := s.next.GetTermsOfUseVersion(ctx, tenantID, versionID)
	if err != nil {
		s.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("tenant_id", tenantID),
			zap.Uint64("version_id", versionID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, ms(now)),
		)
		return nil, err
	}

	s.logger.Debug(ctx, logger.LoggerMessageRequestCompleted, zap.Float64(logger.LoggerKeyElapsedMs, ms(now)))
	return res, nil
}

func (s *termsOfUseStorageLogging) ListTermsOfUseVersions(ctx context.Context, tenantID uint64, pagination *common_model.Pagination) ([]*model.TermsOfUseVersion, *common_model.PaginationResult, error) {
	s.logger.Debug(ctx, logger.LoggerMessageRequestStarted)
	now := time.Now()

	res, pag, err := s.next.ListTermsOfUseVersions(ctx, tenantID, pagination)
	if err != nil {
		s.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("tenant_id", tenantID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, ms(now)),
		)
		return nil, nil, err
	}

	s.logger.Debug(ctx, logger.LoggerMessageRequestCompleted, zap.Float64(logger.LoggerKeyElapsedMs, ms(now)))
	return res, pag, nil
}

func (s *termsOfUseStorageLogging) GetCurrentTermsOfUseVersion(ctx context.Context, tenantID uint64) (*model.TermsOfUseVersion, error) {
	s.logger.Debug(ctx, logger.LoggerMessageRequestStarted)
	now := time.Now()

	res, err := s.next.GetCurrentTermsOfUseVersion(ctx, tenantID)
	if err != nil {
		s.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("tenant_id", tenantID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, ms(now)),
		)
		return nil, err
	}

	s.logger.Debug(ctx, logger.LoggerMessageRequestCompleted, zap.Float64(logger.LoggerKeyElapsedMs, ms(now)))
	return res, nil
}

func (s *termsOfUseStorageLogging) ListTermsOfUseAcceptances(ctx context.Context, tenantID, userID uint64) ([]*model.TermsOfUseAcceptance, error) {
	s.logger.Debug(ctx, logger.LoggerMessageRequestStarted)
	now := time.Now()

	res, err := s.next.ListTermsOfUseAcceptances(ctx, tenantID, userID)
	if err != nil {
		s.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("tenant_id", tenantID),
			zap.Uint64("user_id", userID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, ms(now)),
		)
		return nil, err
	}

	s.logger.Debug(ctx, logger.LoggerMessageRequestCompleted, zap.Float64(logger.LoggerKeyElapsedMs, ms(now)))
	return res, nil
}

func (s *termsOfUseStorageLogging) GetMyTermsOfUseStatus(ctx context.Context, tenantID, userID uint64) (bool, error) {
	s.logger.Debug(ctx, logger.LoggerMessageRequestStarted)
	now := time.Now()

	res, err := s.next.GetMyTermsOfUseStatus(ctx, tenantID, userID)
	if err != nil {
		s.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("tenant_id", tenantID),
			zap.Uint64("user_id", userID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, ms(now)),
		)
		return false, err
	}

	s.logger.Debug(ctx, logger.LoggerMessageRequestCompleted, zap.Float64(logger.LoggerKeyElapsedMs, ms(now)))
	return res, nil
}

func (s *termsOfUseStorageLogging) AcceptTermsOfUse(ctx context.Context, tenantID, userID, versionID uint64) (*model.TermsOfUseAcceptance, error) {
	s.logger.Debug(ctx, logger.LoggerMessageRequestStarted)
	now := time.Now()

	res, err := s.next.AcceptTermsOfUse(ctx, tenantID, userID, versionID)
	if err != nil {
		s.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("tenant_id", tenantID),
			zap.Uint64("user_id", userID),
			zap.Uint64("version_id", versionID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, ms(now)),
		)
		return nil, err
	}

	s.logger.Debug(ctx, logger.LoggerMessageRequestCompleted, zap.Float64(logger.LoggerKeyElapsedMs, ms(now)))
	return res, nil
}

func ms(start time.Time) float64 {
	return float64(time.Since(start).Nanoseconds()) / 1_000_000.0
}
