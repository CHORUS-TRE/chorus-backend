package middleware

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/authorization"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/model"
)

var _ chorus.AppInstanceServiceServer = (*appInstanceControllerAuthorization)(nil)

type AppInstanceResolver interface {
	GetAppInstance(ctx context.Context, tenantID uint64, appInstanceID uint64) (*model.AppInstance, error)
}

type appInstanceControllerAuthorization struct {
	Authorization
	resolver AppInstanceResolver
	next     chorus.AppInstanceServiceServer
}

func AppInstanceAuthorizing(logger *logger.ContextLogger, authorizer authorization.Authorizer, resolver AppInstanceResolver) func(chorus.AppInstanceServiceServer) chorus.AppInstanceServiceServer {
	return func(next chorus.AppInstanceServiceServer) chorus.AppInstanceServiceServer {
		return &appInstanceControllerAuthorization{
			Authorization: Authorization{
				logger:     logger,
				authorizer: authorizer,
			},
			resolver: resolver,
			next:     next,
		}
	}
}

func (c appInstanceControllerAuthorization) ListAppInstances(ctx context.Context, req *chorus.ListAppInstancesRequest) (*chorus.ListAppInstancesReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionListAppInstances) // TODO: add workbench/workspace scoping for app instances listing
	if err != nil {
		return nil, err
	}

	return c.next.ListAppInstances(ctx, req)
}

func (c appInstanceControllerAuthorization) GetAppInstance(ctx context.Context, req *chorus.GetAppInstanceRequest) (*chorus.GetAppInstanceReply, error) {
	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	appInstance, err := c.resolver.GetAppInstance(ctx, tenantID, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "unable to resolve app instance %v: %v", req.Id, err)
	}

	err = c.IsAuthorized(ctx, authorization.PermissionGetAppInstance,
		authorization.WithWorkbench(appInstance.WorkbenchID),
		authorization.WithWorkspace(appInstance.WorkspaceID),
	)
	if err != nil {
		return nil, err
	}

	return c.next.GetAppInstance(ctx, req)
}

func (c appInstanceControllerAuthorization) CreateAppInstance(ctx context.Context, req *chorus.AppInstance) (*chorus.CreateAppInstanceReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionCreateAppInstance, authorization.WithWorkbench(req.WorkbenchId), authorization.WithWorkspace(req.WorkspaceId))
	if err != nil {
		return nil, err
	}

	return c.next.CreateAppInstance(ctx, req)
}

func (c appInstanceControllerAuthorization) UpdateAppInstance(ctx context.Context, req *chorus.AppInstance) (*chorus.UpdateAppInstanceReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionUpdateAppInstance, authorization.WithWorkbench(req.WorkbenchId), authorization.WithWorkspace(req.WorkspaceId))
	if err != nil {
		return nil, err
	}

	return c.next.UpdateAppInstance(ctx, req)
}

func (c appInstanceControllerAuthorization) DeleteAppInstance(ctx context.Context, req *chorus.DeleteAppInstanceRequest) (*chorus.DeleteAppInstanceReply, error) {
	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	appInstance, err := c.resolver.GetAppInstance(ctx, tenantID, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "unable to resolve app instance %v: %v", req.Id, err)
	}

	err = c.IsAuthorized(ctx, authorization.PermissionDeleteAppInstance,
		authorization.WithWorkbench(appInstance.WorkbenchID),
		authorization.WithWorkspace(appInstance.WorkspaceID),
	)
	if err != nil {
		return nil, err
	}

	return c.next.DeleteAppInstance(ctx, req)
}
