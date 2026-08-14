package model

import (
	"fmt"
	"net/url"
	"sort"
	"strings"
)

const Wildcard = "*"

// Context binds a role or permission to concrete resources, e.g.
// workspace=42. Keys are the dimensions, values the resource identifiers
// (or Wildcard).
type Context map[ContextDimension]string

func (c Context) String() string {
	var parts []string
	for k, v := range c {
		parts = append(parts, fmt.Sprintf("%s=%s", k, url.QueryEscape(v)))
	}
	sort.Strings(parts)
	return strings.Join(parts, "&")
}

type ContextDimension string

const (
	ContextWorkspace ContextDimension = "workspace"
	ContextWorkbench ContextDimension = "workbench"
	ContextRequest   ContextDimension = "request"
	ContextUser      ContextDimension = "user"
)

func (r ContextDimension) String() string {
	return string(r)
}

func ToContextDimension(r string) (ContextDimension, error) {
	switch r {
	case string(ContextWorkspace):
		return ContextWorkspace, nil
	case string(ContextWorkbench):
		return ContextWorkbench, nil
	case string(ContextRequest):
		return ContextRequest, nil
	case string(ContextUser):
		return ContextUser, nil
	}

	return "", fmt.Errorf("unknown context dimension: %s", r)
}

// ContextQuantifier expresses how many values of a dimension a role
// definition binds: exactly one ("x") or any ("*").
type ContextQuantifier string

const (
	ContextQuantifierOne ContextQuantifier = "x"
	ContextQuantifierAny ContextQuantifier = "*"
)
