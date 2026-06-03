package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	authorization "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	authorization_service "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"
)

var _ chorus.TermsOfUseServiceServer = (*termsOfUseControllerAuthorization)(nil)

type termsOfUseControllerAuthorization struct {
	Authorization
	next chorus.TermsOfUseServiceServer
}

func TermsOfUseAuthorizing(logger *logger.ContextLogger, authorizer authorization_service.Authorizer) func(chorus.TermsOfUseServiceServer) chorus.TermsOfUseServiceServer {
	return func(next chorus.TermsOfUseServiceServer) chorus.TermsOfUseServiceServer {
		return &termsOfUseControllerAuthorization{
			Authorization: Authorization{
				logger:     logger,
				authorizer: authorizer,
			},
			next: next,
		}
	}
}

func (c termsOfUseControllerAuthorization) CreateTermsOfUseVersion(ctx context.Context, req *chorus.CreateTermsOfUseVersionRequest) (*chorus.CreateTermsOfUseVersionReply, error) {
	if err := c.IsAuthorized(ctx, authorization.PermissionCreateTermsOfUseVersion); err != nil {
		return nil, err
	}
	return c.next.CreateTermsOfUseVersion(ctx, req)
}

func (c termsOfUseControllerAuthorization) UpdateTermsOfUseVersion(ctx context.Context, req *chorus.UpdateTermsOfUseVersionRequest) (*chorus.UpdateTermsOfUseVersionReply, error) {
	if err := c.IsAuthorized(ctx, authorization.PermissionUpdateTermsOfUseVersion); err != nil {
		return nil, err
	}
	return c.next.UpdateTermsOfUseVersion(ctx, req)
}

func (c termsOfUseControllerAuthorization) PublishTermsOfUseVersion(ctx context.Context, req *chorus.PublishTermsOfUseVersionRequest) (*chorus.PublishTermsOfUseVersionReply, error) {
	if err := c.IsAuthorized(ctx, authorization.PermissionPublishTermsOfUseVersion); err != nil {
		return nil, err
	}
	return c.next.PublishTermsOfUseVersion(ctx, req)
}

func (c termsOfUseControllerAuthorization) GetTermsOfUseVersion(ctx context.Context, req *chorus.GetTermsOfUseVersionRequest) (*chorus.GetTermsOfUseVersionReply, error) {
	if err := c.IsAuthorized(ctx, authorization.PermissionGetTermsOfUseVersion); err != nil {
		return nil, err
	}
	return c.next.GetTermsOfUseVersion(ctx, req)
}

func (c termsOfUseControllerAuthorization) ListTermsOfUseVersions(ctx context.Context, req *chorus.ListTermsOfUseVersionsRequest) (*chorus.ListTermsOfUseVersionsReply, error) {
	if err := c.IsAuthorized(ctx, authorization.PermissionListTermsOfUseVersions); err != nil {
		return nil, err
	}
	return c.next.ListTermsOfUseVersions(ctx, req)
}

func (c termsOfUseControllerAuthorization) GetCurrentTermsOfUseVersion(ctx context.Context, req *chorus.GetCurrentTermsOfUseVersionRequest) (*chorus.GetCurrentTermsOfUseVersionReply, error) {
	if err := c.IsAuthorized(ctx, authorization.PermissionGetCurrentTermsOfUseVersion); err != nil {
		return nil, err
	}
	return c.next.GetCurrentTermsOfUseVersion(ctx, req)
}

func (c termsOfUseControllerAuthorization) ListTermsOfUseAcceptances(ctx context.Context, req *chorus.ListTermsOfUseAcceptancesRequest) (*chorus.ListTermsOfUseAcceptancesReply, error) {
	if err := c.IsAuthorized(ctx, authorization.PermissionListTermsOfUseAcceptances); err != nil {
		return nil, err
	}
	return c.next.ListTermsOfUseAcceptances(ctx, req)
}

func (c termsOfUseControllerAuthorization) GetMyTermsOfUseStatus(ctx context.Context, req *chorus.GetMyTermsOfUseStatusRequest) (*chorus.GetMyTermsOfUseStatusReply, error) {
	if err := c.IsAuthorized(ctx, authorization.PermissionGetMyTermsOfUseStatus, authorization.WithUserFromCtx(ctx)); err != nil {
		return nil, err
	}
	return c.next.GetMyTermsOfUseStatus(ctx, req)
}

func (c termsOfUseControllerAuthorization) AcceptTermsOfUse(ctx context.Context, req *chorus.AcceptTermsOfUseRequest) (*chorus.AcceptTermsOfUseReply, error) {
	if err := c.IsAuthorized(ctx, authorization.PermissionAcceptTermsOfUse, authorization.WithUserFromCtx(ctx)); err != nil {
		return nil, err
	}
	return c.next.AcceptTermsOfUse(ctx, req)
}
