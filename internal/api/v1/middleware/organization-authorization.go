package middleware

import (
	"context"

	"google.golang.org/genproto/googleapis/api/httpbody"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	authz "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	authorization_service "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"
)

var _ chorus.OrganizationServiceServer = (*organizationControllerAuthorization)(nil)

type organizationControllerAuthorization struct {
	Authorization
	next chorus.OrganizationServiceServer
}

func OrganizationAuthorizing(logger *logger.ContextLogger, authorizer authorization_service.Authorizer) func(chorus.OrganizationServiceServer) chorus.OrganizationServiceServer {
	return func(next chorus.OrganizationServiceServer) chorus.OrganizationServiceServer {
		return &organizationControllerAuthorization{
			Authorization: Authorization{
				logger:     logger,
				authorizer: authorizer,
			},
			next: next,
		}
	}
}

func (c organizationControllerAuthorization) ListOrganizations(ctx context.Context, req *chorus.ListOrganizationsRequest) (*chorus.ListOrganizationsReply, error) {
	if err := c.IsAuthorized(ctx, authz.PermListOrganizations.For()); err != nil {
		return nil, err
	}

	return c.next.ListOrganizations(ctx, req)
}

func (c organizationControllerAuthorization) GetOrganization(ctx context.Context, req *chorus.GetOrganizationRequest) (*chorus.GetOrganizationReply, error) {
	if err := c.IsAuthorized(ctx, authz.PermGetOrganization.For()); err != nil {
		return nil, err
	}

	return c.next.GetOrganization(ctx, req)
}

// GetOrganizationLogo reuses PermissionGetOrganization
func (c organizationControllerAuthorization) GetOrganizationLogo(ctx context.Context, req *chorus.GetOrganizationLogoRequest) (*httpbody.HttpBody, error) {
	if err := c.IsAuthorized(ctx, authz.PermGetOrganization.For()); err != nil {
		return nil, err
	}

	return c.next.GetOrganizationLogo(ctx, req)
}

func (c organizationControllerAuthorization) CreateOrganization(ctx context.Context, req *chorus.Organization) (*chorus.CreateOrganizationReply, error) {
	if err := c.IsAuthorized(ctx, authz.PermCreateOrganization.For()); err != nil {
		return nil, err
	}

	return c.next.CreateOrganization(ctx, req)
}

func (c organizationControllerAuthorization) UpdateOrganization(ctx context.Context, req *chorus.Organization) (*chorus.UpdateOrganizationReply, error) {
	if err := c.IsAuthorized(ctx, authz.PermUpdateOrganization.For()); err != nil {
		return nil, err
	}

	return c.next.UpdateOrganization(ctx, req)
}

func (c organizationControllerAuthorization) DeleteOrganization(ctx context.Context, req *chorus.DeleteOrganizationRequest) (*chorus.DeleteOrganizationReply, error) {
	if err := c.IsAuthorized(ctx, authz.PermDeleteOrganization.For()); err != nil {
		return nil, err
	}

	return c.next.DeleteOrganization(ctx, req)
}
