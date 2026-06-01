package converter

import (
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	authorization_model "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
)

func RoleFromBusiness(role *authorization_model.RoleDefinition) *chorus.AuthorizationRole {
	contextDimensions := make([]string, 0, len(role.RequiredContextDimensions))
	for dim := range role.RequiredContextDimensions {
		contextDimensions = append(contextDimensions, string(dim))
	}
	return &chorus.AuthorizationRole{
		Name:        role.Name.String(),
		Description: role.Description,
		Context:     contextDimensions,
		Permissions: PermissionNames(role.Permissions),
		Scope:       role.Scope.String(),
		Dynamic:     role.Dynamic,
	}
}

func DynamicRoleToBusiness(role *chorus.AuthorizationRole) (*authorization_model.RoleDefinition, error) {
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

func ContextToBusiness(raw map[string]string) (authorization_model.Context, error) {
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

func PermissionNames(permissions []authorization_model.PermissionName) []string {
	result := make([]string, 0, len(permissions))
	for _, permission := range permissions {
		result = append(result, permission.String())
	}
	return result
}
