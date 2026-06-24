//go:build unit

package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	k8s "github.com/CHORUS-TRE/chorus-backend/internal/client/k8s"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	audit_model "github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	audit_service "github.com/CHORUS-TRE/chorus-backend/pkg/audit/service"
	authorization_model "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	notification_model "github.com/CHORUS-TRE/chorus-backend/pkg/notification/model"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	user_service "github.com/CHORUS-TRE/chorus-backend/pkg/user/service"
	workbench_model "github.com/CHORUS-TRE/chorus-backend/pkg/workbench/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
)

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockWorkspaceStore struct {
	createWorkspace      func(ctx context.Context, tenantID uint64, workspace *model.Workspace) (*model.Workspace, error)
	listPublicWorkspaces func(ctx context.Context, tenantID uint64, pagination *common_model.Pagination) ([]*model.Workspace, *common_model.PaginationResult, error)
}

func (m *mockWorkspaceStore) CreateWorkspace(ctx context.Context, tenantID uint64, workspace *model.Workspace) (*model.Workspace, error) {
	return m.createWorkspace(ctx, tenantID, workspace)
}

func (m *mockWorkspaceStore) ListPublicWorkspaces(ctx context.Context, tenantID uint64, pagination *common_model.Pagination) ([]*model.Workspace, *common_model.PaginationResult, error) {
	return m.listPublicWorkspaces(ctx, tenantID, pagination)
}

func (m *mockWorkspaceStore) GetWorkspace(_ context.Context, _ uint64, _ uint64) (*model.Workspace, error) {
	return nil, nil
}

func (m *mockWorkspaceStore) ListWorkspaces(_ context.Context, _ uint64, _ *common_model.Pagination, _ *[]uint64, _ bool) ([]*model.Workspace, *common_model.PaginationResult, error) {
	return nil, nil, nil
}

func (m *mockWorkspaceStore) DeleteOldWorkspaces(_ context.Context, _ time.Duration) ([]*model.Workspace, error) {
	return nil, nil
}

func (m *mockWorkspaceStore) UpdateWorkspace(_ context.Context, _ uint64, _ *model.Workspace) (*model.Workspace, error) {
	return nil, nil
}

func (m *mockWorkspaceStore) DeleteWorkspace(_ context.Context, _ uint64, _ uint64) error {
	return nil
}

func (m *mockWorkspaceStore) UpdateWorkspaceStatus(_ context.Context, _ uint64, _ uint64, _, _ string) error {
	return nil
}

func (m *mockWorkspaceStore) GetWorkspaceServiceInstance(_ context.Context, _, _ uint64) (*model.WorkspaceServiceInstance, error) {
	return nil, nil
}

func (m *mockWorkspaceStore) ListWorkspaceServiceInstances(_ context.Context, _ uint64, _ *common_model.Pagination, _ *[]uint64) ([]*model.WorkspaceServiceInstance, *common_model.PaginationResult, error) {
	return nil, nil, nil
}

func (m *mockWorkspaceStore) ListWorkspaceServiceInstancesByWorkspace(_ context.Context, _ uint64) ([]*model.WorkspaceServiceInstance, error) {
	return nil, nil
}

func (m *mockWorkspaceStore) CreateWorkspaceServiceInstance(_ context.Context, _ uint64, _ *model.WorkspaceServiceInstance) (*model.WorkspaceServiceInstance, error) {
	return nil, nil
}

func (m *mockWorkspaceStore) UpdateWorkspaceServiceInstance(_ context.Context, _ uint64, _ *model.WorkspaceServiceInstance) (*model.WorkspaceServiceInstance, error) {
	return nil, nil
}

func (m *mockWorkspaceStore) DeleteWorkspaceServiceInstance(_ context.Context, _, _ uint64) error {
	return nil
}

func (m *mockWorkspaceStore) UpdateWorkspaceServiceInstanceStatuses(_ context.Context, _ uint64, _ map[string]model.WorkspaceServiceInstanceStatusUpdate) error {
	return nil
}

type mockUserer struct {
	createUserRolesErr error
	capturedRoles      []user_model.UserRole
	removedRoleIDs     []uint64
	getUser            func(ctx context.Context, req user_service.GetUserReq) (*user_model.User, error)
	getUsers           func(ctx context.Context, tenantID uint64, userIDs []uint64) ([]*user_model.User, error)
}

