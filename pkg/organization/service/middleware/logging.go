package middleware

import (
	"context"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	common "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/organization/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/organization/service"

	"go.uber.org/zap"
)

type organizationServiceLogging struct {
	logger *logger.ContextLogger
	next   service.Organizationer
}

func Logging(logger *logger.ContextLogger) func(service.Organizationer) service.Organizationer {
	return func(next service.Organizationer) service.Organizationer {
		return &organizationServiceLogging{
			logger: logger,
			next:   next,
		}
	}
}

func (c organizationServiceLogging) ListOrganizations(ctx context.Context, req service.ListOrganizationsReq) ([]*model.Organization, *common.PaginationResult, error) {
	now := time.Now()

	organizations, paginationRes, err := c.next.ListOrganizations(ctx, req)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, nil, err
	}

	c.logger.Info(ctx, "request completed",
		zap.Int("num_organizations", len(organizations)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return organizations, paginationRes, nil
}

func (c organizationServiceLogging) GetOrganization(ctx context.Context, req service.GetOrganizationReq) (*model.Organization, error) {
	now := time.Now()

	res, err := c.next.GetOrganization(ctx, req)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("organization_id", req.ID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Info(ctx, "request completed",
		zap.Uint64("organization_id", req.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c organizationServiceLogging) GetOrganizationLogo(ctx context.Context, req service.GetOrganizationLogoReq) ([]byte, *string, error) {
	now := time.Now()

	logo, contentType, err := c.next.GetOrganizationLogo(ctx, req)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("organization_id", req.ID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, nil, err
	}

	c.logger.Info(ctx, "request completed",
		zap.Uint64("organization_id", req.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return logo, contentType, nil
}

func (c organizationServiceLogging) CreateOrganization(ctx context.Context, req service.CreateOrganizationReq) (*model.Organization, error) {
	now := time.Now()

	res, err := c.next.CreateOrganization(ctx, req)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Info(ctx, "request completed",
		zap.Uint64("organization_id", res.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c organizationServiceLogging) UpdateOrganization(ctx context.Context, req service.UpdateOrganizationReq) (*model.Organization, error) {
	now := time.Now()

	res, err := c.next.UpdateOrganization(ctx, req)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("organization_id", req.ID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Info(ctx, "request completed",
		zap.Uint64("organization_id", req.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c organizationServiceLogging) DeleteOrganization(ctx context.Context, req service.DeleteOrganizationReq) error {
	now := time.Now()

	err := c.next.DeleteOrganization(ctx, req)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("organization_id", req.ID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}

	c.logger.Info(ctx, "request completed",
		zap.Uint64("organization_id", req.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}
