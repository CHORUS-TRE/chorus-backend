package v1

import (
	"fmt"

	"context"

	"google.golang.org/genproto/googleapis/api/httpbody"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/converter"
	cerr "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/organization/service"
)

// The browser caches the logo for 24h and serves it from cache without a network request.
const organizationLogoCacheControl = "max-age=86400"

var _ chorus.OrganizationServiceServer = (*OrganizationController)(nil)

// OrganizationController is the organization service controller handler.
type OrganizationController struct {
	organization service.Organizationer
}

func NewOrganizationController(organization service.Organizationer) OrganizationController {
	return OrganizationController{organization: organization}
}

func (c OrganizationController) ListOrganizations(ctx context.Context, req *chorus.ListOrganizationsRequest) (*chorus.ListOrganizationsReply, error) {
	if req == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Could not extract tenant ID from token")
	}

	pagination := converter.PaginationToBusiness(req.Pagination)

	organizations, paginationRes, err := c.organization.ListOrganizations(ctx, tenantID, &pagination)
	if err != nil {
		return nil, err
	}

	protoOrganizations := make([]*chorus.OrganizationSummary, 0, len(organizations))
	for _, organization := range organizations {
		proto, err := converter.OrganizationSummaryFromBusiness(organization)
		if err != nil {
			return nil, cerr.ErrConversion.Wrap(err, "Failed to convert organization")
		}
		protoOrganizations = append(protoOrganizations, proto)
	}

	return &chorus.ListOrganizationsReply{
		Result:     &chorus.ListOrganizationsResult{Organizations: protoOrganizations},
		Pagination: converter.PaginationResultFromBusiness(paginationRes),
	}, nil
}

func (c OrganizationController) GetOrganization(ctx context.Context, req *chorus.GetOrganizationRequest) (*chorus.GetOrganizationReply, error) {
	if req == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Could not extract tenant ID from token")
	}

	organization, err := c.organization.GetOrganization(ctx, tenantID, req.Id)
	if err != nil {
		return nil, err
	}

	proto, err := converter.OrganizationSummaryFromBusiness(organization)
	if err != nil {
		return nil, cerr.ErrConversion.Wrap(err, "Failed to convert organization")
	}

	return &chorus.GetOrganizationReply{Result: &chorus.GetOrganizationResult{Organization: proto}}, nil
}

func (c OrganizationController) GetOrganizationLogo(ctx context.Context, req *chorus.GetOrganizationLogoRequest) (*httpbody.HttpBody, error) {
	if req == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Could not extract tenant ID from token")
	}

	logo, err := c.organization.GetOrganizationLogo(ctx, tenantID, req.Id)
	if err != nil {
		return nil, err
	}
	if logo == nil {
		return nil, cerr.ErrNotFound.WithMessage(fmt.Sprintf("organization %d has no logo", req.Id))
	}

	if err := grpc.SetHeader(ctx, metadata.Pairs("Cache-Control", organizationLogoCacheControl)); err != nil {
		return nil, cerr.ErrInternal.Wrap(err, "unable to set Cache-Control header")
	}

	return &httpbody.HttpBody{
		ContentType: logo.LogoContentType,
		Data:        logo.Logo,
	}, nil
}

func (c OrganizationController) CreateOrganization(ctx context.Context, req *chorus.Organization) (*chorus.CreateOrganizationReply, error) {
	if req == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Could not extract tenant ID from token")
	}

	organization := converter.OrganizationToBusiness(req)
	organization.TenantID = tenantID

	createdOrganization, err := c.organization.CreateOrganization(ctx, organization)
	if err != nil {
		return nil, err
	}

	proto, err := converter.OrganizationSummaryFromBusiness(createdOrganization)
	if err != nil {
		return nil, cerr.ErrConversion.Wrap(err, "Failed to convert organization")
	}

	return &chorus.CreateOrganizationReply{Result: &chorus.CreateOrganizationResult{Organization: proto}}, nil
}

func (c OrganizationController) UpdateOrganization(ctx context.Context, req *chorus.Organization) (*chorus.UpdateOrganizationReply, error) {
	if req == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Could not extract tenant ID from token")
	}

	organization := converter.OrganizationToBusiness(req)
	organization.TenantID = tenantID

	updatedOrganization, err := c.organization.UpdateOrganization(ctx, organization)
	if err != nil {
		return nil, err
	}

	proto, err := converter.OrganizationSummaryFromBusiness(updatedOrganization)
	if err != nil {
		return nil, cerr.ErrConversion.Wrap(err, "Failed to convert organization")
	}

	return &chorus.UpdateOrganizationReply{Result: &chorus.UpdateOrganizationResult{Organization: proto}}, nil
}

func (c OrganizationController) DeleteOrganization(ctx context.Context, req *chorus.DeleteOrganizationRequest) (*chorus.DeleteOrganizationReply, error) {
	if req == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Empty request")
	}

	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Could not extract tenant ID from token")
	}

	if err := c.organization.DeleteOrganization(ctx, tenantID, req.Id); err != nil {
		return nil, err
	}

	return &chorus.DeleteOrganizationReply{Result: &chorus.DeleteOrganizationResult{}}, nil
}
