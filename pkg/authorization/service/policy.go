package service

// policy.go is the minimal authorization kernel: expanding a user's roles
// into contextual permissions and matching a permission check against them.
// It is pure — plain functions over plain maps, no service state, no I/O —
// so it can be exhaustively unit-tested.

import (
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
)

// expandUserPermissions expands the user's roles into permissions, each
// carrying the context values the granting role binds for the dimensions the
// permission requires.
func expandUserPermissions(
	roles map[model.RoleName]*model.RoleDefinition,
	permissions map[model.PermissionName]model.PermissionDefinition,
	userRoles []model.Role,
) ([]model.Permission, error) {
	expanded := make([]model.Permission, 0)

	for _, role := range userRoles {
		roleDefinition, ok := roles[role.Name]
		if !ok {
			return nil, fmt.Errorf("role %q not found in schema", role.Name)
		}

		for _, permissionName := range roleDefinition.Permissions {
			required := permissions[permissionName].RequiredContextDimensions
			context := make(model.Context, len(required))
			for _, dimension := range required {
				if value, ok := role.Context[dimension]; ok {
					context[dimension] = value
				}
			}
			if len(context) == len(required) {
				expanded = append(expanded, model.Permission{Name: permissionName, Context: context})
			}
		}
	}
	return expanded, nil
}

// isUserAllowed reports whether one of the user's roles grants the checked
// permission in its context.
func isUserAllowed(
	roles map[model.RoleName]*model.RoleDefinition,
	permissions map[model.PermissionName]model.PermissionDefinition,
	user []model.Role,
	permission model.Permission,
) (bool, error) {
	if _, ok := permissions[permission.Name]; !ok {
		return false, fmt.Errorf("unknown permission: %s", permission)
	}
	userPermissions, err := expandUserPermissions(roles, permissions, user)
	if err != nil {
		return false, err
	}
	for _, userPermission := range userPermissions {
		if isPermissionIdentical(userPermission, permission) {
			return true, nil
		}
	}
	return false, nil
}

// isPermissionIdentical reports whether userPermission grants permission:
// names must match and every context dimension the user permission binds
// must match the checked permission's value, or be the wildcard.
func isPermissionIdentical(userPermission, permission model.Permission) bool {
	if userPermission.Name != permission.Name {
		return false
	}
	for dimension, value := range userPermission.Context {
		if value != model.Wildcard && value != permission.Context[dimension] {
			return false
		}
	}
	return true
}

// explainIsPermissionIdentical mirrors isPermissionIdentical and renders the
// comparison; it must stay in lockstep with the matching rules above.
func explainIsPermissionIdentical(userPermission, permission model.Permission) (bool, string) {
	format := func(res bool) (bool, string) {
		comparison := "=="
		if !res {
			comparison = "!="
		}
		return res, fmt.Sprintf("%s %s %s", userPermission.String(), comparison, permission.String())
	}

	if userPermission.Name != permission.Name {
		return format(false)
	}
	for dimension, value := range userPermission.Context {
		if value != model.Wildcard && value != permission.Context[dimension] {
			return format(false)
		}
	}
	return format(true)
}
