//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"

	authorization_model "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	integration "github.com/CHORUS-TRE/chorus-backend/tests/integration/postgres"
)

func testRolesGrantingPermissions() map[authorization_model.PermissionName][]authorization_model.RoleName {
	return map[authorization_model.PermissionName][]authorization_model.RoleName{
		authorization_model.PermissionListWorkspaces: {
			authorization_model.RoleWorkspaceAdmin,
			authorization_model.RoleWorkspaceMember,
			authorization_model.RoleWorkspaceGuest,
		},
		authorization_model.PermissionCreateWorkspace: {
			authorization_model.RoleAuthenticated,
		},
		authorization_model.PermissionApproveRequest: {
			authorization_model.RoleWorkspacePI,
			authorization_model.RoleWorkspaceAdmin,
			authorization_model.RoleSuperAdmin,
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

func TestUserPermissionStorage_FindUsersWithPermission_NoRolesGrant(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() {
		integration.CleanupTables(db)
	})

	fixtures := setupTestFixtures(t, db)
	assignRole(t, db, &fixtures, fixtures.userIDs["alice"], fixtures.roleIDs["Authenticated"])

	store := NewUserPermissionStorage(db, testRolesGrantingPermissions())

	userIDs, err := store.FindUsersWithPermission(context.Background(), fixtures.tenantID, authorization_model.FindUsersWithPermissionFilter{
		PermissionName: authorization_model.PermissionDeleteApp,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "no roles grant permission")
	require.Nil(t, userIDs)
}

func TestUserPermissionStorage_FindUsersWithPermission_NoContext(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() {
		integration.CleanupTables(db)
	})

	fixtures := setupTestFixtures(t, db)

	assignRole(t, db, &fixtures, fixtures.userIDs["alice"], fixtures.roleIDs["Authenticated"])
	assignRole(t, db, &fixtures, fixtures.userIDs["bob"], fixtures.roleIDs["Authenticated"])
	assignRole(t, db, &fixtures, fixtures.userIDs["inactive_user"], fixtures.roleIDs["Authenticated"])

	store := NewUserPermissionStorage(db, testRolesGrantingPermissions())

	userIDs, err := store.FindUsersWithPermission(context.Background(), fixtures.tenantID, authorization_model.FindUsersWithPermissionFilter{
		PermissionName: authorization_model.PermissionCreateWorkspace,
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []uint64{fixtures.userIDs["alice"], fixtures.userIDs["bob"]}, userIDs)
}

func TestUserPermissionStorage_FindUsersWithPermission_WithContext(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() {
		integration.CleanupTables(db)
	})

	fixtures := setupTestFixtures(t, db)

	assignRoleWithContext(t, db, &fixtures, fixtures.userIDs["alice"], fixtures.roleIDs["WorkspaceAdmin"], "workspace", "100")
	assignRoleWithContext(t, db, &fixtures, fixtures.userIDs["bob"], fixtures.roleIDs["WorkspaceMember"], "workspace", "100")
	assignRoleWithContext(t, db, &fixtures, fixtures.userIDs["charlie"], fixtures.roleIDs["WorkspaceMember"], "workspace", "200")

	store := NewUserPermissionStorage(db, testRolesGrantingPermissions())

	userIDs, err := store.FindUsersWithPermission(context.Background(), fixtures.tenantID, authorization_model.FindUsersWithPermissionFilter{
		PermissionName: authorization_model.PermissionListWorkspaces,
		Context: authorization_model.Context{
			authorization_model.RoleContextWorkspace: "100",
		},
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []uint64{fixtures.userIDs["alice"], fixtures.userIDs["bob"]}, userIDs)
}

func TestUserPermissionStorage_FindUsersWithPermission_WithWildcardContext(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() {
		integration.CleanupTables(db)
	})

	fixtures := setupTestFixtures(t, db)

	assignRoleWithContext(t, db, &fixtures, fixtures.userIDs["alice"], fixtures.roleIDs["WorkspaceAdmin"], "workspace", "*")
	assignRoleWithContext(t, db, &fixtures, fixtures.userIDs["bob"], fixtures.roleIDs["WorkspaceMember"], "workspace", "100")

	store := NewUserPermissionStorage(db, testRolesGrantingPermissions())

	userIDs, err := store.FindUsersWithPermission(context.Background(), fixtures.tenantID, authorization_model.FindUsersWithPermissionFilter{
		PermissionName: authorization_model.PermissionListWorkspaces,
		Context: authorization_model.Context{
			authorization_model.RoleContextWorkspace: "999",
		},
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []uint64{fixtures.userIDs["alice"]}, userIDs)
}

func TestUserPermissionStorage_FindUsersWithPermission_ViaRolesFilter(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() {
		integration.CleanupTables(db)
	})

	fixtures := setupTestFixtures(t, db)

	assignRoleWithContext(t, db, &fixtures, fixtures.userIDs["alice"], fixtures.roleIDs["WorkspacePI"], "workspace", "100")
	assignRoleWithContext(t, db, &fixtures, fixtures.userIDs["bob"], fixtures.roleIDs["WorkspaceAdmin"], "workspace", "100")
	assignRole(t, db, &fixtures, fixtures.userIDs["charlie"], fixtures.roleIDs["SuperAdmin"])

	store := NewUserPermissionStorage(db, testRolesGrantingPermissions())

	userIDs, err := store.FindUsersWithPermission(context.Background(), fixtures.tenantID, authorization_model.FindUsersWithPermissionFilter{
		PermissionName: authorization_model.PermissionApproveRequest,
		Context: authorization_model.Context{
			authorization_model.RoleContextWorkspace: "100",
		},
		ViaRoles: []authorization_model.RoleName{authorization_model.RoleWorkspacePI},
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []uint64{fixtures.userIDs["alice"]}, userIDs)
}

func TestUserPermissionStorage_FindUsersWithPermission_PreferExactContextMatch(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() {
		integration.CleanupTables(db)
	})

	fixtures := setupTestFixtures(t, db)

	assignRoleWithContext(t, db, &fixtures, fixtures.userIDs["alice"], fixtures.roleIDs["WorkspaceAdmin"], "workspace", "100")
	assignRoleWithContext(t, db, &fixtures, fixtures.userIDs["bob"], fixtures.roleIDs["WorkspaceAdmin"], "workspace", "*")

	store := NewUserPermissionStorage(db, testRolesGrantingPermissions())

	userIDs, err := store.FindUsersWithPermission(context.Background(), fixtures.tenantID, authorization_model.FindUsersWithPermissionFilter{
		PermissionName: authorization_model.PermissionListWorkspaces,
		Context: authorization_model.Context{
			authorization_model.RoleContextWorkspace: "100",
		},
		PreferExactContextMatch: true,
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []uint64{fixtures.userIDs["alice"]}, userIDs)
}

func TestUserPermissionStorage_FindUsersWithPermission_PreferExactContextMatch_FallbackToWildcard(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() {
		integration.CleanupTables(db)
	})

	fixtures := setupTestFixtures(t, db)

	assignRoleWithContext(t, db, &fixtures, fixtures.userIDs["bob"], fixtures.roleIDs["WorkspaceAdmin"], "workspace", "*")

	store := NewUserPermissionStorage(db, testRolesGrantingPermissions())

	userIDs, err := store.FindUsersWithPermission(context.Background(), fixtures.tenantID, authorization_model.FindUsersWithPermissionFilter{
		PermissionName: authorization_model.PermissionListWorkspaces,
		Context: authorization_model.Context{
			authorization_model.RoleContextWorkspace: "999",
		},
		PreferExactContextMatch: true,
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []uint64{fixtures.userIDs["bob"]}, userIDs)
}

func TestUserPermissionStorage_FindUsersWithPermission_NoMatchingViaRoles(t *testing.T) {
	db, err := integration.GetDB()
	require.NoError(t, err)
	t.Cleanup(func() {
		integration.CleanupTables(db)
	})

	fixtures := setupTestFixtures(t, db)

	assignRoleWithContext(t, db, &fixtures, fixtures.userIDs["alice"], fixtures.roleIDs["WorkspaceAdmin"], "workspace", "100")

	store := NewUserPermissionStorage(db, testRolesGrantingPermissions())

	userIDs, err := store.FindUsersWithPermission(context.Background(), fixtures.tenantID, authorization_model.FindUsersWithPermissionFilter{
		PermissionName: authorization_model.PermissionApproveRequest,
		Context: authorization_model.Context{
			authorization_model.RoleContextWorkspace: "100",
		},
		ViaRoles: []authorization_model.RoleName{authorization_model.RoleWorkbenchAdmin},
	})
	require.NoError(t, err)
	require.Empty(t, userIDs)
}

func TestUserPermissionStorage_FindUsersWithPermission_MultiTenant(t *testing.T) {
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

	assignRole(t, db, &fixtures, fixtures.userIDs["alice"], fixtures.roleIDs["Authenticated"])
	assignRole(t, db, &fixtures, testUserOtherID, fixtures.roleIDs["Authenticated"])

	store := NewUserPermissionStorage(db, testRolesGrantingPermissions())

	userIDs, err := store.FindUsersWithPermission(context.Background(), fixtures.tenantID, authorization_model.FindUsersWithPermissionFilter{
		PermissionName: authorization_model.PermissionCreateWorkspace,
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []uint64{fixtures.userIDs["alice"]}, userIDs)

	userIDs2, err := store.FindUsersWithPermission(context.Background(), testTenant2ID, authorization_model.FindUsersWithPermissionFilter{
		PermissionName: authorization_model.PermissionCreateWorkspace,
	})
	require.NoError(t, err)
	require.ElementsMatch(t, []uint64{testUserOtherID}, userIDs2)
}
