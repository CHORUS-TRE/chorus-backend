//go:build unit

package service

import (
	"testing"

	"github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
)

// TestExpandDropsUnboundRequiredContext is the regression test for the
// cross-workspace escalation: a role that binds only `workspace` but grants a
// permission requiring `workbench` must never match a workbench check.
func TestExpandDropsUnboundRequiredContext(t *testing.T) {
	schema := &model.AuthorizationSchema{
		Permissions: []model.PermissionDefinition{
			{
				Name:                      model.PermissionGetWorkbench,
				RequiredContextDimensions: []model.ContextDimension{model.ContextWorkbench},
			},
		},
		Roles: []*model.RoleDefinition{
			// WorkspaceAdmin shape: binds workspace, grants a workbench-scoped permission.
			{Name: model.RoleWorkspaceAdmin, Permissions: []model.PermissionName{model.PermissionGetWorkbench}},
			// WorkbenchAdmin shape: binds workbench, grants the same permission.
			{Name: model.RoleWorkbenchAdmin, Permissions: []model.PermissionName{model.PermissionGetWorkbench}},
		},
	}

	svc, err := newTestAuthorizationService(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		name  string
		roles []model.Role
		check model.Permission
		want  bool
	}{
		{
			name:  "WorkspaceAdmin (binds workspace, not workbench) denied on any workbench — the bug",
			roles: []model.Role{{Name: model.RoleWorkspaceAdmin, Context: model.Context{model.ContextWorkspace: "4"}}},
			check: model.NewPermission(model.PermissionGetWorkbench, model.WithWorkbench(7)),
			want:  false,
		},
		{
			name:  "WorkbenchAdmin bound to workbench 7 allowed on workbench 7",
			roles: []model.Role{{Name: model.RoleWorkbenchAdmin, Context: model.Context{model.ContextWorkbench: "7"}}},
			check: model.NewPermission(model.PermissionGetWorkbench, model.WithWorkbench(7)),
			want:  true,
		},
		{
			name:  "WorkbenchAdmin bound to workbench 9 denied on workbench 7",
			roles: []model.Role{{Name: model.RoleWorkbenchAdmin, Context: model.Context{model.ContextWorkbench: "9"}}},
			check: model.NewPermission(model.PermissionGetWorkbench, model.WithWorkbench(7)),
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.IsUserAllowed(tt.roles, tt.check)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("IsUserAllowed = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestExpandKeepsZeroContextAndWildcard covers the non-regression cases: a
// context-free permission still matches, and a wildcard binding matches any value.
func TestExpandKeepsZeroContextAndWildcard(t *testing.T) {
	schema := &model.AuthorizationSchema{
		Permissions: []model.PermissionDefinition{
			{Name: model.PermissionCreateWorkspace}, // no required context
			{
				Name:                      model.PermissionGetWorkbench,
				RequiredContextDimensions: []model.ContextDimension{model.ContextWorkbench},
			},
		},
		Roles: []*model.RoleDefinition{
			{Name: model.RoleAuthenticated, Permissions: []model.PermissionName{model.PermissionCreateWorkspace}},
			{Name: model.RoleSuperAdmin, Permissions: []model.PermissionName{model.PermissionGetWorkbench}},
		},
	}
	svc, err := newTestAuthorizationService(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// context-free permission matches with no context supplied.
	if allowed, err := svc.IsUserAllowed(
		[]model.Role{{Name: model.RoleAuthenticated}},
		model.NewPermission(model.PermissionCreateWorkspace),
	); err != nil || !allowed {
		t.Fatalf("context-free permission should be allowed: allowed=%v err=%v", allowed, err)
	}

	// wildcard workbench binding matches any workbench.
	if allowed, err := svc.IsUserAllowed(
		[]model.Role{{Name: model.RoleSuperAdmin, Context: model.Context{model.ContextWorkbench: model.Wildcard}}},
		model.NewPermission(model.PermissionGetWorkbench, model.WithWorkbench(999)),
	); err != nil || !allowed {
		t.Fatalf("wildcard workbench should match any workbench: allowed=%v err=%v", allowed, err)
	}
}
