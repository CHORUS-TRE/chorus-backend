package service

import (
	"testing"

	"github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
)

// TestIsUserAllowedDefaultSchemaContextMatrix tests the policy kernel against
// the real default schema. The cases encode the intended context-matching
// semantics: OR across a user's roles, AND across the dimensions a role
// binds, wildcard only where the role definition declares
// ContextQuantifierAny, and no grant at all when a role provides none of the
// context a permission requires.
func TestIsUserAllowedDefaultSchemaContextMatrix(t *testing.T) {
	schema := model.GetDefaultSchema()
	structures, err := extractAuthorizationStructures(&schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	workbenchViewer := model.Role{Name: model.RoleWorkbenchViewer, Context: model.Context{
		model.RoleContextWorkbench: "42",
		model.RoleContextWorkspace: "5",
	}}
	workbenchAdmin := model.Role{Name: model.RoleWorkbenchAdmin, Context: model.Context{
		model.RoleContextWorkbench: "42",
		model.RoleContextWorkspace: "5",
	}}
	workspaceAdmin := model.Role{Name: model.RoleWorkspaceAdmin, Context: model.Context{
		model.RoleContextWorkspace: "5",
	}}

	tests := []struct {
		name       string
		user       []model.Role
		permission model.Permission
		allowed    bool
	}{
		{
			name:       "workbench viewer can stream own workbench",
			user:       []model.Role{workbenchViewer},
			permission: model.NewPermission(model.PermissionStreamWorkbench, model.WithWorkbench(42), model.WithWorkspace(5)),
			allowed:    true,
		},
		{
			name:       "workbench viewer cannot stream sibling workbench",
			user:       []model.Role{workbenchViewer},
			permission: model.NewPermission(model.PermissionStreamWorkbench, model.WithWorkbench(99), model.WithWorkspace(5)),
			allowed:    false,
		},
		{
			name:       "workbench admin cannot delete foreign workbench",
			user:       []model.Role{workbenchAdmin},
			permission: model.NewPermission(model.PermissionDeleteWorkbench, model.WithWorkbench(99), model.WithWorkspace(5)),
			allowed:    false,
		},
		{
			name:       "workspace admin can stream workbench in own workspace",
			user:       []model.Role{workspaceAdmin},
			permission: model.NewPermission(model.PermissionStreamWorkbench, model.WithWorkbench(42), model.WithWorkspace(5)),
			allowed:    true,
		},
		{
			name:       "workspace admin cannot stream workbench in foreign workspace",
			user:       []model.Role{workspaceAdmin},
			permission: model.NewPermission(model.PermissionStreamWorkbench, model.WithWorkbench(42), model.WithWorkspace(9)),
			allowed:    false,
		},
		{
			name:       "workspace admin cannot delete workbench in foreign workspace",
			user:       []model.Role{workspaceAdmin},
			permission: model.NewPermission(model.PermissionDeleteWorkbench, model.WithWorkbench(42), model.WithWorkspace(9)),
			allowed:    false,
		},
		{
			name:       "super admin can stream any workbench (explicit Any)",
			user:       []model.Role{{Name: model.RoleSuperAdmin}},
			permission: model.NewPermission(model.PermissionStreamWorkbench, model.WithWorkbench(42), model.WithWorkspace(9)),
			allowed:    true,
		},
		{
			name:       "platform user manager can get any user (explicit Any)",
			user:       []model.Role{{Name: model.RolePlateformUserManager}},
			permission: model.NewPermission(model.PermissionGetUser, model.WithUser(7)),
			allowed:    true,
		},
		{
			name:       "data manager can get any workspace (explicit Any)",
			user:       []model.Role{{Name: model.RoleDataManager}},
			permission: model.NewPermission(model.PermissionGetWorkspace, model.WithWorkspace(9)),
			allowed:    true,
		},
		{
			name:       "authenticated user cannot stream workbench",
			user:       []model.Role{{Name: model.RoleAuthenticated, Context: model.Context{model.RoleContextUser: "7"}}},
			permission: model.NewPermission(model.PermissionStreamWorkbench, model.WithWorkbench(42), model.WithWorkspace(5)),
			allowed:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := isUserAllowed(structures.RoleMap, structures.PermissionMap, tt.user, tt.permission)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if allowed != tt.allowed {
				t.Errorf("isUserAllowed() = %v, want %v", allowed, tt.allowed)
			}
		})
	}
}
