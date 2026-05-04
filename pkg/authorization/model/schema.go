package model

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

const Wildcard = "*"

type ContextQuantifier string

const (
	ContextQuantifierOne ContextQuantifier = "x"
	ContextQuantifierAny ContextQuantifier = "*"
)

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

type RoleDefinition struct {
	Name        RoleName
	Description string
	Scope       RoleScope
	Dynamic     bool

	RequiredContextDimensions map[ContextDimension]ContextQuantifier
	Permissions               []PermissionName
}

type PermissionDefinition struct {
	Name        PermissionName
	Description string

	RequiredContextDimensions []ContextDimension
}

func (r Role) String() string {
	if len(r.Context) == 0 {
		return r.Name.String()
	}

	return fmt.Sprintf("%s@%s", r.Name, r.Context.String())
}

func (p Permission) String() string {
	if len(p.Context) == 0 {
		return p.Name.String()
	}

	return fmt.Sprintf("%s@%s", p.Name, p.Context.String())
}

func (c Context) String() string {
	var parts []string
	for k, v := range c {
		parts = append(parts, fmt.Sprintf("%s=%s", k, url.QueryEscape(v)))
	}
	sort.Strings(parts)
	return strings.Join(parts, "&")
}

func (p Permission) Copy() Permission {
	newContext := make(Context, len(p.Context))
	for k, v := range p.Context {
		newContext[k] = v
	}
	return Permission{
		Name:    p.Name,
		Context: newContext,
	}
}
