package v1

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	cerr "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	authorization_model "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	authorization_service "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"
)

var _ chorus.AuthorizationServiceServer = (*AuthorizationController)(nil)

type AuthorizationController struct {
	authorization authorization_service.AuthorizationServiceInterface
}

func NewAuthorizationController(authorization authorization_service.AuthorizationServiceInterface) chorus.AuthorizationServiceServer {
	return &AuthorizationController{
		authorization: authorization,
	}
}

func (c AuthorizationController) ListRoles(ctx context.Context, req *chorus.ListRolesRequest) (*chorus.ListRolesReply, error) {
	schema := c.authorization.GetAuthorizationSchema()
	if schema == nil {
		return &chorus.ListRolesReply{
			Result: &chorus.ListRolesResult{
				Roles: []*chorus.AuthorizationRole{},
			},
		}, nil
	}

	roles := make([]*chorus.AuthorizationRole, 0, len(schema.Roles))
	for _, role := range schema.Roles {
		contextDimensions := make([]string, 0, len(role.RequiredContextDimensions))
		for dim := range role.RequiredContextDimensions {
			contextDimensions = append(contextDimensions, string(dim))
		}

		roles = append(roles, &chorus.AuthorizationRole{
			Name:        role.Name.String(),
			Description: role.Description,
			Context:     contextDimensions,
			Permissions: permissionNames(role.Permissions),
			Scope:       role.Scope.String(),
			Dynamic:     role.Dynamic,
		})
	}

	return &chorus.ListRolesReply{
		Result: &chorus.ListRolesResult{
			Roles: roles,
		},
	}, nil
}

func (c AuthorizationController) CreateDynamicRole(ctx context.Context, req *chorus.CreateDynamicRoleRequest) (*chorus.CreateDynamicRoleReply, error) {
	if req == nil || req.Role == nil {
		return nil, cerr.ErrInvalidRequest.WithMessage("Empty request")
	}

	roles, err := authorizationRolesFromContext(ctx)
	if err != nil {
		return nil, err
	}

	role, err := dynamicRoleToBusiness(req.Role)
	if err != nil {
		return nil, cerr.ErrInvalidRequest.Wrap(err, "Invalid dynamic role")
	}

	validationContext, err := contextToBusiness(req.ValidationContext)
	if err != nil {
		return nil, cerr.ErrInvalidRequest.Wrap(err, "Invalid validation context")
	}

	createdRole, err := c.authorization.CreateDynamicRole(ctx, roles, role, validationContext)
	if err != nil {
		return nil, err
	}

	return &chorus.CreateDynamicRoleReply{Result: &chorus.CreateDynamicRoleResult{Role: roleFromBusiness(createdRole)}}, nil
}

func authorizationRolesFromContext(ctx context.Context) ([]authorization_model.Role, error) {
	claims, ok := ctx.Value(jwt_model.JWTClaimsContextKey).(*jwt_model.JWTClaims)
	if !ok {
		return nil, cerr.ErrInvalidRequest.WithMessage("Malformed JWT token")
	}

	roles := make([]authorization_model.Role, 0, len(claims.Roles))
	for _, claimRole := range claims.Roles {
		role, err := authorization_model.ToRole(claimRole.Name, claimRole.Context)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

func roleFromBusiness(role *authorization_model.RoleDefinition) *chorus.AuthorizationRole {
	contextDimensions := make([]string, 0, len(role.RequiredContextDimensions))
	for dim := range role.RequiredContextDimensions {
		contextDimensions = append(contextDimensions, string(dim))
	}
	return &chorus.AuthorizationRole{
		Name:        role.Name.String(),
		Description: role.Description,
		Context:     contextDimensions,
		Permissions: permissionNames(role.Permissions),
		Scope:       role.Scope.String(),
		Dynamic:     role.Dynamic,
	}
}

func dynamicRoleToBusiness(role *chorus.AuthorizationRole) (*authorization_model.RoleDefinition, error) {
	name, err := authorization_model.ToRoleName(role.Name)
	if err != nil {
		return nil, err
	}
	scope, err := authorization_model.ToRoleScope(role.Scope)
	if err != nil {
		return nil, err
	}
	permissions := make([]authorization_model.PermissionName, 0, len(role.Permissions))
	for _, permission := range role.Permissions {
		permissionName, err := authorization_model.ToPermissionName(permission)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permissionName)
	}
	return &authorization_model.RoleDefinition{
		Name:        name,
		Description: role.Description,
		Scope:       scope,
		Dynamic:     true,
		Permissions: permissions,
	}, nil
}

func contextToBusiness(raw map[string]string) (authorization_model.Context, error) {
	ctx := make(authorization_model.Context, len(raw))
	for key, value := range raw {
		dimension, err := authorization_model.ToRoleContext(key)
		if err != nil {
			return nil, fmt.Errorf("invalid context dimension %q: %w", key, err)
		}
		ctx[dimension] = value
	}
	return ctx, nil
}

func permissionNames(permissions []authorization_model.PermissionName) []string {
	result := make([]string, 0, len(permissions))
	for _, permission := range permissions {
		result = append(result, permission.String())
	}
	return result
}

func (c AuthorizationController) ListPermissions(ctx context.Context, req *chorus.ListPermissionsRequest) (*chorus.ListPermissionsReply, error) {
	schema := c.authorization.GetAuthorizationSchema()
	if schema == nil {
		return &chorus.ListPermissionsReply{
			Result: &chorus.ListPermissionsResult{
				Permissions: []*chorus.AuthorizationPermission{},
			},
		}, nil
	}

	permissions := make([]*chorus.AuthorizationPermission, 0, len(schema.Permissions))
	for _, perm := range schema.Permissions {
		contextDimensions := make([]string, 0, len(perm.RequiredContextDimensions))
		for _, dim := range perm.RequiredContextDimensions {
			contextDimensions = append(contextDimensions, string(dim))
		}

		permissions = append(permissions, &chorus.AuthorizationPermission{
			Name:        perm.Name.String(),
			Description: perm.Description,
			Context:     contextDimensions,
		})
	}

	return &chorus.ListPermissionsReply{
		Result: &chorus.ListPermissionsResult{
			Permissions: permissions,
		},
	}, nil
}