func (m *mockUserer) CreateUserRoles(_ context.Context, _, _ uint64, roles []user_model.UserRole) error {
	m.capturedRoles = append(m.capturedRoles, roles...)
	return m.createUserRolesErr
}

func (m *mockUserer) RemoveUserRoles(_ context.Context, _, _ uint64, roleIDs []uint64) error {
	m.removedRoleIDs = append(m.removedRoleIDs, roleIDs...)
	return nil
}

func (m *mockUserer) RemoveRolesByContext(_ context.Context, _, _ string) ([]uint64, error) {
	return nil, nil
}

func (m *mockUserer) GetUser(ctx context.Context, req user_service.GetUserReq) (*user_model.User, error) {
	if m.getUser != nil {
		return m.getUser(ctx, req)
	}
	return nil, nil
}

func (m *mockUserer) GetUsers(ctx context.Context, tenantID uint64, userIDs []uint64) ([]*user_model.User, error) {
	if m.getUsers != nil {
		return m.getUsers(ctx, tenantID, userIDs)
	}
	return nil, nil
}

type mockK8s struct {
	createWorkspaceErr error
}

func (m *mockK8s) CreateWorkspace(_ k8s.WorkspaceInput) error                         { return m.createWorkspaceErr }
func (m *mockK8s) UpdateWorkspace(_ k8s.WorkspaceInput) error                         { return nil }
func (m *mockK8s) DeleteWorkspace(_ string) error                                     { return nil }
func (m *mockK8s) CreateWorkbench(_ k8s.Workbench) error                              { return nil }
func (m *mockK8s) UpdateWorkbench(_ k8s.Workbench) error                              { return nil }
func (m *mockK8s) DeleteWorkbench(_, _ string) error                                  { return nil }
func (m *mockK8s) CreateAppInstance(_, _ string, _ k8s.AppInstance) error             { return nil }
func (m *mockK8s) UpdateAppInstance(_, _ string, _ k8s.AppInstance) error             { return nil }
func (m *mockK8s) DeleteAppInstance(_, _ string, _ k8s.AppInstance) error             { return nil }
func (m *mockK8s) CreatePortForward(_, _ string) (uint16, chan struct{}, error)       { return 0, nil, nil }
func (m *mockK8s) PrePullImageOnAllNodes(_ string)                                    {}
func (m *mockK8s) RegisterOnNewWorkbenchHandler(_ func(k8s.Workbench) error) error    { return nil }
func (m *mockK8s) RegisterOnUpdateWorkbenchHandler(_ func(k8s.Workbench) error) error { return nil }
func (m *mockK8s) RegisterOnDeleteWorkbenchHandler(_ func(k8s.Workbench) error) error { return nil }
func (m *mockK8s) RegisterOnUpdateWorkspaceHandler(_ func(k8s.WorkspaceOutput) error) error {
	return nil
}

type mockWorkbencher struct{}

func (m *mockWorkbencher) ListWorkbenches(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, filter workbench_model.WorkbenchFilter) ([]*workbench_model.Workbench, *common_model.PaginationResult, error) {
	return nil, nil, nil
}

func (m *mockWorkbencher) DeleteWorkbenchesInWorkspace(_ context.Context, _, _ uint64) error {
	return nil
}

type mockNotificationStore struct{}

func (m *mockNotificationStore) CreateNotification(_ context.Context, _ *notification_model.Notification, _ []uint64) error {
	return nil
}

type mockAuditWriter struct{}

