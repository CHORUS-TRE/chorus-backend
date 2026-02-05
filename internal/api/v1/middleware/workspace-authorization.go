package middleware

import (
	"context"
	"fmt"
	"strconv"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	authorization "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	authorization_service "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"
	"go.uber.org/zap"
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
			err := c.IsAuthorized(ctx, authorization.PermissionGetWorkspace, authorization.WithWorkspace(id))
			if err != nil {
				logger.TechLog.Error(ctx, fmt.Sprintf("not authorized to access workspace %d: %v", id, err.Error()))
				return nil, status.Error(codes.PermissionDenied, fmt.Sprintf("not authorized to access workspace %d", id))
			}
		}
	} else {
		attrs, err := c.GetContextListForPermission(ctx, authorization.PermissionGetWorkspace)
		if err != nil {
			logger.TechLog.Error(ctx, fmt.Sprintf("unable to get context list for permission: %v", err.Error()))
			return nil, status.Error(codes.Internal, fmt.Sprintf("unable to get context list for permission: %v", err.Error()))
		}

		if len(attrs) == 0 {
			return &chorus.ListWorkspacesReply{Result: &chorus.ListWorkspacesResult{Workspaces: []*chorus.Workspace{}}}, nil
		}

		fmt.Println("attrs:", attrs)
		claims, ok := ctx.Value(jwt_model.JWTClaimsContextKey).(*jwt_model.JWTClaims)
		if ok {
			aRoles, err := claimRolesToAuthRoles(claims)
			var permission []authorization.Permission
			if err == nil {
				permission, _ = c.authorizer.GetUserPermissions(aRoles)
				fmt.Println("permissions:", permission)
			}
		}

		for _, attr := range attrs {
			if workspaceIDStr, ok := attr[authorization.RoleContextWorkspace]; ok {
				if workspaceIDStr == "" {
					continue
				}
				if workspaceIDStr == "*" {
					fmt.Println("wildcard found, returning all workspaces")
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
				fmt.Println("adding workspace ID to filter:", workspaceID)
				req.Filter.WorkspaceIdsIn = append(req.Filter.WorkspaceIdsIn, workspaceID)
			}
		}
	}

	return c.next.ListWorkspaces(ctx, req)
}

func (c workspaceControllerAuthorization) GetWorkspace(ctx context.Context, req *chorus.GetWorkspaceRequest) (*chorus.GetWorkspaceReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionGetWorkspace, authorization.WithWorkspace(req.Id))
	if err != nil {
		return nil, err
	}

	return c.next.GetWorkspace(ctx, req)
}

func (c workspaceControllerAuthorization) CreateWorkspace(ctx context.Context, req *chorus.Workspace) (*chorus.CreateWorkspaceReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionCreateWorkspace)
	if err != nil {
		explanation := c.ExplainIsAuthorized(ctx, authorization.PermissionCreateWorkspace)
		c.logger.Info(ctx, "explanation for failed authorization", zap.String("explanation", explanation))
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
	err := c.IsAuthorized(ctx, authorization.PermissionUpdateWorkspace, authorization.WithWorkspace(req.Id))
	if err != nil {
		return nil, err
	}

	return c.next.UpdateWorkspace(ctx, req)
}

func (c workspaceControllerAuthorization) DeleteWorkspace(ctx context.Context, req *chorus.DeleteWorkspaceRequest) (*chorus.DeleteWorkspaceReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionDeleteWorkspace, authorization.WithWorkspace(req.Id))
	if err != nil {
		return nil, err
	}

	return c.next.DeleteWorkspace(ctx, req)
}

func (c workspaceControllerAuthorization) ManageUserRoleInWorkspace(ctx context.Context, req *chorus.ManageUserRoleInWorkspaceRequest) (*chorus.ManageUserRoleInWorkspaceReply, error) {
	roleName, err := authorization.ToRoleName(req.Role.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid role name: %w", err)
	}

	if !authorization.RoleIn(roleName, authorization.GetWorkspaceRoles()) {
		return nil, fmt.Errorf("user is not authorized to manage role %q in workspace", roleName)
	}

	err = c.IsAuthorized(ctx, authorization.PermissionManageUsersInWorkspace, authorization.WithWorkspace(req.Id))
	if err != nil {
		return nil, err
	}

	return c.next.ManageUserRoleInWorkspace(ctx, req)
}

func (c workspaceControllerAuthorization) RemoveUserFromWorkspace(ctx context.Context, req *chorus.RemoveUserFromWorkspaceRequest) (*chorus.RemoveUserFromWorkspaceReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionManageUsersInWorkspace, authorization.WithWorkspace(req.Id))
	if err != nil {
		return nil, err
	}

	return c.next.RemoveUserFromWorkspace(ctx, req)
}
