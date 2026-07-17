package model

import (
	"fmt"
)

// AuthorizationSchema is the full declared authorization model: every
// permission and every role. The default schema is declared in
// default_schema.go; dynamic roles are added at runtime from the store.
type AuthorizationSchema struct {
	Roles       []*RoleDefinition
	Permissions []PermissionDefinition
}

// RoleDefinition declares what a role grants: its permissions and the
// context dimensions an assignment must bind. See Role for the assignment
// counterpart carried by users.
type RoleDefinition struct {
	Name        RoleName
	Description string
	Scope       RoleScope
	Dynamic     bool

	RequiredContextDimensions map[ContextDimension]ContextQuantifier
	Permissions               []PermissionName
}

// PermissionDefinition declares a permission and the context dimensions a
// check against it requires. See Permission for the check counterpart.
type PermissionDefinition struct {
	Name        PermissionName
	Description string

	RequiredContextDimensions []ContextDimension
}

type RoleScope string

const (
	RoleScopeSystem    RoleScope = "system"
	RoleScopePlatform  RoleScope = "platform"
	RoleScopeWorkspace RoleScope = "workspace"
	RoleScopeWorkbench RoleScope = "workbench"
)

func (s RoleScope) String() string {
	return string(s)
}

func ToRoleScope(scope string) (RoleScope, error) {
	switch scope {
	case string(RoleScopeSystem):
		return RoleScopeSystem, nil
	case string(RoleScopePlatform):
		return RoleScopePlatform, nil
	case string(RoleScopeWorkspace):
		return RoleScopeWorkspace, nil
	case string(RoleScopeWorkbench):
		return RoleScopeWorkbench, nil
	}
	return "", fmt.Errorf("unknown role scope: %s", scope)
}
