package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/authorization"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

var _ chorus.WorkspaceServiceServer = (*workspaceControllerAuthorization)(nil)

type workspaceControllerAuthorization struct {
	Authorization
	next chorus.WorkspaceServiceServer
}

func WorkspaceAuthorizing(logger *logger.ContextLogger, authorizer authorization.Authorizer) func(chorus.WorkspaceServiceServer) chorus.WorkspaceServiceServer {
	return func(next chorus.WorkspaceServiceServer) chorus.WorkspaceServiceServer {
		return &workspaceControllerAuthorization{
			Authorization: Authorization{
				logger:     logger,
				authorizer: authorizer,
			},
			next: next,
		}
	}
}

func (c workspaceControllerAuthorization) ListWorkspaces(ctx context.Context, req *chorus.ListWorkspacesRequest) (*chorus.ListWorkspacesReply, error) {
	err := c.IsAuthorized(ctx, authorization.NewPermission(authorization.PermissionListWorkspaces))
	if err != nil {
		return nil, err
	}

	return c.next.ListWorkspaces(ctx, req)
}

func (c workspaceControllerAuthorization) GetWorkspace(ctx context.Context, req *chorus.GetWorkspaceRequest) (*chorus.GetWorkspaceReply, error) {
	err := c.IsAuthorized(ctx, authorization.NewPermission(authorization.PermissionGetWorkspace, authorization.WithWorkspace(req.Id)))
	if err != nil {
		return nil, err
	}

	return c.next.GetWorkspace(ctx, req)
}

func (c workspaceControllerAuthorization) CreateWorkspace(ctx context.Context, req *chorus.Workspace) (*chorus.CreateWorkspaceReply, error) {
	err := c.IsAuthorized(ctx, authorization.NewPermission(authorization.PermissionCreateWorkspace))
	if err != nil {
		return nil, err
	}

	return c.next.CreateWorkspace(ctx, req)
}

func (c workspaceControllerAuthorization) UpdateWorkspace(ctx context.Context, req *chorus.Workspace) (*chorus.UpdateWorkspaceReply, error) {
	err := c.IsAuthorized(ctx, authorization.NewPermission(authorization.PermissionUpdateWorkspace, authorization.WithWorkspace(req.Id)))
	if err != nil {
		return nil, err
	}

	return c.next.UpdateWorkspace(ctx, req)
}

func (c workspaceControllerAuthorization) DeleteWorkspace(ctx context.Context, req *chorus.DeleteWorkspaceRequest) (*chorus.DeleteWorkspaceReply, error) {
	err := c.IsAuthorized(ctx, authorization.NewPermission(authorization.PermissionDeleteWorkspace, authorization.WithWorkspace(req.Id)))
	if err != nil {
		return nil, err
	}

	return c.next.DeleteWorkspace(ctx, req)
}