func (m *mockAuditWriter) Record(_ context.Context, _ *audit_model.AuditEntry) (*audit_model.AuditEntry, error) {
	return nil, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newSvc(cfg config.Config, store WorkspaceStore, k8sClient k8s.K8sClienter, userer Userer) *WorkspaceService {
	return &WorkspaceService{
		cfg:               cfg,
		store:             store,
		k8sClient:         k8sClient,
		workbencher:       &mockWorkbencher{},
		userer:            userer,
		notificationStore: &mockNotificationStore{},
		auditWriter:       audit_service.AuditWriter(&mockAuditWriter{}),
	}
}

func storeReturning(ws *model.Workspace) *mockWorkspaceStore {
	return &mockWorkspaceStore{
		createWorkspace: func(_ context.Context, _ uint64, _ *model.Workspace) (*model.Workspace, error) {
			return ws, nil
		},
	}
}

func wsRole(id, workspaceID uint64, name authorization_model.RoleName) user_model.UserRole {
	return user_model.UserRole{
		ID: id,
		Role: authorization_model.Role{
			Name: name,
			Context: authorization_model.Context{
				authorization_model.RoleContextWorkspace: fmt.Sprintf("%d", workspaceID),
			},
		},
	}
}

func wbRole(id, workspaceID, workbenchID uint64, name authorization_model.RoleName) user_model.UserRole {
	return user_model.UserRole{
		ID: id,
		Role: authorization_model.Role{
			Name: name,
			Context: authorization_model.Context{
				authorization_model.RoleContextWorkspace: fmt.Sprintf("%d", workspaceID),
				authorization_model.RoleContextWorkbench: fmt.Sprintf("%d", workbenchID),
			},
		},
	}
}

func userWithRoles(roles ...user_model.UserRole) func(context.Context, user_service.GetUserReq) (*user_model.User, error) {
	return func(_ context.Context, _ user_service.GetUserReq) (*user_model.User, error) {
		return &user_model.User{ID: 42, Roles: roles}, nil
	}
}

// ---------------------------------------------------------------------------
// CreateWorkspace
// ---------------------------------------------------------------------------

func TestCreateWorkspace_AssignsWorkspaceAdminWhenConfigured(t *testing.T) {
	cfg := config.Config{}
	cfg.Services.WorkspaceService.CreatorIsAdmin = true

	created := &model.Workspace{ID: 10, TenantID: 1, UserID: 42}
	userer := &mockUserer{}

	svc := newSvc(cfg, storeReturning(created), &mockK8s{}, userer)
	_, err := svc.CreateWorkspace(context.Background(), &model.Workspace{TenantID: 1, UserID: 42})

	require.NoError(t, err)
	require.Len(t, userer.capturedRoles, 1)
	assert.Equal(t, "WorkspaceAdmin", userer.capturedRoles[0].Role.Name.String())
	assert.Equal(t, "10", userer.capturedRoles[0].Role.Context["workspace"])
}

func TestCreateWorkspace_AssignsDataManagerWhenConfigured(t *testing.T) {
	cfg := config.Config{}
	cfg.Services.WorkspaceService.CreatorIsDataManager = true

	created := &model.Workspace{ID: 11, TenantID: 1, UserID: 42}
	userer := &mockUserer{}

	svc := newSvc(cfg, storeReturning(created), &mockK8s{}, userer)
	_, err := svc.CreateWorkspace(context.Background(), &model.Workspace{TenantID: 1, UserID: 42})

	require.NoError(t, err)
	require.Len(t, userer.capturedRoles, 1)
	assert.Equal(t, "WorkspaceDataManager", userer.capturedRoles[0].Role.Name.String())
}

func TestCreateWorkspace_AssignsBothRolesWhenBothConfigured(t *testing.T) {
	cfg := config.Config{}
	cfg.Services.WorkspaceService.CreatorIsAdmin = true
	cfg.Services.WorkspaceService.CreatorIsDataManager = true

	created := &model.Workspace{ID: 12, TenantID: 1, UserID: 42}
	userer := &mockUserer{}

	svc := newSvc(cfg, storeReturning(created), &mockK8s{}, userer)
	_, err := svc.CreateWorkspace(context.Background(), &model.Workspace{TenantID: 1, UserID: 42})

	require.NoError(t, err)
	assert.Len(t, userer.capturedRoles, 2)
}

func TestCreateWorkspace_DoesNotAssignRolesWhenNotConfigured(t *testing.T) {
	cfg := config.Config{} // both false by default

	created := &model.Workspace{ID: 13, TenantID: 1, UserID: 42}
	userer := &mockUserer{}

	svc := newSvc(cfg, storeReturning(created), &mockK8s{}, userer)
	_, err := svc.CreateWorkspace(context.Background(), &model.Workspace{TenantID: 1, UserID: 42})

	require.NoError(t, err)
	assert.Empty(t, userer.capturedRoles)
}

func TestCreateWorkspace_PropagatesStoreError(t *testing.T) {
	cfg := config.Config{}
	store := &mockWorkspaceStore{
		createWorkspace: func(_ context.Context, _ uint64, _ *model.Workspace) (*model.Workspace, error) {
			return nil, errors.New("db down")
		},
	}

	svc := newSvc(cfg, store, &mockK8s{}, &mockUserer{})
	_, err := svc.CreateWorkspace(context.Background(), &model.Workspace{TenantID: 1})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "db down")
}

