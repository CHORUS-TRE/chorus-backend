package model

import (
	"fmt"
	"slices"
	"strings"
)

// Typed resource ids. Each knows its context dimension, so a permission or role
// derives its dimension from the id type — no separate ContextDimension to pass
// or keep in sync.
type (
	WorkspaceID uint64
	WorkbenchID uint64
	UserID      uint64
)

func (WorkspaceID) dimension() ContextDimension { return ContextWorkspace }
func (WorkbenchID) dimension() ContextDimension { return ContextWorkbench }
func (UserID) dimension() ContextDimension      { return ContextUser }

// contextID is a typed resource id that knows its dimension.
type contextID interface {
	~uint64
	dimension() ContextDimension
}

// PermissionDefinition declares a permission and the context dimensions a check
// against it requires. See Permission for the contextual check counterpart.
type PermissionDefinition struct {
	Name                      PermissionName
	Description               string
	RequiredContextDimensions []ContextDimension
}

// RoleDefinition declares what a role grants: the permissions it confers and the
// context dimensions an assignment must bind. See Role for the assignment
// counterpart carried by users.
type RoleDefinition struct {
	Name                      RoleName
	Description               string
	Scope                     RoleScope
	Dynamic                   bool
	RequiredContextDimensions map[ContextDimension]ContextQuantifier
	Permissions               []PermissionName
}

// FindUsersWithPermissionFilter narrows a user search to those granted a
// permission, optionally in a specific context or via specific roles.
type FindUsersWithPermissionFilter struct {
	PermissionName          PermissionName
	Context                 Context
	ViaRoles                []RoleName
	PreferExactContextMatch bool
}

// Permission is a permission check target: a PermissionName plus the concrete
// context it applies to. See PermissionDefinition for the schema counterpart.
type Permission struct {
	Name    PermissionName
	Context Context
}

type PermissionName string

func (p PermissionName) String() string {
	return string(p)
}

// ToPermissionName resolves a raw string to its declared PermissionName; a
// permission absent from the registry is not resolvable.
func ToPermissionName(p string) (PermissionName, error) {
	for _, def := range GetPermissionDefinitions() {
		if string(def.Name) == p {
			return def.Name, nil
		}
	}
	return "", fmt.Errorf("unknown permission type: %s", p)
}

func (p Permission) String() string {
	if len(p.Context) == 0 {
		return p.Name.String()
	}

	return fmt.Sprintf("%s@%s", p.Name, p.Context.String())
}

// UniquePermissionNames returns a sorted, deduplicated list of permission names from a slice of permissions.
func UniquePermissionNames(permissions []Permission) []string {
	seen := make(map[string]struct{}, len(permissions))
	for _, p := range permissions {
		seen[string(p.Name)] = struct{}{}
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

// Role is a role assignment: a RoleName plus the concrete context it applies
// to. It is what users carry (e.g. in the JWT). See RoleDefinition for the
// schema counterpart declaring what a role grants.
type Role struct {
	Name    RoleName `json:"name"`
	Context Context  `json:"context"`
}

func ToRole(name string, context map[string]string) (Role, error) {
	roleName, err := ToRoleName(name)
	if err != nil {
		return Role{}, err
	}

	ctx := make(Context)
	for k, v := range context {
		ctx[ContextDimension(k)] = v
	}

	return Role{
		Name:    roleName,
		Context: ctx,
	}, nil
}

func (r Role) String() string {
	if len(r.Context) == 0 {
		return r.Name.String()
	}

	return fmt.Sprintf("%s@%s", r.Name, r.Context.String())
}

type RoleName string

func (r RoleName) String() string {
	return string(r)
}

// ToRoleName resolves a raw string to a RoleName; unknown names are accepted
// as dynamic roles. Callers that need the role to exist check the schema.
func ToRoleName(r string) (RoleName, error) {
	if strings.TrimSpace(r) == "" {
		return "", fmt.Errorf("empty role type")
	}

	return RoleName(r), nil
}

// IsSystemRole reports whether the name is one of the code-defined roles.
func IsSystemRole(role RoleName) bool {
	return slices.ContainsFunc(GetRoleDefinitions(), func(def *RoleDefinition) bool {
		return def.Name == role
	})
}

// AuthorizationSchema is the full authorization model in effect: every
// permission (always code-defined) and every role (the code defaults plus any
// dynamic roles loaded from the store).
type AuthorizationSchema struct {
	Roles       []*RoleDefinition
	Permissions []PermissionDefinition
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
