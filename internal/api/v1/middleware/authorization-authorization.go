package middleware

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	authorization "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	authorization_service "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"
)

var _ chorus.AuthorizationServiceServer = (*authorizationControllerAuthorization)(nil)

type authorizationControllerAuthorization struct {
	Authorization
	next chorus.AuthorizationServiceServer
}

func AuthorizationAuthorizing(logger *logger.ContextLogger, authorizer authorization_service.Authorizer, cfg config.Config, refresher Refresher) func(chorus.AuthorizationServiceServer) chorus.AuthorizationServiceServer {
	return func(next chorus.AuthorizationServiceServer) chorus.AuthorizationServiceServer {
		return &authorizationControllerAuthorization{
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

func (c authorizationControllerAuthorization) ListRoles(ctx context.Context, req *chorus.ListRolesRequest) (*chorus.ListRolesReply, error) {
	return c.next.ListRoles(ctx, req)
}

func (c authorizationControllerAuthorization) ListPermissions(ctx context.Context, req *chorus.ListPermissionsRequest) (*chorus.ListPermissionsReply, error) {
	return c.next.ListPermissions(ctx, req)
}

func (c authorizationControllerAuthorization) CreateDynamicRole(ctx context.Context, req *chorus.CreateDynamicRoleRequest) (*chorus.CreateDynamicRoleReply, error) {
	if req == nil || req.Role == nil {
		return c.next.CreateDynamicRole(ctx, req)
	}

	scope, err := authorization.ToRoleScope(req.Role.Scope)
	if err != nil {
		return nil, fmt.Errorf("invalid role scope: %w", err)
	}

	switch scope {
	case authorization.RoleScopeWorkspace:
		workspaceID := req.ValidationContext[authorization.RoleContextWorkspace.String()]
		if workspaceID == "" {
			return nil, fmt.Errorf("workspace-scoped dynamic roles require workspace validation context")
		}
		if err := c.IsAuthorized(ctx, authorization.PermissionManageUsersInWorkspace, authorization.WithWorkspace(workspaceID)); err != nil {
			return nil, err
		}
	case authorization.RoleScopePlatform:
		if err := c.IsAuthorized(ctx, authorization.PermissionManageDynamicRoles); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("dynamic roles cannot use scope %q", scope)
	}

	return c.next.CreateDynamicRole(ctx, req)
}
