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
)

var _ chorus.WorkbenchServiceServer = (*workbenchControllerAuthorization)(nil)

type workbenchControllerAuthorization struct {
	Authorization
	next chorus.WorkbenchServiceServer
}

func WorkbenchAuthorizing(logger *logger.ContextLogger, authorizer authorization_service.Authorizer, cfg config.Config, refresher Refresher) func(chorus.WorkbenchServiceServer) chorus.WorkbenchServiceServer {
	return func(next chorus.WorkbenchServiceServer) chorus.WorkbenchServiceServer {
		return &workbenchControllerAuthorization{
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

func (c workbenchControllerAuthorization) ListWorkbenches(ctx context.Context, req *chorus.ListWorkbenchesRequest) (*chorus.ListWorkbenchesReply, error) {
	if req.Filter != nil && len(req.Filter.WorkspaceIdsIn) > 0 {
		for _, id := range req.Filter.WorkspaceIdsIn {
			err := c.IsAuthorized(ctx, authorization.PermissionGetWorkspace, authorization.WithWorkspace(id))
			if err != nil {
				return nil, err
			}
		}
	} else {
		attrs, err := c.GetContextListForPermission(ctx, authorization.PermissionGetWorkspace)
		if err != nil {
			return nil, err
		}

		if len(attrs) == 0 {
			return &chorus.ListWorkbenchesReply{Result: &chorus.ListWorkbenchesResult{Workbenches: []*chorus.Workbench{}}}, nil
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
					fmt.Println("wildcard found, returning all workbenches")
					req.Filter = nil
					return c.next.ListWorkbenches(ctx, req)
				}
				if req.Filter == nil {
					req.Filter = &chorus.WorkbenchFilter{}
				}
				workspaceID, err := strconv.ParseUint(workspaceIDStr, 10, 64)
				if err != nil {
					return nil, err
				}
				fmt.Println("adding workspace ID to filter:", workspaceID)
				req.Filter.WorkspaceIdsIn = append(req.Filter.WorkspaceIdsIn, workspaceID)
			}
		}
	}

	return c.next.ListWorkbenches(ctx, req)
}

func (c workbenchControllerAuthorization) GetWorkbench(ctx context.Context, req *chorus.GetWorkbenchRequest) (*chorus.GetWorkbenchReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionGetWorkbench, authorization.WithWorkbench(req.Id))
	if err != nil {
		return nil, err
	}

	return c.next.GetWorkbench(ctx, req)
}

func (c workbenchControllerAuthorization) CreateWorkbench(ctx context.Context, req *chorus.Workbench) (*chorus.CreateWorkbenchReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionCreateWorkbench, authorization.WithWorkspace(req.WorkspaceId))
	if err != nil {
		return nil, err
	}

	res, err := c.next.CreateWorkbench(ctx, req)
	if err != nil {
		return nil, err
	}

	err = c.TriggerRefreshToken(ctx)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (c workbenchControllerAuthorization) UpdateWorkbench(ctx context.Context, req *chorus.Workbench) (*chorus.UpdateWorkbenchReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionUpdateWorkbench, authorization.WithWorkbench(req.Id))
	if err != nil {
		return nil, err
	}

	return c.next.UpdateWorkbench(ctx, req)
}

func (c workbenchControllerAuthorization) DeleteWorkbench(ctx context.Context, req *chorus.DeleteWorkbenchRequest) (*chorus.DeleteWorkbenchReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionDeleteWorkbench, authorization.WithWorkbench(req.Id))
	if err != nil {
		return nil, err
	}

	return c.next.DeleteWorkbench(ctx, req)
}

func (c workbenchControllerAuthorization) ManageUserRoleInWorkbench(ctx context.Context, req *chorus.ManageUserRoleInWorkbenchRequest) (*chorus.ManageUserRoleInWorkbenchReply, error) {
	roleName, err := authorization.ToRoleName(req.Role.Name)
	if err != nil {
		return nil, fmt.Errorf("invalid role name: %w", err)
	}

	if !authorization.RoleIn(roleName, authorization.GetWorkbenchRoles()) {
		return nil, fmt.Errorf("role %q is not a workbench role", roleName)
	}

	err = c.IsAuthorized(ctx, authorization.PermissionManageUsersInWorkbench, authorization.WithWorkbench(req.Id))
	if err != nil {
		return nil, err
	}

	return c.next.ManageUserRoleInWorkbench(ctx, req)
}

func (c workbenchControllerAuthorization) RemoveUserFromWorkbench(ctx context.Context, req *chorus.RemoveUserFromWorkbenchRequest) (*chorus.RemoveUserFromWorkbenchReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionManageUsersInWorkbench, authorization.WithWorkbench(req.Id))
	if err != nil {
		return nil, err
	}

	return c.next.RemoveUserFromWorkbench(ctx, req)
}
