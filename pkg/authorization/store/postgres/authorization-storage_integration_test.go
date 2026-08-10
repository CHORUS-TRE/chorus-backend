//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	"github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	integration "github.com/CHORUS-TRE/chorus-backend/tests/integration/postgres"
)

func findUsers(ctx context.Context, store *AuthorizationStorage, tenantID uint64, filter model.FindUsersWithPermissionFilter) ([]uint64, error) {
	return store.FindUsersWithPermission(ctx, tenantID, filter, testRolesGrantingPermissions()[filter.PermissionName])
}

func testRolesGrantingPermissions() map[model.PermissionName][]model.RoleName {
	return map[model.PermissionName][]model.RoleName{
		model.ListWorkspaces.Name: {
			model.RoleWorkspaceAdmin,
			model.RoleWorkspaceMember,
			model.RoleWorkspaceGuest,
		},
		model.CreateWorkspace.Name: {
			model.RolePlatformWorkspaceManager,
			model.RoleSuperAdmin,
		},
		model.ApproveRequest.Name: {
			model.RoleWorkspaceDataManager,
			model.RoleWorkspaceAdmin,
			model.RoleSuperAdmin,
		},
	}
}

const (
	testTenantID       = uint64(88888)
	testTenant2ID      = uint64(88889)
	testUserAliceID    = uint64(90000)
	testUserBobID      = uint64(90001)
	testUserCharlieID  = uint64(90002)
	testUserInactiveID = uint64(90003)
	testUserOtherID    = uint64(90004)

	testUserRoleBaseID = uint64(92000)
)

type testFixtures struct {
	tenantID       uint64
	userIDs        map[string]uint64
	roleIDs        map[string]uint64
	userRoleNextID uint64
}

func setupTestFixtures(t *testing.T, db *sqlx.DB) testFixtures {
	t.Helper()
	ctx := context.Background()

	_, err := db.ExecContext(ctx, `INSERT INTO tenants (id, name, createdat, updatedat) VALUES ($1, 'test_tenant', NOW(), NOW())`, testTenantID)
	require.NoError(t, err)

	userIDs := map[string]uint64{
		"alice":         testUserAliceID,
		"bob":           testUserBobID,
		"charlie":       testUserCharlieID,
		"inactive_user": testUserInactiveID,
	}
	users := []struct {
		id     uint64
		name   string
		status string
	}{
		{testUserAliceID, "alice", "active"},
		{testUserBobID, "bob", "active"},
		{testUserCharlieID, "charlie", "active"},
		{testUserInactiveID, "inactive_user", "inactive"},
	}
	for _, u := range users {
		_, err := db.ExecContext(ctx, `
			INSERT INTO users (id, tenantid, firstname, lastname, username, status, createdat, updatedat)
			VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		`, u.id, testTenantID, u.name, u.name, u.name+"@test.com", u.status)
		require.NoError(t, err)
	}

	roleIDs := make(map[string]uint64)
	rows, err := db.QueryContext(ctx, `SELECT id, name FROM role_definitions`)
	require.NoError(t, err)
	defer rows.Close()
	for rows.Next() {
		var id uint64
		var name string
		require.NoError(t, rows.Scan(&id, &name))
		roleIDs[name] = id
	}

	return testFixtures{
		tenantID:       testTenantID,
		userIDs:        userIDs,
		roleIDs:        roleIDs,
		userRoleNextID: testUserRoleBaseID,
	}
}

func assignRole(t *testing.T, db *sqlx.DB, fixtures *testFixtures, userID, roleID uint64) uint64 {
	t.Helper()
	userRoleID := fixtures.userRoleNextID
	fixtures.userRoleNextID++
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO user_role (id, userid, roleid) VALUES ($1, $2, $3)
	`, userRoleID, userID, roleID)
	require.NoError(t, err)
	return userRoleID
}

func assignRoleWithContext(t *testing.T, db *sqlx.DB, fixtures *testFixtures, userID, roleID uint64, contextDim, contextValue string) uint64 {
	t.Helper()
	userRoleID := assignRole(t, db, fixtures, userID, roleID)
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO user_role_context (userroleid, contextdimension, value) VALUES ($1, $2, $3)
	`, userRoleID, contextDim, contextValue)
	require.NoError(t, err)
	return userRoleID
}

