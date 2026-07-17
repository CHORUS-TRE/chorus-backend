package model

import (
	"context"
	"fmt"
	"net/url"
	"sort"
	"strings"

	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
)

const Wildcard = "*"

// Context binds a role or permission to concrete resources, e.g.
// workspace=42. Keys are the dimensions, values the resource identifiers
// (or Wildcard).
type Context map[ContextDimension]string

func NewContext(opts ...NewContextOption) Context {
	c := make(Context)
	for _, v := range opts {
		v(&c)
	}
	return c
}

func (c Context) String() string {
	var parts []string
	for k, v := range c {
		parts = append(parts, fmt.Sprintf("%s=%s", k, url.QueryEscape(v)))
	}
	sort.Strings(parts)
	return strings.Join(parts, "&")
}

type NewContextOption func(*Context)

func WithWorkspace(workspace any) NewContextOption {
	return func(c *Context) {
		(*c)[RoleContextWorkspace] = fmt.Sprintf("%v", workspace)
	}
}

func WithWorkbench(workbench any) NewContextOption {
	return func(c *Context) {
		(*c)[RoleContextWorkbench] = fmt.Sprintf("%v", workbench)
	}
}

func WithRequest(request any) NewContextOption {
	return func(c *Context) {
		(*c)[RoleContextRequest] = fmt.Sprintf("%v", request)
	}
}

func WithUser(user any) NewContextOption {
	return func(c *Context) {
		(*c)[RoleContextUser] = fmt.Sprintf("%v", user)
	}
}

func WithUserFromCtx(ctx context.Context) NewContextOption {
	uID := ""
	f := func(c *Context) {
		(*c)[RoleContextUser] = uID
	}

	claims, ok := ctx.Value(jwt_model.JWTClaimsContextKey).(*jwt_model.JWTClaims)
	if !ok {
		return f
	}

	uID = fmt.Sprintf("%v", claims.ID)

	return f
}

type ContextDimension string

const (
	RoleContextWorkspace ContextDimension = "workspace"
	RoleContextWorkbench ContextDimension = "workbench"
	RoleContextRequest   ContextDimension = "request"
	RoleContextUser      ContextDimension = "user"
)

func (r ContextDimension) String() string {
	return string(r)
}

func ToRoleContext(r string) (ContextDimension, error) {
	switch r {
	case string(RoleContextWorkspace):
		return RoleContextWorkspace, nil
	case string(RoleContextWorkbench):
		return RoleContextWorkbench, nil
	case string(RoleContextRequest):
		return RoleContextRequest, nil
	case string(RoleContextUser):
		return RoleContextUser, nil
	}

	return "", fmt.Errorf("unknown role context type: %s", r)
}

// ContextQuantifier expresses how many values of a dimension a role
// definition binds: exactly one ("x") or any ("*").
type ContextQuantifier string

const (
	ContextQuantifierOne ContextQuantifier = "x"
	ContextQuantifierAny ContextQuantifier = "*"
)
