package middleware

import (
	"context"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/common"
	tenant_model "github.com/CHORUS-TRE/chorus-backend/pkg/tenant/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/tenant/service"
)

type tenantServiceLogging struct {
	logger *logger.ContextLogger
	next   service.Tenanter
}

func Logging(log *logger.ContextLogger) func(tenanter service.Tenanter) service.Tenanter {
	l := logger.With(log, logger.WithLayerField("service"))
	return func(next service.Tenanter) service.Tenanter {
		return &tenantServiceLogging{
			logger: l,
			next:   next,
		}
	}
}

func (l tenantServiceLogging) CreateTenant(ctx context.Context, name string) (*tenant_model.Tenant, error) {
	now := time.Now()

	tenant, err := l.next.CreateTenant(ctx, name)

	if tenant != nil {
		log := logger.With(l.logger, logger.WithTenantIDField(tenant.ID))
		return tenant, common.LogErrorIfAny(err, ctx, now, log)
	}

	return tenant, common.LogErrorIfAny(err, ctx, now, l.logger)
}

func (l tenantServiceLogging) GetTenantByName(ctx context.Context, name string) (*tenant_model.Tenant, error) {
	now := time.Now()

	tenant, err := l.next.GetTenantByName(ctx, name)

	if tenant != nil {
		log := logger.With(l.logger, logger.WithTenantIDField(tenant.ID))
		return tenant, common.LogErrorIfAny(err, ctx, now, log)
	}

	return tenant, common.LogErrorIfAny(err, ctx, now, l.logger)
}
