package middleware

import (
	"context"

	tenant_model "github.com/CHORUS-TRE/chorus-backend/pkg/tenant/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/tenant/service"

	val "github.com/go-playground/validator/v10"
)

type validation struct {
	next     service.Tenanter
	validate *val.Validate
}

func Validation(validate *val.Validate) func(service.Tenanter) service.Tenanter {
	return func(next service.Tenanter) service.Tenanter {
		return &validation{
			next:     next,
			validate: validate,
		}
	}
}

func (v validation) CreateTenant(ctx context.Context, tenantID uint64, name string) error {
	return v.next.CreateTenant(ctx, tenantID, name)
}

func (v validation) GetTenant(ctx context.Context, tenantID uint64) (*tenant_model.Tenant, error) {
	return v.next.GetTenant(ctx, tenantID)
}
