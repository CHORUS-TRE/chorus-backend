package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/authorization"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

var _ chorus.WorkspaceFileServiceServer = (*workspaceFileControllerAuthorization)(nil)

type workspaceFileControllerAuthorization struct {
	Authorization
	next chorus.WorkspaceFileServiceServer
}

func WorkspaceFileAuthorizing(logger *logger.ContextLogger, authorizer authorization.Authorizer, cfg config.Config, refresher Refresher) func(chorus.WorkspaceFileServiceServer) chorus.WorkspaceFileServiceServer {
	return func(next chorus.WorkspaceFileServiceServer) chorus.WorkspaceFileServiceServer {
		return &workspaceFileControllerAuthorization{
			Authorization: Authorization{
				logger:     logger,
				authorizer: authorizer,
				cfg:        cfg,
				refresher:  refresher,
			},
			next: next,
		}
	}
}

func (c workspaceFileControllerAuthorization) GetWorkspaceFile(ctx context.Context, req *chorus.GetWorkspaceFileRequest) (*chorus.GetWorkspaceFileReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionDownloadFilesFromWorkspace, authorization.WithWorkspace(req.WorkspaceId))
	if err != nil {
		return nil, err
	}

	return c.next.GetWorkspaceFile(ctx, req)
}

func (c workspaceFileControllerAuthorization) ListWorkspaceFiles(ctx context.Context, req *chorus.ListWorkspaceFilesRequest) (*chorus.ListWorkspaceFilesReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionListFilesInWorkspace, authorization.WithWorkspace(req.WorkspaceId))
	if err != nil {
		return nil, err
	}

	return c.next.ListWorkspaceFiles(ctx, req)
}

func (c workspaceFileControllerAuthorization) CreateWorkspaceFile(ctx context.Context, req *chorus.CreateWorkspaceFileRequest) (*chorus.CreateWorkspaceFileReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionUploadFilesToWorkspace, authorization.WithWorkspace(req.WorkspaceId))
	if err != nil {
		return nil, err
	}

	return c.next.CreateWorkspaceFile(ctx, req)
}

func (c workspaceFileControllerAuthorization) UpdateWorkspaceFile(ctx context.Context, req *chorus.UpdateWorkspaceFileRequest) (*chorus.UpdateWorkspaceFileReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionModifyFilesInWorkspace, authorization.WithWorkspace(req.WorkspaceId))
	if err != nil {
		return nil, err
	}

	return c.next.UpdateWorkspaceFile(ctx, req)
}

func (c workspaceFileControllerAuthorization) DeleteWorkspaceFile(ctx context.Context, req *chorus.DeleteWorkspaceFileRequest) (*chorus.DeleteWorkspaceFileReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionModifyFilesInWorkspace, authorization.WithWorkspace(req.WorkspaceId))
	if err != nil {
		return nil, err
	}

	return c.next.DeleteWorkspaceFile(ctx, req)
}
