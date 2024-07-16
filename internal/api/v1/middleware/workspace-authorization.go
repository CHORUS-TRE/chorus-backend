package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

type workspaceControllerAuthorization struct {
	authorization
	next chorus.WorkspaceServiceServer
}

func WorkspaceAuthorizing(logger *logger.ContextLogger, authorizedRoles []string) func(chorus.WorkspaceServiceServer) chorus.WorkspaceServiceServer {
	return func(next chorus.WorkspaceServiceServer) chorus.WorkspaceServiceServer {
		return &workspaceControllerAuthorization{
			authorization: authorization{
				logger:          logger,
				authorizedRoles: authorizedRoles,
			},
			next: next,
		}
	}
}

func (c workspaceControllerAuthorization) ListWorkspaces(ctx context.Context, req *chorus.ListWorkspacesRequest) (*chorus.ListWorkspacesReply, error) {
	// TODO check for permission

	err := c.isAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.ListWorkspaces(ctx, req)
}

func (c workspaceControllerAuthorization) GetWorkspace(ctx context.Context, req *chorus.GetWorkspaceRequest) (*chorus.GetWorkspaceReply, error) {
	// TODO check for permission

	err := c.isAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.GetWorkspace(ctx, req)
}

func (c workspaceControllerAuthorization) CreateWorkspace(ctx context.Context, req *chorus.Workspace) (*chorus.CreateWorkspaceReply, error) {
	// TODO check for permission

	err := c.isAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	// nolint: staticcheck
	return c.next.CreateWorkspace(ctx, req)
}

func (c workspaceControllerAuthorization) UpdateWorkspace(ctx context.Context, req *chorus.UpdateWorkspaceRequest) (*chorus.UpdateWorkspaceReply, error) {
	// TODO check for permission

	err := c.isAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.UpdateWorkspace(ctx, req)
}

func (c workspaceControllerAuthorization) DeleteWorkspace(ctx context.Context, req *chorus.DeleteWorkspaceRequest) (*chorus.DeleteWorkspaceReply, error) {
	// TODO check for permission

	err := c.isAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.DeleteWorkspace(ctx, req)
}
