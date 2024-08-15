package middleware

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

type stewardControllerAuthorization struct {
	authorization
	next chorus.StewardServiceServer
}

func StewardAuthorizing(logger *logger.ContextLogger, authorizedRoles []string) func(chorus.StewardServiceServer) chorus.StewardServiceServer {
	return func(next chorus.StewardServiceServer) chorus.StewardServiceServer {
		return &stewardControllerAuthorization{
			authorization: authorization{
				logger:          logger,
				authorizedRoles: authorizedRoles,
			},
			next: next,
		}
	}
}

func (c stewardControllerAuthorization) InitializeTenant(ctx context.Context, request *chorus.InitializeTenantRequest) (*empty.Empty, error) {
	err := c.IsAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	return c.next.InitializeTenant(ctx, request)
}
