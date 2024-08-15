package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

type appInstanceControllerAuthorization struct {
	authorization
	next chorus.AppInstanceServiceServer
}

func AppInstanceAuthorizing(logger *logger.ContextLogger, authorizedRoles []string) func(chorus.AppInstanceServiceServer) chorus.AppInstanceServiceServer {
	return func(next chorus.AppInstanceServiceServer) chorus.AppInstanceServiceServer {
		return &appInstanceControllerAuthorization{
			authorization: authorization{
				logger:          logger,
				authorizedRoles: authorizedRoles,
			},
			next: next,
		}
	}
}

func (c appInstanceControllerAuthorization) ListAppInstances(ctx context.Context, req *chorus.ListAppInstancesRequest) (*chorus.ListAppInstancesReply, error) {
	// TODO check for permission

	err := c.IsAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.ListAppInstances(ctx, req)
}

func (c appInstanceControllerAuthorization) GetAppInstance(ctx context.Context, req *chorus.GetAppInstanceRequest) (*chorus.GetAppInstanceReply, error) {
	// TODO check for permission

	err := c.IsAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.GetAppInstance(ctx, req)
}

func (c appInstanceControllerAuthorization) CreateAppInstance(ctx context.Context, req *chorus.AppInstance) (*chorus.CreateAppInstanceReply, error) {
	// TODO check for permission

	err := c.IsAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	// nolint: staticcheck
	return c.next.CreateAppInstance(ctx, req)
}

func (c appInstanceControllerAuthorization) UpdateAppInstance(ctx context.Context, req *chorus.UpdateAppInstanceRequest) (*chorus.UpdateAppInstanceReply, error) {
	// TODO check for permission

	err := c.IsAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.UpdateAppInstance(ctx, req)
}

func (c appInstanceControllerAuthorization) DeleteAppInstance(ctx context.Context, req *chorus.DeleteAppInstanceRequest) (*chorus.DeleteAppInstanceReply, error) {
	// TODO check for permission

	err := c.IsAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.DeleteAppInstance(ctx, req)
}