func TestAuthorizationStorage_FindUsersWithPermission_NoRolesGrant(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() {
		integration.CleanupTables(db)
	})

	fixtures := setupTestFixtures(t, db)
	assignRole(t, db, &fixtures, fixtures.userIDs["alice"], fixtures.roleIDs["Authenticated"])

	store := NewAuthorizationStorage(db)

	userIDs, err := findUsers(context.Background(), store, fixtures.tenantID, model.FindUsersWithPermissionFilter{
		PermissionName: model.DeleteApp.Name,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "no roles grant permission")
	require.Nil(t, userIDs)
}

func TestAuthorizationStorage_FindUsersWithPermission_NoContext(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() {
		integration.CleanupTables(db)
	})

	fixtures := setupTestFixtures(t, db)

	assignRole(t, db, &fixtures, fixtures.userIDs["alice"], fixtures.roleIDs["PlatformWorkspaceManager"])
	assignRole(t, db, &fixtures, fixtures.userIDs["bob"], fixtures.roleIDs["PlatformWorkspaceManager"])
	assignRole(t, db, &fixtures, fixtures.userIDs["inactive_user"], fixtures.roleIDs["PlatformWorkspaceManager"])

	store := NewAuthorizationStorage(db)

	userIDs, err := findUsers(context.Background(), store, fixtures.tenantID, model.FindUsersWithPermissionFilter{
		PermissionName: model.CreateWorkspace.Name,
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []uint64{fixtures.userIDs["alice"], fixtures.userIDs["bob"]}, userIDs)
}

func TestAuthorizationStorage_FindUsersWithPermission_WithContext(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() {
		integration.CleanupTables(db)
	})

	fixtures := setupTestFixtures(t, db)

	assignRoleWithContext(t, db, &fixtures, fixtures.userIDs["alice"], fixtures.roleIDs["WorkspaceAdmin"], "workspace", "100")
	assignRoleWithContext(t, db, &fixtures, fixtures.userIDs["bob"], fixtures.roleIDs["WorkspaceMember"], "workspace", "100")
	assignRoleWithContext(t, db, &fixtures, fixtures.userIDs["charlie"], fixtures.roleIDs["WorkspaceMember"], "workspace", "200")

	store := NewAuthorizationStorage(db)

	userIDs, err := findUsers(context.Background(), store, fixtures.tenantID, model.FindUsersWithPermissionFilter{
		PermissionName: model.ListWorkspaces.Name,
		Context: model.Context{
			model.ContextWorkspace: "100",
		},
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []uint64{fixtures.userIDs["alice"], fixtures.userIDs["bob"]}, userIDs)
}

func TestAuthorizationStorage_FindUsersWithPermission_WithWildcardContext(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() {
		integration.CleanupTables(db)
	})

	fixtures := setupTestFixtures(t, db)

	assignRoleWithContext(t, db, &fixtures, fixtures.userIDs["alice"], fixtures.roleIDs["WorkspaceAdmin"], "workspace", "*")
	assignRoleWithContext(t, db, &fixtures, fixtures.userIDs["bob"], fixtures.roleIDs["WorkspaceMember"], "workspace", "100")

	store := NewAuthorizationStorage(db)

	userIDs, err := findUsers(context.Background(), store, fixtures.tenantID, model.FindUsersWithPermissionFilter{
		PermissionName: model.ListWorkspaces.Name,
		Context: model.Context{
			model.ContextWorkspace: "999",
		},
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []uint64{fixtures.userIDs["alice"]}, userIDs)
}

func TestAuthorizationStorage_FindUsersWithPermission_ViaRolesFilter(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() {
		integration.CleanupTables(db)
	})

	fixtures := setupTestFixtures(t, db)

	assignRoleWithContext(t, db, &fixtures, fixtures.userIDs["alice"], fixtures.roleIDs["WorkspaceDataManager"], "workspace", "100")
	assignRoleWithContext(t, db, &fixtures, fixtures.userIDs["bob"], fixtures.roleIDs["WorkspaceAdmin"], "workspace", "100")
	assignRole(t, db, &fixtures, fixtures.userIDs["charlie"], fixtures.roleIDs["SuperAdmin"])

	store := NewAuthorizationStorage(db)

	userIDs, err := findUsers(context.Background(), store, fixtures.tenantID, model.FindUsersWithPermissionFilter{
		PermissionName: model.ApproveRequest.Name,
		Context: model.Context{
			model.ContextWorkspace: "100",
		},
		ViaRoles: []model.RoleName{model.RoleWorkspaceDataManager},
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []uint64{fixtures.userIDs["alice"]}, userIDs)
}

