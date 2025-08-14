package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/authorization"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

var _ chorus.AppInstanceServiceServer = (*appInstanceControllerAuthorization)(nil)

type appInstanceControllerAuthorization struct {
	Authorization
	authorizer authorization.Authorizer
	next       chorus.AppInstanceServiceServer
}

func AppInstanceAuthorizing(logger *logger.ContextLogger, authorizer authorization.Authorizer) func(chorus.AppInstanceServiceServer) chorus.AppInstanceServiceServer {
	return func(next chorus.AppInstanceServiceServer) chorus.AppInstanceServiceServer {
		return &appInstanceControllerAuthorization{
			Authorization: Authorization{
				logger:     logger,
				authorizer: authorizer,
			},
			next: next,
		}
	}
}

func (c appInstanceControllerAuthorization) ListAppInstances(ctx context.Context, req *chorus.ListAppInstancesRequest) (*chorus.ListAppInstancesReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionListAppInstances)
	if err != nil {
		return nil, err
	}

	return c.next.ListAppInstances(ctx, req)
}

func (c appInstanceControllerAuthorization) GetAppInstance(ctx context.Context, req *chorus.GetAppInstanceRequest) (*chorus.GetAppInstanceReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionGetAppInstance)
	if err != nil {
		return nil, err
	}

	return c.next.GetAppInstance(ctx, req)
}

func (c appInstanceControllerAuthorization) CreateAppInstance(ctx context.Context, req *chorus.AppInstance) (*chorus.CreateAppInstanceReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionCreateAppInstance)
	if err != nil {
		return nil, err
	}

	return c.next.CreateAppInstance(ctx, req)
}

func (c appInstanceControllerAuthorization) UpdateAppInstance(ctx context.Context, req *chorus.AppInstance) (*chorus.UpdateAppInstanceReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionUpdateAppInstance)
	if err != nil {
		return nil, err
	}

	return c.next.UpdateAppInstance(ctx, req)
}

func (c appInstanceControllerAuthorization) DeleteAppInstance(ctx context.Context, req *chorus.DeleteAppInstanceRequest) (*chorus.DeleteAppInstanceReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionDeleteAppInstance)
	if err != nil {
		return nil, err
	}

	return c.next.DeleteAppInstance(ctx, req)
}
