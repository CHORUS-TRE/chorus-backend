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

func (l tenantServiceLogging) CreateTenant(ctx context.Context, tenantID uint64, name string) error {
	now := time.Now()

	log := logger.With(l.logger,
		logger.WithTenantIDField(tenantID),
	)

	err := l.next.CreateTenant(ctx, tenantID, name)

	return common.LogErrorIfAny(err, ctx, now, log)
}

func (l tenantServiceLogging) GetTenant(ctx context.Context, tenantID uint64) (*tenant_model.Tenant, error) {
	now := time.Now()

	log := logger.With(l.logger,
		logger.WithTenantIDField(tenantID),
	)

	tenant, err := l.next.GetTenant(ctx, tenantID)

	return tenant, common.LogErrorIfAny(err, ctx, now, log)
}
