package middleware

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/terms-of-use/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/terms-of-use/service"
)

type termsOfUseServiceLogging struct {
	logger *logger.ContextLogger
	next   service.TermsOfUseer
}

func Logging(log *logger.ContextLogger) func(service.TermsOfUseer) service.TermsOfUseer {
	return func(next service.TermsOfUseer) service.TermsOfUseer {
		return &termsOfUseServiceLogging{logger: log, next: next}
	}
}

func (c *termsOfUseServiceLogging) CreateTermsOfUseVersion(ctx context.Context, tenantID uint64, content string) (*model.TermsOfUseVersion, error) {
	now := time.Now()

	res, err := c.next.CreateTermsOfUseVersion(ctx, tenantID, content)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("tenant_id", tenantID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, ms(now)),
		)
		return nil, fmt.Errorf("unable to create terms of use version: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted, zap.Float64(logger.LoggerKeyElapsedMs, ms(now)))
	return res, nil
}

func (c *termsOfUseServiceLogging) UpdateTermsOfUseVersion(ctx context.Context, tenantID, versionID uint64, content string) (*model.TermsOfUseVersion, error) {
	now := time.Now()

	res, err := c.next.UpdateTermsOfUseVersion(ctx, tenantID, versionID, content)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("tenant_id", tenantID),
			zap.Uint64("version_id", versionID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, ms(now)),
		)
		return nil, fmt.Errorf("unable to update terms of use version: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted, zap.Float64(logger.LoggerKeyElapsedMs, ms(now)))
	return res, nil
}

func (c *termsOfUseServiceLogging) PublishTermsOfUseVersion(ctx context.Context, tenantID, versionID uint64) (*model.TermsOfUseVersion, error) {
	now := time.Now()

	res, err := c.next.PublishTermsOfUseVersion(ctx, tenantID, versionID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("tenant_id", tenantID),
			zap.Uint64("version_id", versionID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, ms(now)),
		)
		return nil, fmt.Errorf("unable to publish terms of use version: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted, zap.Float64(logger.LoggerKeyElapsedMs, ms(now)))
	return res, nil
}

func (c *termsOfUseServiceLogging) GetTermsOfUseVersion(ctx context.Context, tenantID, versionID uint64) (*model.TermsOfUseVersion, error) {
	now := time.Now()

	res, err := c.next.GetTermsOfUseVersion(ctx, tenantID, versionID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("tenant_id", tenantID),
			zap.Uint64("version_id", versionID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, ms(now)),
		)
		return nil, fmt.Errorf("unable to get terms of use version: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted, zap.Float64(logger.LoggerKeyElapsedMs, ms(now)))
	return res, nil
}

func (c *termsOfUseServiceLogging) ListTermsOfUseVersions(ctx context.Context, tenantID uint64, pagination *common_model.Pagination) ([]*model.TermsOfUseVersion, *common_model.PaginationResult, error) {
	now := time.Now()

	res, pag, err := c.next.ListTermsOfUseVersions(ctx, tenantID, pagination)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("tenant_id", tenantID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, ms(now)),
		)
		return nil, nil, fmt.Errorf("unable to list terms of use versions: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted, zap.Float64(logger.LoggerKeyElapsedMs, ms(now)))
	return res, pag, nil
}

func (c *termsOfUseServiceLogging) GetCurrentTermsOfUseVersion(ctx context.Context, tenantID uint64) (*model.TermsOfUseVersion, error) {
	now := time.Now()

	res, err := c.next.GetCurrentTermsOfUseVersion(ctx, tenantID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("tenant_id", tenantID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, ms(now)),
		)
		return nil, fmt.Errorf("unable to get current terms of use version: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted, zap.Float64(logger.LoggerKeyElapsedMs, ms(now)))
	return res, nil
}

func (c *termsOfUseServiceLogging) ListTermsOfUseAcceptances(ctx context.Context, tenantID, userID uint64) ([]*model.TermsOfUseAcceptance, error) {
	now := time.Now()

	res, err := c.next.ListTermsOfUseAcceptances(ctx, tenantID, userID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("tenant_id", tenantID),
			zap.Uint64("user_id", userID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, ms(now)),
		)
		return nil, fmt.Errorf("unable to list terms of use acceptances: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted, zap.Float64(logger.LoggerKeyElapsedMs, ms(now)))
	return res, nil
}

func (c *termsOfUseServiceLogging) GetMyTermsOfUseStatus(ctx context.Context, tenantID, userID uint64) (bool, error) {
	now := time.Now()

	res, err := c.next.GetMyTermsOfUseStatus(ctx, tenantID, userID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("tenant_id", tenantID),
			zap.Uint64("user_id", userID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, ms(now)),
		)
		return false, fmt.Errorf("unable to get terms of use status: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted, zap.Float64(logger.LoggerKeyElapsedMs, ms(now)))
	return res, nil
}

func (c *termsOfUseServiceLogging) AcceptTermsOfUse(ctx context.Context, tenantID, userID uint64) (*model.TermsOfUseAcceptance, error) {
	now := time.Now()

	res, err := c.next.AcceptTermsOfUse(ctx, tenantID, userID)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("tenant_id", tenantID),
			zap.Uint64("user_id", userID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, ms(now)),
		)
		return nil, fmt.Errorf("unable to accept terms of use: %w", err)
	}

	c.logger.Info(ctx, logger.LoggerMessageRequestCompleted, zap.Float64(logger.LoggerKeyElapsedMs, ms(now)))
	return res, nil
}

func ms(start time.Time) float64 {
	return float64(time.Since(start).Nanoseconds()) / 1_000_000.0
}
