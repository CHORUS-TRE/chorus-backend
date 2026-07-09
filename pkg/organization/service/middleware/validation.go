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

func (v validation) ListOrganizations(ctx context.Context, req service.ListOrganizationsReq) ([]*model.Organization, *common.PaginationResult, error) {
	if err := v.validate.Struct(req); err != nil {
		return nil, nil, cerr.WrapValidationError(err)
	}
	return v.next.ListOrganizations(ctx, req)
}

func (v validation) GetOrganization(ctx context.Context, req service.GetOrganizationReq) (*model.Organization, error) {
	if err := v.validate.Struct(req); err != nil {
		return nil, cerr.WrapValidationError(err)
	}
	return v.next.GetOrganization(ctx, req)
}

func (v validation) GetOrganizationLogo(ctx context.Context, req service.GetOrganizationLogoReq) ([]byte, *string, error) {
	if err := v.validate.Struct(req); err != nil {
		return nil, nil, cerr.WrapValidationError(err)
	}
	return v.next.GetOrganizationLogo(ctx, req)
}

func (v validation) CreateOrganization(ctx context.Context, req service.CreateOrganizationReq) (*model.Organization, error) {
	if err := v.validate.Struct(req); err != nil {
		return nil, cerr.WrapValidationError(err)
	}
	if (len(req.Logo) > 0) != (req.LogoContentType != nil) {
		return nil, cerr.ErrValidation.WithMessage("Logo and LogoContentType must be provided together")
	}
	return v.next.CreateOrganization(ctx, req)
}

func (v validation) UpdateOrganization(ctx context.Context, req service.UpdateOrganizationReq) (*model.Organization, error) {
	if err := v.validate.Struct(req); err != nil {
		return nil, cerr.WrapValidationError(err)
	}
	// A non-nil Logo pointing at an empty slice means "clear the logo" and needs no
	// content type; only a logo carrying actual bytes must be paired with one.
	logoHasBytes := req.Logo != nil && len(*req.Logo) > 0
	if logoHasBytes != (req.LogoContentType != nil) {
		return nil, cerr.ErrValidation.WithMessage("Logo and LogoContentType must be provided together")
	}
	return v.next.UpdateOrganization(ctx, req)
}

func (v validation) DeleteOrganization(ctx context.Context, req service.DeleteOrganizationReq) error {
	if err := v.validate.Struct(req); err != nil {
		return cerr.WrapValidationError(err)
	}
	return v.next.DeleteOrganization(ctx, req)
}
