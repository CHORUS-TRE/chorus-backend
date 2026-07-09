package middleware

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	common "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/organization/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/organization/service"
)

type organizationStorageLogging struct {
	logger *logger.ContextLogger
	next   service.OrganizationStore
}

func Logging(logger *logger.ContextLogger) func(service.OrganizationStore) service.OrganizationStore {
	return func(next service.OrganizationStore) service.OrganizationStore {
		return &organizationStorageLogging{
			logger: logger,
			next:   next,
		}
	}
}

func (c organizationStorageLogging) ListOrganizations(ctx context.Context, tenantID uint64, pagination *common.Pagination) ([]*model.Organization, *common.PaginationResult, error) {
	c.logger.Debug(ctx, logger.LoggerMessageRequestStarted)
	now := time.Now()

	organizations, paginationRes, err := c.next.ListOrganizations(ctx, tenantID, pagination)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("tenant_id", tenantID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, nil, err
	}

	c.logger.Debug(ctx, logger.LoggerMessageRequestCompleted,
		zap.Int("num_organizations", len(organizations)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return organizations, paginationRes, nil
}

func (c organizationStorageLogging) GetOrganization(ctx context.Context, tenantID, id uint64) (*model.Organization, error) {
	c.logger.Debug(ctx, logger.LoggerMessageRequestStarted)
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

	c.logger.Debug(ctx, logger.LoggerMessageRequestCompleted,
		zap.Uint64("organization_id", id),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c organizationStorageLogging) GetOrganizationLogo(ctx context.Context, tenantID, id uint64) ([]byte, *string, error) {
	c.logger.Debug(ctx, logger.LoggerMessageRequestStarted)
	now := time.Now()

	logo, contentType, err := c.next.GetOrganizationLogo(ctx, tenantID, id)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("organization_id", id),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, nil, err
	}

	c.logger.Debug(ctx, logger.LoggerMessageRequestCompleted,
		zap.Uint64("organization_id", id),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return logo, contentType, nil
}

func (c organizationStorageLogging) CreateOrganization(ctx context.Context, tenantID uint64, organization *model.Organization) (*model.Organization, error) {
	c.logger.Debug(ctx, logger.LoggerMessageRequestStarted)
	now := time.Now()

	res, err := c.next.CreateOrganization(ctx, tenantID, organization)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("tenant_id", tenantID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, logger.LoggerMessageRequestCompleted,
		zap.Uint64("organization_id", res.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c organizationStorageLogging) UpdateOrganization(ctx context.Context, tenantID uint64, organization *model.Organization, updateLogo bool) (*model.Organization, error) {
	c.logger.Debug(ctx, logger.LoggerMessageRequestStarted)
	now := time.Now()

	res, err := c.next.UpdateOrganization(ctx, tenantID, organization, updateLogo)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Uint64("organization_id", organization.ID),
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Debug(ctx, logger.LoggerMessageRequestCompleted,
		zap.Uint64("organization_id", res.ID),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return res, nil
}

func (c organizationStorageLogging) DeleteOrganization(ctx context.Context, tenantID, id uint64) error {
	c.logger.Debug(ctx, logger.LoggerMessageRequestStarted)
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

	c.logger.Debug(ctx, logger.LoggerMessageRequestCompleted,
		zap.Uint64("organization_id", id),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return nil
}
