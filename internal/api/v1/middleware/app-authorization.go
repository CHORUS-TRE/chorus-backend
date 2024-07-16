package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

type appControllerAuthorization struct {
	authorization
	next chorus.AppServiceServer
}

func AppAuthorizing(logger *logger.ContextLogger, authorizedRoles []string) func(chorus.AppServiceServer) chorus.AppServiceServer {
	return func(next chorus.AppServiceServer) chorus.AppServiceServer {
		return &appControllerAuthorization{
			authorization: authorization{
				logger:          logger,
				authorizedRoles: authorizedRoles,
			},
			next: next,
		}
	}
}

func (c appControllerAuthorization) ListApps(ctx context.Context, req *chorus.ListAppsRequest) (*chorus.ListAppsReply, error) {
	// TODO check for permission

	err := c.isAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.ListApps(ctx, req)
}

func (c appControllerAuthorization) GetApp(ctx context.Context, req *chorus.GetAppRequest) (*chorus.GetAppReply, error) {
	// TODO check for permission

	err := c.isAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.GetApp(ctx, req)
}

func (c appControllerAuthorization) CreateApp(ctx context.Context, req *chorus.App) (*chorus.CreateAppReply, error) {
	// TODO check for permission

	err := c.isAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	// nolint: staticcheck
	return c.next.CreateApp(ctx, req)
}

func (c appControllerAuthorization) UpdateApp(ctx context.Context, req *chorus.UpdateAppRequest) (*chorus.UpdateAppReply, error) {
	// TODO check for permission

	err := c.isAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.UpdateApp(ctx, req)
}

func (c appControllerAuthorization) DeleteApp(ctx context.Context, req *chorus.DeleteAppRequest) (*chorus.DeleteAppReply, error) {
	// TODO check for permission

	err := c.isAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.DeleteApp(ctx, req)
}
