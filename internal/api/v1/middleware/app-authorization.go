package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/authorization"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

var _ chorus.AppServiceServer = (*appControllerAuthorization)(nil)

type appControllerAuthorization struct {
	Authorization
	next chorus.AppServiceServer
}

func AppAuthorizing(logger *logger.ContextLogger, authorizer authorization.Authorizer) func(chorus.AppServiceServer) chorus.AppServiceServer {
	return func(next chorus.AppServiceServer) chorus.AppServiceServer {
		return &appControllerAuthorization{
			Authorization: Authorization{
				logger:     logger,
				authorizer: authorizer,
			},
			next: next,
		}
	}
}

func (c appControllerAuthorization) ListApps(ctx context.Context, req *chorus.ListAppsRequest) (*chorus.ListAppsReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionListApps)
	if err != nil {
		return nil, err
	}

	return c.next.ListApps(ctx, req)
}

func (c appControllerAuthorization) GetApp(ctx context.Context, req *chorus.GetAppRequest) (*chorus.GetAppReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionGetApp)
	if err != nil {
		return nil, err
	}

	return c.next.GetApp(ctx, req)
}

func (c appControllerAuthorization) CreateApp(ctx context.Context, req *chorus.App) (*chorus.CreateAppReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionCreateApp)
	if err != nil {
		return nil, err
	}

	return c.next.CreateApp(ctx, req)
}

func (c appControllerAuthorization) BulkCreateApps(ctx context.Context, req *chorus.BulkCreateAppsRequest) (*chorus.BulkCreateAppsReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionCreateApp)
	if err != nil {
		return nil, err
	}

	return c.next.BulkCreateApps(ctx, req)
}

func (c appControllerAuthorization) UpdateApp(ctx context.Context, req *chorus.App) (*chorus.UpdateAppReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionUpdateApp)
	if err != nil {
		return nil, err
	}

	return c.next.UpdateApp(ctx, req)
}

func (c appControllerAuthorization) DeleteApp(ctx context.Context, req *chorus.DeleteAppRequest) (*chorus.DeleteAppReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionDeleteApp)
	if err != nil {
		return nil, err
	}

	return c.next.DeleteApp(ctx, req)
}
