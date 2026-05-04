package authorization

import (
	"reflect"
	"sort"
	"testing"

	"github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
)

func buildSchema(roles map[model.RoleName][]model.PermissionName) *model.AuthorizationSchema {
	permissionSet := make(map[model.PermissionName]bool)
	for _, permissions := range roles {
		for _, permission := range permissions {
			permissionSet[permission] = true
		}
	}

	schema := &model.AuthorizationSchema{}
	for permission := range permissionSet {
		schema.Permissions = append(schema.Permissions, model.PermissionDefinition{Name: permission})
	}
	for roleName, permissions := range roles {
		schema.Roles = append(schema.Roles, &model.RoleDefinition{Name: roleName, Permissions: permissions})
	}
	return schema
}

func TestAuthorizationServiceValidatesFlatSchema(t *testing.T) {
	tests := []struct {
		name      string
		schema    *model.AuthorizationSchema
		expectErr bool
	}{
		{
			name:   "valid direct permissions",
			schema: buildSchema(map[model.RoleName][]model.PermissionName{"admin": {model.PermissionGetWorkspace}}),
		},
		{
			name: "duplicate role",
			schema: &model.AuthorizationSchema{
				Roles: []*model.RoleDefinition{
					{Name: "admin"},
					{Name: "admin"},
				},
			},
			expectErr: true,
		},
		{
			name: "duplicate permission",
			schema: &model.AuthorizationSchema{
				Permissions: []model.PermissionDefinition{
					{Name: model.PermissionGetWorkspace},
					{Name: model.PermissionGetWorkspace},
				},
			},
			expectErr: true,
		},
		{
			name: "unknown role permission",
			schema: &model.AuthorizationSchema{
				Roles: []*model.RoleDefinition{{Name: "admin", Permissions: []model.PermissionName{model.PermissionGetWorkspace}}},
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAuthorizationService(tt.schema)
			if tt.expectErr && err == nil {
				t.Fatal("expected error but got none")
			}
			if !tt.expectErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestGetUserPermissionsUsesDirectRolesOnly(t *testing.T) {
	policy, err := NewAuthorizationService(buildSchema(map[model.RoleName][]model.PermissionName{
		"viewer": {model.PermissionGetWorkspace},
		"admin":  {},
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	permissions, err := policy.GetUserPermissions([]model.Role{{Name: "admin"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(permissions) != 0 {
		t.Fatalf("expected no inherited permissions, got %v", permissions)
	}
}

func TestGetUserPermissionsDeduplicatesPermissions(t *testing.T) {
	policy, err := NewAuthorizationService(buildSchema(map[model.RoleName][]model.PermissionName{
		"viewer": {model.PermissionGetWorkspace},
		"admin":  {model.PermissionGetWorkspace, model.PermissionUpdateWorkspace},
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	permissions, err := policy.GetUserPermissions([]model.Role{{Name: "viewer"}, {Name: "admin"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := make([]model.PermissionName, 0, len(permissions))
	for _, permission := range permissions {
		got = append(got, permission.Name)
	}
	sort.Slice(got, func(i, j int) bool { return got[i] < got[j] })

	expected := []model.PermissionName{model.PermissionGetWorkspace, model.PermissionUpdateWorkspace}
	if !reflect.DeepEqual(got, expected) {
		t.Errorf("GetUserPermissions() = %v, want %v", got, expected)
	}
}

func TestIsUserAllowedRequiresExplicitRolePermission(t *testing.T) {
	schema := &model.AuthorizationSchema{
		Permissions: []model.PermissionDefinition{
			{
				Name:                      model.PermissionGetWorkspace,
				RequiredContextDimensions: []model.ContextDimension{model.RoleContextWorkspace},
			},
		},
		Roles: []*model.RoleDefinition{
			{Name: model.RoleWorkspaceGuest, Permissions: []model.PermissionName{model.PermissionGetWorkspace}},
			{Name: model.RoleWorkspaceAdmin},
		},
	}

	policy, err := NewAuthorizationService(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	allowed, err := policy.IsUserAllowed(
		[]model.Role{{Name: model.RoleWorkspaceAdmin, Context: model.Context{model.RoleContextWorkspace: "42"}}},
		model.NewPermission(model.PermissionGetWorkspace, model.WithWorkspace(42)),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatal("expected role without explicit permission to be denied")
	}

	allowed, err = policy.IsUserAllowed(
		[]model.Role{{Name: model.RoleWorkspaceGuest, Context: model.Context{model.RoleContextWorkspace: "42"}}},
		model.NewPermission(model.PermissionGetWorkspace, model.WithWorkspace(42)),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("expected explicit workspace permission to be allowed")
	}
}

func TestDefaultSchemaSuperAdminGrantsEveryPermission(t *testing.T) {
	schema := model.GetDefaultSchema()
	policy, err := NewAuthorizationService(&schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	permissions, err := policy.GetUserPermissions([]model.Role{{Name: model.RoleSuperAdmin}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := make(map[model.PermissionName]bool)
	for _, permission := range permissions {
		got[permission.Name] = true
	}
	for _, permission := range schema.Permissions {
		if !got[permission.Name] {
			t.Errorf("super admin is missing permission %s", permission.Name)
		}
	}
}
