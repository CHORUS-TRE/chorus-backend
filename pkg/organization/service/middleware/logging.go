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

func (c organizationServiceLogging) ListOrganizations(ctx context.Context, tenantID uint64, pagination *common.Pagination) ([]*model.Organization, *common.PaginationResult, error) {
	now := time.Now()

	organizations, paginationRes, err := c.next.ListOrganizations(ctx, tenantID, pagination)
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

func (c organizationServiceLogging) GetOrganization(ctx context.Context, tenantID, id uint64) (*model.Organization, error) {
	now := time.Now()

	res, err := c.next.GetOrganization(ctx, tenantID, id)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("organization_id", id),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Info(ctx, "request completed",
		zap.Uint64("organization_id", id),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c organizationServiceLogging) GetOrganizationLogo(ctx context.Context, tenantID, id uint64) (*model.OrganizationLogo, error) {
	now := time.Now()

	logo, err := c.next.GetOrganizationLogo(ctx, tenantID, id)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("organization_id", id),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Info(ctx, "request completed",
		zap.Uint64("organization_id", id),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return logo, nil
}

func (c organizationServiceLogging) CreateOrganization(ctx context.Context, organization *model.Organization) (*model.Organization, error) {
	now := time.Now()

	res, err := c.next.CreateOrganization(ctx, organization)
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

func (c organizationServiceLogging) UpdateOrganization(ctx context.Context, organization *model.Organization) (*model.Organization, error) {
	now := time.Now()

	res, err := c.next.UpdateOrganization(ctx, organization)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("organization_id", organization.ID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Info(ctx, "request completed",
		zap.Uint64("organization_id", organization.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c organizationServiceLogging) DeleteOrganization(ctx context.Context, tenantID, id uint64) error {
	now := time.Now()

	err := c.next.DeleteOrganization(ctx, tenantID, id)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("organization_id", id),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return err
	}

	c.logger.Info(ctx, "request completed",
		zap.Uint64("organization_id", id),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}
