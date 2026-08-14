package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	authz "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
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
	if err := c.IsAuthorized(ctx, authz.PermCreateTermsOfUseVersion.For()); err != nil {
		return nil, err
	}
	return c.next.CreateTermsOfUseVersion(ctx, req)
}

func (c termsOfUseControllerAuthorization) UpdateTermsOfUseVersion(ctx context.Context, req *chorus.UpdateTermsOfUseVersionRequest) (*chorus.UpdateTermsOfUseVersionReply, error) {
	if err := c.IsAuthorized(ctx, authz.PermUpdateTermsOfUseVersion.For()); err != nil {
		return nil, err
	}
	return c.next.UpdateTermsOfUseVersion(ctx, req)
}

func (c termsOfUseControllerAuthorization) PublishTermsOfUseVersion(ctx context.Context, req *chorus.PublishTermsOfUseVersionRequest) (*chorus.PublishTermsOfUseVersionReply, error) {
	if err := c.IsAuthorized(ctx, authz.PermPublishTermsOfUseVersion.For()); err != nil {
		return nil, err
	}
	return c.next.PublishTermsOfUseVersion(ctx, req)
}

func (c termsOfUseControllerAuthorization) GetTermsOfUseVersion(ctx context.Context, req *chorus.GetTermsOfUseVersionRequest) (*chorus.GetTermsOfUseVersionReply, error) {
	if err := c.IsAuthorized(ctx, authz.PermGetTermsOfUseVersion.For()); err != nil {
		return nil, err
	}
	return c.next.GetTermsOfUseVersion(ctx, req)
}

func (c termsOfUseControllerAuthorization) ListTermsOfUseVersions(ctx context.Context, req *chorus.ListTermsOfUseVersionsRequest) (*chorus.ListTermsOfUseVersionsReply, error) {
	if err := c.IsAuthorized(ctx, authz.PermListTermsOfUseVersions.For()); err != nil {
		return nil, err
	}
	return c.next.ListTermsOfUseVersions(ctx, req)
}

func (c termsOfUseControllerAuthorization) GetCurrentTermsOfUseVersion(ctx context.Context, req *chorus.GetCurrentTermsOfUseVersionRequest) (*chorus.GetCurrentTermsOfUseVersionReply, error) {
	if err := c.IsAuthorized(ctx, authz.PermGetCurrentTermsOfUseVersion.For()); err != nil {
		return nil, err
	}
	return c.next.GetCurrentTermsOfUseVersion(ctx, req)
}

func (c termsOfUseControllerAuthorization) ListTermsOfUseAcceptances(ctx context.Context, req *chorus.ListTermsOfUseAcceptancesRequest) (*chorus.ListTermsOfUseAcceptancesReply, error) {
	if err := c.IsAuthorized(ctx, authz.PermListTermsOfUseAcceptances.For()); err != nil {
		return nil, err
	}
	return c.next.ListTermsOfUseAcceptances(ctx, req)
}

func (c termsOfUseControllerAuthorization) GetMyTermsOfUseStatus(ctx context.Context, req *chorus.GetMyTermsOfUseStatusRequest) (*chorus.GetMyTermsOfUseStatusReply, error) {
	if err := c.IsAuthorized(ctx, authz.PermGetMyTermsOfUseStatus.For(userFromCtx(ctx))); err != nil {
		return nil, err
	}
	return c.next.GetMyTermsOfUseStatus(ctx, req)
}

func (c termsOfUseControllerAuthorization) AcceptTermsOfUse(ctx context.Context, req *chorus.AcceptTermsOfUseRequest) (*chorus.AcceptTermsOfUseReply, error) {
	if err := c.IsAuthorized(ctx, authz.PermAcceptTermsOfUse.For(userFromCtx(ctx))); err != nil {
		return nil, err
	}
	return c.next.AcceptTermsOfUse(ctx, req)
}
