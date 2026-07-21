//go:build unit

package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	authorization_model "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	notification_model "github.com/CHORUS-TRE/chorus-backend/pkg/notification/model"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	user_service "github.com/CHORUS-TRE/chorus-backend/pkg/user/service"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/model"
	"github.com/CHORUS-TRE/chorus-backend/tests/unit"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockWorkbenchStore struct {
	WorkbenchStore
	getWorkbench func(ctx context.Context, tenantID, workbenchID uint64) (*model.Workbench, error)
}

func (m *mockWorkbenchStore) GetWorkbench(ctx context.Context, tenantID, workbenchID uint64) (*model.Workbench, error) {
	if m.getWorkbench != nil {
		return m.getWorkbench(ctx, tenantID, workbenchID)
	}
	return &model.Workbench{ID: workbenchID}, nil
}

type mockUserer struct {
	user_service.Userer
	getUser            func(ctx context.Context, req user_service.GetUserReq) (*user_model.User, error)
	capturedRoles      []user_model.UserRole
	removedRoleIDs     []uint64
	createUserRolesErr error
}

func (m *mockUserer) GetUser(ctx context.Context, req user_service.GetUserReq) (*user_model.User, error) {
	if m.getUser != nil {
		return m.getUser(ctx, req)
	}
	return &user_model.User{}, nil
}

func (m *mockUserer) CreateUserRoles(_ context.Context, _, _ uint64, roles []user_model.UserRole) error {
	m.capturedRoles = append(m.capturedRoles, roles...)
	return m.createUserRolesErr
}

func (m *mockUserer) RemoveUserRoles(_ context.Context, _, _ uint64, roleIDs []uint64) error {
	m.removedRoleIDs = append(m.removedRoleIDs, roleIDs...)
	return nil
}

type mockNotificationStore struct{}