func TestAuthorizationStorage_FindUsersWithPermission_PreferExactContextMatch(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() {
		integration.CleanupTables(db)
	})

	fixtures := setupTestFixtures(t, db)

	assignRoleWithContext(t, db, &fixtures, fixtures.userIDs["alice"], fixtures.roleIDs["WorkspaceAdmin"], "workspace", "100")
	assignRoleWithContext(t, db, &fixtures, fixtures.userIDs["bob"], fixtures.roleIDs["WorkspaceAdmin"], "workspace", "*")

	store := NewAuthorizationStorage(db)

	userIDs, err := findUsers(context.Background(), store, fixtures.tenantID, model.FindUsersWithPermissionFilter{
		PermissionName: model.ListWorkspaces.Name,
		Context: model.Context{
			model.ContextWorkspace: "100",
		},
		PreferExactContextMatch: true,
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []uint64{fixtures.userIDs["alice"]}, userIDs)
}

func TestAuthorizationStorage_FindUsersWithPermission_PreferExactContextMatch_FallbackToWildcard(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() {
		integration.CleanupTables(db)
	})

	fixtures := setupTestFixtures(t, db)

	assignRoleWithContext(t, db, &fixtures, fixtures.userIDs["bob"], fixtures.roleIDs["WorkspaceAdmin"], "workspace", "*")

	store := NewAuthorizationStorage(db)

	userIDs, err := findUsers(context.Background(), store, fixtures.tenantID, model.FindUsersWithPermissionFilter{
		PermissionName: model.ListWorkspaces.Name,
		Context: model.Context{
			model.ContextWorkspace: "999",
		},
		PreferExactContextMatch: true,
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []uint64{fixtures.userIDs["bob"]}, userIDs)
}

func TestAuthorizationStorage_FindUsersWithPermission_NoMatchingViaRoles(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() {
		integration.CleanupTables(db)
	})

	fixtures := setupTestFixtures(t, db)

	assignRoleWithContext(t, db, &fixtures, fixtures.userIDs["alice"], fixtures.roleIDs["WorkspaceAdmin"], "workspace", "100")

	store := NewAuthorizationStorage(db)

	userIDs, err := findUsers(context.Background(), store, fixtures.tenantID, model.FindUsersWithPermissionFilter{
		PermissionName: model.ApproveRequest.Name,
		Context: model.Context{
			model.ContextWorkspace: "100",
		},
		ViaRoles: []model.RoleName{model.RoleWorkbenchAdmin},
	})
	require.NoError(t, err)
	require.Empty(t, userIDs)
}

func TestAuthorizationStorage_FindUsersWithPermission_MultiTenant(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() {
		integration.CleanupTables(db)
	})

	fixtures := setupTestFixtures(t, db)

	_, err = db.ExecContext(context.Background(), `INSERT INTO tenants (id, name, createdat, updatedat) VALUES ($1, 'tenant2', NOW(), NOW())`, testTenant2ID)
	require.NoError(t, err)

	_, err = db.ExecContext(context.Background(), `
		INSERT INTO users (id, tenantid, firstname, lastname, username, status, createdat, updatedat)
		VALUES ($1, $2, 'other', 'other', 'other@test.com', 'active', NOW(), NOW())
	`, testUserOtherID, testTenant2ID)
	require.NoError(t, err)

	assignRole(t, db, &fixtures, fixtures.userIDs["alice"], fixtures.roleIDs["PlatformWorkspaceManager"])
	assignRole(t, db, &fixtures, testUserOtherID, fixtures.roleIDs["PlatformWorkspaceManager"])

	store := NewAuthorizationStorage(db)

	userIDs, err := findUsers(context.Background(), store, fixtures.tenantID, model.FindUsersWithPermissionFilter{
		PermissionName: model.CreateWorkspace.Name,
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []uint64{fixtures.userIDs["alice"]}, userIDs)

	userIDs2, err := findUsers(context.Background(), store, testTenant2ID, model.FindUsersWithPermissionFilter{
		PermissionName: model.CreateWorkspace.Name,
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []uint64{testUserOtherID}, userIDs2)
}
