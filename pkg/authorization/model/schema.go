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

type RoleDefinition struct {
	Name        RoleName
	Description string

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
