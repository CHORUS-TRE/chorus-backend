package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

var _ chorus.WorkspaceServiceServer = (*workspaceControllerAuthorization)(nil)

type workspaceControllerAuthorization struct {
	Authorization
	next chorus.WorkspaceServiceServer
}

func WorkspaceAuthorizing(logger *logger.ContextLogger, authorizedRoles []string) func(chorus.WorkspaceServiceServer) chorus.WorkspaceServiceServer {
	return func(next chorus.WorkspaceServiceServer) chorus.WorkspaceServiceServer {
		return &workspaceControllerAuthorization{
			Authorization: Authorization{
				logger:          logger,
				authorizedRoles: authorizedRoles,
			},
			next: next,
		}
	}
}

func (c workspaceControllerAuthorization) ListWorkspaces(ctx context.Context, req *chorus.ListWorkspacesRequest) (*chorus.ListWorkspacesReply, error) {
	// TODO check for permission

	err := c.IsAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.ListWorkspaces(ctx, req)
}

func (c workspaceControllerAuthorization) GetWorkspace(ctx context.Context, req *chorus.GetWorkspaceRequest) (*chorus.GetWorkspaceReply, error) {
	// TODO check for permission

	err := c.IsAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.GetWorkspace(ctx, req)
}

func (c workspaceControllerAuthorization) CreateWorkspace(ctx context.Context, req *chorus.Workspace) (*chorus.CreateWorkspaceReply, error) {
	// TODO check for permission

	err := c.IsAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	// nolint: staticcheck
	return c.next.CreateWorkspace(ctx, req)
}

func (c workspaceControllerAuthorization) UpdateWorkspace(ctx context.Context, req *chorus.Workspace) (*chorus.UpdateWorkspaceReply, error) {
	// TODO check for permission

	err := c.IsAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.UpdateWorkspace(ctx, req)
}

func (c workspaceControllerAuthorization) DeleteWorkspace(ctx context.Context, req *chorus.DeleteWorkspaceRequest) (*chorus.DeleteWorkspaceReply, error) {
	// TODO check for permission

	err := c.IsAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.DeleteWorkspace(ctx, req)
}

func (c workspaceControllerAuthorization) GetWorkspaceFile(ctx context.Context, req *chorus.GetWorkspaceFileRequest) (*chorus.GetWorkspaceFileReply, error) {
	// TODO check for permission

	err := c.IsAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.GetWorkspaceFile(ctx, req)
}

func (c workspaceControllerAuthorization) CreateWorkspaceFile(ctx context.Context, req *chorus.CreateWorkspaceFileRequest) (*chorus.CreateWorkspaceFileReply, error) {
	// TODO check for permission

	err := c.IsAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.CreateWorkspaceFile(ctx, req)
}

func (c workspaceControllerAuthorization) UpdateWorkspaceFile(ctx context.Context, req *chorus.UpdateWorkspaceFileRequest) (*chorus.UpdateWorkspaceFileReply, error) {
	// TODO check for permission

	err := c.IsAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.UpdateWorkspaceFile(ctx, req)
}

func (c workspaceControllerAuthorization) DeleteWorkspaceFile(ctx context.Context, req *chorus.DeleteWorkspaceFileRequest) (*chorus.DeleteWorkspaceFileReply, error) {
	// TODO check for permission

	err := c.IsAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}
	//nolint: staticcheck
	return c.next.DeleteWorkspaceFile(ctx, req)
}