func TestCreateWorkspace_PropagatesCreateUserRolesError(t *testing.T) {
	cfg := config.Config{}
	cfg.Services.WorkspaceService.CreatorIsAdmin = true

	created := &model.Workspace{ID: 14, TenantID: 1, UserID: 42}
	userer := &mockUserer{createUserRolesErr: errors.New("roles store down")}

	svc := newSvc(cfg, storeReturning(created), &mockK8s{}, userer)
	_, err := svc.CreateWorkspace(context.Background(), &model.Workspace{TenantID: 1, UserID: 42})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "roles store down")
}

func TestCreateWorkspace_PropagatesK8sError(t *testing.T) {
	cfg := config.Config{}

	created := &model.Workspace{ID: 15, TenantID: 1, UserID: 42}
	k8sErr := &mockK8s{createWorkspaceErr: errors.New("k8s unreachable")}

	svc := newSvc(cfg, storeReturning(created), k8sErr, &mockUserer{})
	_, err := svc.CreateWorkspace(context.Background(), &model.Workspace{TenantID: 1})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "k8s unreachable")
}

// ---------------------------------------------------------------------------
// ListPublicWorkspaces
// ---------------------------------------------------------------------------

func TestListPublicWorkspaces_EmptyContactWhenNoContactUserID(t *testing.T) {
	ws := &model.Workspace{ID: 1, TenantID: 1, Name: "My WS", ContactUserID: nil}
	store := &mockWorkspaceStore{
		listPublicWorkspaces: func(_ context.Context, _ uint64, _ *common_model.Pagination) ([]*model.Workspace, *common_model.PaginationResult, error) {
			return []*model.Workspace{ws}, nil, nil
		},
	}
	userer := &mockUserer{
		getUsers: func(_ context.Context, _ uint64, _ []uint64) ([]*user_model.User, error) {
			t.Fatal("GetUsers should not be called when there is no contact user")
			return nil, nil
		},
	}

	svc := newSvc(config.Config{}, store, &mockK8s{}, userer)
	result, _, err := svc.ListPublicWorkspaces(context.Background(), 1, &common_model.Pagination{})

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Empty(t, result[0].ContactUsername)
}

func TestListPublicWorkspaces_PopulatesContactFromUser(t *testing.T) {
	contactID := uint64(99)
	ws := &model.Workspace{ID: 1, TenantID: 1, Name: "My WS", ContactUserID: &contactID}
	store := &mockWorkspaceStore{
		listPublicWorkspaces: func(_ context.Context, _ uint64, _ *common_model.Pagination) ([]*model.Workspace, *common_model.PaginationResult, error) {
			return []*model.Workspace{ws}, nil, nil
		},
	}
	userer := &mockUserer{
		getUsers: func(_ context.Context, _ uint64, ids []uint64) ([]*user_model.User, error) {
			assert.Equal(t, []uint64{contactID}, ids)
			return []*user_model.User{{ID: contactID, Username: "jsmith", FirstName: "Jane", LastName: "Smith", Email: "jane@example.com"}}, nil
		},
	}

	svc := newSvc(config.Config{}, store, &mockK8s{}, userer)
	result, _, err := svc.ListPublicWorkspaces(context.Background(), 1, &common_model.Pagination{})

	require.NoError(t, err)
	require.Len(t, result, 1)
	assert.Equal(t, "jsmith", result[0].ContactUsername)
	assert.Equal(t, "Jane", result[0].ContactFirstName)
	assert.Equal(t, "Smith", result[0].ContactLastName)
	assert.Equal(t, "jane@example.com", result[0].ContactEmail)
}