func (m *mockNotificationStore) CreateNotification(_ context.Context, _ *notification_model.Notification, _ []uint64) error {
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func init() {
	unit.InitTestLogger()
}

func newSvc(store WorkbenchStore, userer user_service.Userer) *WorkbenchService {
	return &WorkbenchService{
		store:             store,
		userer:            userer,
		notificationStore: &mockNotificationStore{},
	}
}

func workbenchRole(id, workbenchID uint64, name authorization_model.RoleName) user_model.UserRole {
	return user_model.UserRole{
		ID: id,
		Role: authorization_model.Role{
			Name: name,
			Context: authorization_model.Context{
				authorization_model.ContextWorkbench: fmt.Sprintf("%d", workbenchID),
			},
		},
	}
}

func userWithRoles(roles ...user_model.UserRole) func(context.Context, user_service.GetUserReq) (*user_model.User, error) {
	return func(_ context.Context, _ user_service.GetUserReq) (*user_model.User, error) {
		return &user_model.User{ID: 42, Roles: roles}, nil
	}
}

func workbenchStoreReturning(workspaceID uint64) *mockWorkbenchStore {
	return &mockWorkbenchStore{
		getWorkbench: func(_ context.Context, _, workbenchID uint64) (*model.Workbench, error) {
			return &model.Workbench{ID: workbenchID, WorkspaceID: workspaceID, Name: "wb"}, nil
		},
	}
}

// ---------------------------------------------------------------------------
// AddUserRoleInWorkbench
// ---------------------------------------------------------------------------

func TestAddUserRoleInWorkbench_AssignsRoleWhenNoneExists(t *testing.T) {
	userer := &mockUserer{getUser: userWithRoles()}

	svc := newSvc(workbenchStoreReturning(5), userer)
	err := svc.AddUserRoleInWorkbench(context.Background(), 1, 42, workbenchRole(0, 7, authorization_model.RoleWorkbenchAdmin))

	require.NoError(t, err)
	assert.Empty(t, userer.removedRoleIDs)
	require.Len(t, userer.capturedRoles, 1)
	assert.Equal(t, "7", userer.capturedRoles[0].Context["workbench"])
	assert.Equal(t, "5", userer.capturedRoles[0].Context["workspace"])
}

// A user may hold only a single role per workbench, so adding a role replaces
// the existing one in that workbench.
func TestAddUserRoleInWorkbench_ReplacesExistingRole(t *testing.T) {
	userer := &mockUserer{
		getUser: userWithRoles(workbenchRole(2, 7, authorization_model.RoleWorkbenchAdmin)),
	}

	svc := newSvc(workbenchStoreReturning(5), userer)
	err := svc.AddUserRoleInWorkbench(context.Background(), 1, 42, workbenchRole(0, 7, authorization_model.RoleWorkbenchAdmin))

	require.NoError(t, err)
	assert.Equal(t, []uint64{2}, userer.removedRoleIDs)
	require.Len(t, userer.capturedRoles, 1)
}

// A role in a different workbench is left untouched when assigning a role in
// another workbench.
func TestAddUserRoleInWorkbench_KeepsRolesInOtherWorkbenches(t *testing.T) {
	userer := &mockUserer{
		getUser: userWithRoles(workbenchRole(2, 9, authorization_model.RoleWorkbenchAdmin)),
	}

	svc := newSvc(workbenchStoreReturning(5), userer)
	err := svc.AddUserRoleInWorkbench(context.Background(), 1, 42, workbenchRole(0, 7, authorization_model.RoleWorkbenchAdmin))

	require.NoError(t, err)
	assert.Empty(t, userer.removedRoleIDs)
	require.Len(t, userer.capturedRoles, 1)
}

// An unparseable workbench context is rejected before any role change.
func TestAddUserRoleInWorkbench_InvalidWorkbenchContext(t *testing.T) {
	userer := &mockUserer{getUser: userWithRoles()}

	role := workbenchRole(0, 0, authorization_model.RoleWorkbenchAdmin)
	role.Context[authorization_model.ContextWorkbench] = "not-a-number"

	svc := newSvc(&mockWorkbenchStore{}, userer)
	err := svc.AddUserRoleInWorkbench(context.Background(), 1, 42, role)

	require.Error(t, err)
	assert.Empty(t, userer.capturedRoles)
	assert.Empty(t, userer.removedRoleIDs)
}

// A missing workbench is surfaced as an error and assigns nothing.
func TestAddUserRoleInWorkbench_WorkbenchNotFound(t *testing.T) {
	store := &mockWorkbenchStore{
		getWorkbench: func(_ context.Context, _, _ uint64) (*model.Workbench, error) {
			return nil, errors.New("not found")
		},
	}
	userer := &mockUserer{getUser: userWithRoles()}

	svc := newSvc(store, userer)
	err := svc.AddUserRoleInWorkbench(context.Background(), 1, 42, workbenchRole(0, 7, authorization_model.RoleWorkbenchAdmin))

	require.Error(t, err)
	assert.Empty(t, userer.capturedRoles)
}

// ---------------------------------------------------------------------------
// RemoveUserFromWorkbench
// ---------------------------------------------------------------------------

// The user's role in the workbench is removed.
func TestRemoveUserFromWorkbench_RemovesWorkbenchRole(t *testing.T) {
	userer := &mockUserer{
		getUser: userWithRoles(workbenchRole(2, 7, authorization_model.RoleWorkbenchAdmin)),
	}

	svc := newSvc(workbenchStoreReturning(5), userer)
	err := svc.RemoveUserFromWorkbench(context.Background(), 1, 42, 7)

	require.NoError(t, err)
	assert.Equal(t, []uint64{2}, userer.removedRoleIDs)
}

// A user without a role in the workbench is a no-op (no removal, no error).
func TestRemoveUserFromWorkbench_NoRoleIsNoop(t *testing.T) {
	userer := &mockUserer{
		getUser: userWithRoles(workbenchRole(2, 9, authorization_model.RoleWorkbenchAdmin)),
	}

	svc := newSvc(workbenchStoreReturning(5), userer)
	err := svc.RemoveUserFromWorkbench(context.Background(), 1, 42, 7)

	require.NoError(t, err)
	assert.Empty(t, userer.removedRoleIDs)
}

// A missing workbench is surfaced as an error and removes nothing.
func TestRemoveUserFromWorkbench_WorkbenchNotFound(t *testing.T) {
	store := &mockWorkbenchStore{
		getWorkbench: func(_ context.Context, _, _ uint64) (*model.Workbench, error) {
			return nil, errors.New("not found")
		},
	}
	userer := &mockUserer{getUser: userWithRoles()}

	svc := newSvc(store, userer)
	err := svc.RemoveUserFromWorkbench(context.Background(), 1, 42, 7)

	require.Error(t, err)
	assert.Empty(t, userer.removedRoleIDs)
}
