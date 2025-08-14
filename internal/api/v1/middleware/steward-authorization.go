package middleware

import (
	"context"

	empty "google.golang.org/protobuf/types/known/emptypb"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/authorization"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

var _ chorus.StewardServiceServer = (*stewardControllerAuthorization)(nil)

type stewardControllerAuthorization struct {
	Authorization
	next chorus.StewardServiceServer
}

func StewardAuthorizing(logger *logger.ContextLogger, authorizer authorization.Authorizer) func(chorus.StewardServiceServer) chorus.StewardServiceServer {
	return func(next chorus.StewardServiceServer) chorus.StewardServiceServer {
		return &stewardControllerAuthorization{
			Authorization: Authorization{
				logger:     logger,
				authorizer: authorizer,
			},
			next: next,
		}
	}
}

func (c stewardControllerAuthorization) InitializeTenant(ctx context.Context, request *chorus.InitializeTenantRequest) (*empty.Empty, error) {
	err := c.IsAuthorized(ctx, authorization.NewPermission(authorization.PermissionInitializeTenant))
	if err != nil {
		return nil, err
	}
	return c.next.InitializeTenant(ctx, request)
}