func TestListPublicWorkspaces_PropagatesStoreError(t *testing.T) {
	store := &mockWorkspaceStore{
		listPublicWorkspaces: func(_ context.Context, _ uint64, _ *common_model.Pagination) ([]*model.Workspace, *common_model.PaginationResult, error) {
			return nil, nil, errors.New("db down")
		},
	}

	svc := newSvc(config.Config{}, store, &mockK8s{}, &mockUserer{})
	_, _, err := svc.ListPublicWorkspaces(context.Background(), 1, &common_model.Pagination{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "db down")
}

func TestListPublicWorkspaces_PropagatesGetUsersError(t *testing.T) {
	contactID := uint64(99)
	ws := &model.Workspace{ID: 1, TenantID: 1, ContactUserID: &contactID}
	store := &mockWorkspaceStore{
		listPublicWorkspaces: func(_ context.Context, _ uint64, _ *common_model.Pagination) ([]*model.Workspace, *common_model.PaginationResult, error) {
			return []*model.Workspace{ws}, nil, nil
		},
	}
	userer := &mockUserer{
		getUsers: func(_ context.Context, _ uint64, _ []uint64) ([]*user_model.User, error) {
			return nil, errors.New("user not found")
		},
	}

	svc := newSvc(config.Config{}, store, &mockK8s{}, userer)
	_, _, err := svc.ListPublicWorkspaces(context.Background(), 1, &common_model.Pagination{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

// ---------------------------------------------------------------------------
// RemoveUserRoleInWorkspace
// ---------------------------------------------------------------------------

func TestRemoveUserRoleInWorkspace_LastRoleRemovesUserFromWorkspace(t *testing.T) {
	userer := &mockUserer{
		getUser: userWithRoles(
			wsRole(1, 5, authorization_model.RoleWorkspaceAdmin),
			wbRole(2, 5, 7, authorization_model.RoleWorkbenchAdmin),
		),
	}

	svc := newSvc(config.Config{}, &mockWorkspaceStore{}, &mockK8s{}, userer)
	err := svc.RemoveUserRoleInWorkspace(context.Background(), 1, 42, 5, authorization_model.RoleWorkspaceAdmin)

	require.NoError(t, err)
	assert.ElementsMatch(t, []uint64{1, 2}, userer.removedRoleIDs)
}

func TestRemoveUserRoleInWorkspace_KeepsOtherWorkspaceRoles(t *testing.T) {
	userer := &mockUserer{
		getUser: userWithRoles(
			wsRole(1, 5, authorization_model.RoleWorkspaceAdmin),
			wsRole(3, 5, authorization_model.RoleWorkspaceDataManager),
			wbRole(2, 5, 7, authorization_model.RoleWorkbenchAdmin),
		),
	}

	svc := newSvc(config.Config{}, &mockWorkspaceStore{}, &mockK8s{}, userer)
	err := svc.RemoveUserRoleInWorkspace(context.Background(), 1, 42, 5, authorization_model.RoleWorkspaceAdmin)

	require.NoError(t, err)
	assert.Equal(t, []uint64{1}, userer.removedRoleIDs)
}

func TestRemoveUserRoleInWorkspace_WorkbenchRoleDoesNotCountAsWorkspaceRole(t *testing.T) {
	userer := &mockUserer{
		getUser: userWithRoles(
			wsRole(1, 5, authorization_model.RoleWorkspaceDataManager),
			wbRole(2, 5, 7, authorization_model.RoleWorkbenchAdmin),
		),
	}

	svc := newSvc(config.Config{}, &mockWorkspaceStore{}, &mockK8s{}, userer)
	err := svc.RemoveUserRoleInWorkspace(context.Background(), 1, 42, 5, authorization_model.RoleWorkspaceDataManager)

	require.NoError(t, err)
	assert.ElementsMatch(t, []uint64{1, 2}, userer.removedRoleIDs)
}

func TestRemoveUserRoleInWorkspace_RoleNotFound(t *testing.T) {
	userer := &mockUserer{
		getUser: userWithRoles(
			wsRole(1, 5, authorization_model.RoleWorkspaceAdmin),
		),
	}

	svc := newSvc(config.Config{}, &mockWorkspaceStore{}, &mockK8s{}, userer)
	err := svc.RemoveUserRoleInWorkspace(context.Background(), 1, 42, 5, authorization_model.RoleWorkspaceDataManager)

	require.Error(t, err)
	assert.Empty(t, userer.removedRoleIDs)
}

// ---------------------------------------------------------------------------
// AddUserRoleInWorkspace
// ---------------------------------------------------------------------------

func TestAddUserRoleInWorkspace_AssignsNewRole(t *testing.T) {
	userer := &mockUserer{
		getUser: userWithRoles(),
	}

	svc := newSvc(config.Config{}, &mockWorkspaceStore{}, &mockK8s{}, userer)
	err := svc.AddUserRoleInWorkspace(context.Background(), 1, 42, wsRole(0, 5, authorization_model.RoleWorkspaceAdmin))

	require.NoError(t, err)
	require.Len(t, userer.capturedRoles, 1)
	assert.Equal(t, authorization_model.RoleWorkspaceAdmin, userer.capturedRoles[0].Role.Name)
	assert.Equal(t, "5", userer.capturedRoles[0].Context["workspace"])
}

func TestAddUserRoleInWorkspace_AllowsMultipleDistinctRolesInSameWorkspace(t *testing.T) {
	userer := &mockUserer{
		getUser: userWithRoles(
			wsRole(1, 5, authorization_model.RoleWorkspaceDataManager),
		),
	}

	svc := newSvc(config.Config{}, &mockWorkspaceStore{}, &mockK8s{}, userer)
	err := svc.AddUserRoleInWorkspace(context.Background(), 1, 42, wsRole(0, 5, authorization_model.RoleWorkspaceAdmin))

	require.NoError(t, err)
	require.Len(t, userer.capturedRoles, 1)
	assert.Equal(t, authorization_model.RoleWorkspaceAdmin, userer.capturedRoles[0].Role.Name)
}

func TestAddUserRoleInWorkspace_RejectsDuplicateRole(t *testing.T) {
	userer := &mockUserer{
		getUser: userWithRoles(
			wsRole(1, 5, authorization_model.RoleWorkspaceAdmin),
		),
	}

	svc := newSvc(config.Config{}, &mockWorkspaceStore{}, &mockK8s{}, userer)
	err := svc.AddUserRoleInWorkspace(context.Background(), 1, 42, wsRole(0, 5, authorization_model.RoleWorkspaceAdmin))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "already has role")
	assert.Empty(t, userer.capturedRoles)
}

func TestAddUserRoleInWorkspace_SameRoleDifferentWorkspaceAllowed(t *testing.T) {
	userer := &mockUserer{
		getUser: userWithRoles(
			wsRole(1, 5, authorization_model.RoleWorkspaceAdmin),
		),
	}

	svc := newSvc(config.Config{}, &mockWorkspaceStore{}, &mockK8s{}, userer)
	err := svc.AddUserRoleInWorkspace(context.Background(), 1, 42, wsRole(0, 9, authorization_model.RoleWorkspaceAdmin))

	require.NoError(t, err)
	require.Len(t, userer.capturedRoles, 1)
	assert.Equal(t, "9", userer.capturedRoles[0].Context["workspace"])
}

// ---------------------------------------------------------------------------
// RemoveUserFromWorkspace
// ---------------------------------------------------------------------------

func TestRemoveUserFromWorkspace_RemovesAllRolesInWorkspaceOnly(t *testing.T) {
	userer := &mockUserer{
		getUser: userWithRoles(
			wsRole(1, 5, authorization_model.RoleWorkspaceAdmin),
			wbRole(2, 5, 7, authorization_model.RoleWorkbenchAdmin),
			wsRole(3, 9, authorization_model.RoleWorkspaceAdmin),
		),
	}

	svc := newSvc(config.Config{}, &mockWorkspaceStore{}, &mockK8s{}, userer)
	err := svc.RemoveUserFromWorkspace(context.Background(), 1, 42, 5)

	require.NoError(t, err)
	assert.ElementsMatch(t, []uint64{1, 2}, userer.removedRoleIDs)
}

func TestRemoveUserFromWorkspace_NoRolesIsNoop(t *testing.T) {
	userer := &mockUserer{
		getUser: userWithRoles(
			wsRole(3, 9, authorization_model.RoleWorkspaceAdmin),
		),
	}

	svc := newSvc(config.Config{}, &mockWorkspaceStore{}, &mockK8s{}, userer)
	err := svc.RemoveUserFromWorkspace(context.Background(), 1, 42, 5)

	require.NoError(t, err)
	assert.Empty(t, userer.removedRoleIDs)
}
