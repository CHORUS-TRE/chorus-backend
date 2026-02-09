package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	authorization_service "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"
)

var _ chorus.AuthorizationServiceServer = (*authorizationControllerAuthorization)(nil)

type authorizationControllerAuthorization struct {
	Authorization
	next chorus.AuthorizationServiceServer
}

func AuthorizationAuthorizing(logger *logger.ContextLogger, authorizer authorization_service.Authorizer, cfg config.Config, refresher Refresher) func(chorus.AuthorizationServiceServer) chorus.AuthorizationServiceServer {
	return func(next chorus.AuthorizationServiceServer) chorus.AuthorizationServiceServer {
		return &authorizationControllerAuthorization{
			Authorization: Authorization{
				logger:     logger,
				authorizer: authorizer,
				cfg:        cfg,
				refresher:  refresher,
			},
			next: next,
		}
	}
}

func (c authorizationControllerAuthorization) ListRoles(ctx context.Context, req *chorus.ListRolesRequest) (*chorus.ListRolesReply, error) {
	return c.next.ListRoles(ctx, req)
}

func (c authorizationControllerAuthorization) ListPermissions(ctx context.Context, req *chorus.ListPermissionsRequest) (*chorus.ListPermissionsReply, error) {
	return c.next.ListPermissions(ctx, req)
}
