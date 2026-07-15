package middleware

import (
	"context"

	cerr "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	common "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/organization/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/organization/service"

	val "github.com/go-playground/validator/v10"
)

type validation struct {
	next     service.Organizationer
	validate *val.Validate
}

func Validation(validate *val.Validate) func(service.Organizationer) service.Organizationer {
	return func(next service.Organizationer) service.Organizationer {
		return &validation{
			next:     next,
			validate: validate,
		}
	}
}

func (v validation) ListOrganizations(ctx context.Context, tenantID uint64, pagination *common.Pagination) ([]*model.Organization, *common.PaginationResult, error) {
	return v.next.ListOrganizations(ctx, tenantID, pagination)
}

func (v validation) GetOrganization(ctx context.Context, tenantID, id uint64) (*model.Organization, error) {
	return v.next.GetOrganization(ctx, tenantID, id)
}

func (v validation) GetOrganizationLogo(ctx context.Context, tenantID, id uint64) (*model.OrganizationLogo, error) {
	return v.next.GetOrganizationLogo(ctx, tenantID, id)
}

func (v validation) CreateOrganization(ctx context.Context, organization *model.Organization) (*model.Organization, error) {
	if err := v.validate.Struct(organization); err != nil {
		return nil, cerr.WrapValidationError(err)
	}
	return v.next.CreateOrganization(ctx, organization)
}

func (v validation) UpdateOrganization(ctx context.Context, organization *model.Organization) (*model.Organization, error) {
	if err := v.validate.Struct(organization); err != nil {
		return nil, cerr.WrapValidationError(err)
	}
	return v.next.UpdateOrganization(ctx, organization)
}

func (v validation) DeleteOrganization(ctx context.Context, tenantID, id uint64) error {
	return v.next.DeleteOrganization(ctx, tenantID, id)
}
