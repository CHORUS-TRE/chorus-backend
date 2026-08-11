package middleware

import (
	"context"
	"fmt"
	"strconv"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	authz "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	authorization_service "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ chorus.WorkspaceServiceServer = (*workspaceControllerAuthorization)(nil)

type workspaceControllerAuthorization struct {
	Authorization
	next chorus.WorkspaceServiceServer
}

func WorkspaceAuthorizing(logger *logger.ContextLogger, authorizer authorization_service.Authorizer, cfg config.Config, refresher Refresher) func(chorus.WorkspaceServiceServer) chorus.WorkspaceServiceServer {
	return func(next chorus.WorkspaceServiceServer) chorus.WorkspaceServiceServer {
		return &workspaceControllerAuthorization{
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

func (c workspaceControllerAuthorization) ListWorkspaces(ctx context.Context, req *chorus.ListWorkspacesRequest) (*chorus.ListWorkspacesReply, error) {
	if req.Filter != nil && len(req.Filter.WorkspaceIdsIn) > 0 {
		for _, id := range req.Filter.WorkspaceIdsIn {
			err := c.IsAuthorized(ctx, authz.PermGetWorkspace.For(authz.WorkspaceID(id)))
			if err != nil {
				logger.TechLog.Error(ctx, fmt.Sprintf("not authorized to access workspace %d: %v", id, err.Error()))
				return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("not authorized to access workspace %d", id))
			}
		}
	} else {
		attrs, err := c.GetContextListForPermission(ctx, authz.PermGetWorkspace.Name)
		if err != nil {
			logger.TechLog.Error(ctx, fmt.Sprintf("unable to get context list for permission: %v", err.Error()))
			return nil, status.Error(codes.Internal, fmt.Sprintf("unable to get context list for permission: %v", err.Error()))
		}

		if len(attrs) == 0 {
			return &chorus.ListWorkspacesReply{Result: &chorus.ListWorkspacesResult{Workspaces: []*chorus.Workspace{}}}, nil
		}

		for _, attr := range attrs {
			if workspaceIDStr, ok := attr[authz.ContextWorkspace]; ok {
				if workspaceIDStr == "" {
					continue
				}
				if workspaceIDStr == "*" {
					req.Filter = nil
					return c.next.ListWorkspaces(ctx, req)
				}
				if req.Filter == nil {
					req.Filter = &chorus.WorkspaceFilter{}
				}
				workspaceID, err := strconv.ParseUint(workspaceIDStr, 10, 64)
				if err != nil {
					logger.TechLog.Error(ctx, fmt.Sprintf("unable to parse workspace ID from context: %v", err.Error()))
					return nil, status.Error(codes.Internal, fmt.Sprintf("unable to parse workspace ID from context: %v", err.Error()))
				}
				req.Filter.WorkspaceIdsIn = append(req.Filter.WorkspaceIdsIn, workspaceID)
			}
		}
	}

	return c.next.ListWorkspaces(ctx, req)
}

func (c workspaceControllerAuthorization) ListPublicWorkspaces(ctx context.Context, req *chorus.ListPublicWorkspacesRequest) (*chorus.ListPublicWorkspacesReply, error) {
	err := c.IsAuthorized(ctx, authz.PermListPublicWorkspaces.For())
	if err != nil {
		return nil, err
	}

	return c.next.ListPublicWorkspaces(ctx, req)
}

func (c workspaceControllerAuthorization) GetWorkspace(ctx context.Context, req *chorus.GetWorkspaceRequest) (*chorus.GetWorkspaceReply, error) {
	err := c.IsAuthorized(ctx, authz.PermGetWorkspace.For(authz.WorkspaceID(req.Id)))
	if err != nil {
		return nil, err
	}

	return c.next.GetWorkspace(ctx, req)
}

func (c workspaceControllerAuthorization) CreateWorkspace(ctx context.Context, req *chorus.Workspace) (*chorus.CreateWorkspaceReply, error) {
	err := c.IsAuthorized(ctx, authz.PermCreateWorkspace.For())
	if err != nil {
		return nil, err
	}

	res, err := c.next.CreateWorkspace(ctx, req)
	if err != nil {
		return nil, err
	}

	err = c.TriggerRefreshToken(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c workspaceControllerAuthorization) UpdateWorkspace(ctx context.Context, req *chorus.Workspace) (*chorus.UpdateWorkspaceReply, error) {
	err := c.IsAuthorized(ctx, authz.PermUpdateWorkspace.For(authz.WorkspaceID(req.Id)))
	if err != nil {
		return nil, err
	}

	return c.next.UpdateWorkspace(ctx, req)
}

func (c workspaceControllerAuthorization) DeleteWorkspace(ctx context.Context, req *chorus.DeleteWorkspaceRequest) (*chorus.DeleteWorkspaceReply, error) {
	err := c.IsAuthorized(ctx, authz.PermDeleteWorkspace.For(authz.WorkspaceID(req.Id)))
	if err != nil {
		return nil, err
	}

	return c.next.DeleteWorkspace(ctx, req)
}

func (c workspaceControllerAuthorization) AddUserRoleInWorkspace(ctx context.Context, req *chorus.AddUserRoleInWorkspaceRequest) (*chorus.AddUserRoleInWorkspaceReply, error) {
	roleName, err := authz.ToRoleName(req.Role.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid role name: %w", err)
	}

	if !c.IsRoleInScope(roleName, authz.RoleScopeWorkspace) {
		return nil, fmt.Errorf("role %q is not a valid workspace role", roleName)
	}

	if roleName == authz.WorkspaceDataManager.Name {
		err = c.IsAuthorized(ctx, authz.PermManageUsersDataRoleInWorkspace.For(authz.WorkspaceID(req.Id)))
		if err != nil {
			return nil, err
		}
	} else {
		err = c.IsAuthorized(ctx, authz.PermManageUsersInWorkspace.For(authz.WorkspaceID(req.Id)))
		if err != nil {
			return nil, err
		}
	}

	assignmentContext := authz.Context{
		authz.ContextWorkspace: fmt.Sprintf("%d", req.Id),
		authz.ContextUser:      fmt.Sprintf("%d", req.UserId),
	}
	if err := c.CanAssignRole(ctx, roleName, assignmentContext); err != nil {
		return nil, err
	}

	return c.next.AddUserRoleInWorkspace(ctx, req)
}

func (c workspaceControllerAuthorization) RemoveUserRoleInWorkspace(ctx context.Context, req *chorus.RemoveUserRoleInWorkspaceRequest) (*chorus.RemoveUserRoleInWorkspaceReply, error) {
	roleName, err := authz.ToRoleName(req.RoleName)
	if err != nil {
		return nil, fmt.Errorf("invalid role name: %w", err)
	}

	if !c.IsRoleInScope(roleName, authz.RoleScopeWorkspace) {
		return nil, fmt.Errorf("role %q is not a valid workspace role", roleName)
	}

	if roleName == authz.WorkspaceDataManager.Name {
		err = c.IsAuthorized(ctx, authz.PermManageUsersDataRoleInWorkspace.For(authz.WorkspaceID(req.Id)))
		if err != nil {
			return nil, err
		}
	} else {
		err = c.IsAuthorized(ctx, authz.PermManageUsersInWorkspace.For(authz.WorkspaceID(req.Id)))
		if err != nil {
			return nil, err
		}
	}

	return c.next.RemoveUserRoleInWorkspace(ctx, req)
}

func (c workspaceControllerAuthorization) RemoveUserFromWorkspace(ctx context.Context, req *chorus.RemoveUserFromWorkspaceRequest) (*chorus.RemoveUserFromWorkspaceReply, error) {
	err := c.IsAuthorized(ctx, authz.PermManageUsersInWorkspace.For(authz.WorkspaceID(req.Id)))
	if err != nil {
		return nil, err
	}

	return c.next.RemoveUserFromWorkspace(ctx, req)
}
