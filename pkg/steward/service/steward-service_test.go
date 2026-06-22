//go:build unit

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/lib/pq"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	tenant_model "github.com/CHORUS-TRE/chorus-backend/pkg/tenant/model"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	user_service "github.com/CHORUS-TRE/chorus-backend/pkg/user/service"
)

// --- fakes ---

type fakeTenanter struct {
	createTenantFn    func(ctx context.Context, name string) (*tenant_model.Tenant, error)
	getTenantByNameFn func(ctx context.Context, name string) (*tenant_model.Tenant, error)
}

func (f *fakeTenanter) CreateTenant(ctx context.Context, name string) (*tenant_model.Tenant, error) {
	return f.createTenantFn(ctx, name)
}

func (f *fakeTenanter) GetTenantByName(ctx context.Context, name string) (*tenant_model.Tenant, error) {
	return f.getTenantByNameFn(ctx, name)
}

type fakeUserer struct {
	createUserFn      func(ctx context.Context, req user_service.CreateUserReq) (*user_model.User, error)
	createUserRolesFn func(ctx context.Context, tenantID, userID uint64, roles []user_model.UserRole) error
	createRoleFn      func(ctx context.Context, role string) error
}

func (f *fakeUserer) CreateUser(ctx context.Context, req user_service.CreateUserReq) (*user_model.User, error) {
	return f.createUserFn(ctx, req)
}

func (f *fakeUserer) CreateUserRoles(ctx context.Context, tenantID, userID uint64, roles []user_model.UserRole) error {
	return f.createUserRolesFn(ctx, tenantID, userID, roles)
}

func (f *fakeUserer) CreateRole(ctx context.Context, role string) error {
	return f.createRoleFn(ctx, role)
}

// --- helpers ---

func stewardConf(username, password string) config.Config {
	var conf config.Config
	conf.Daemon.TenantID = 9999999
	conf.Services.Steward.User.Username = username
	conf.Services.Steward.User.Password = config.Sensitive(password)
	conf.Services.Steward.Tenant.Name = "default"
	return conf
}

func alwaysOKTenanter() *fakeTenanter {
	return &fakeTenanter{
		createTenantFn:    func(_ context.Context, _ string) (*tenant_model.Tenant, error) { return &tenant_model.Tenant{}, nil },
		getTenantByNameFn: func(_ context.Context, _ string) (*tenant_model.Tenant, error) { return &tenant_model.Tenant{}, nil },
	}
}

func tenantAlreadyExistsTenanter() *fakeTenanter {
	return &fakeTenanter{
		createTenantFn:    func(_ context.Context, _ string) (*tenant_model.Tenant, error) { return nil, &pq.Error{Code: "23505"} },
		getTenantByNameFn: func(_ context.Context, _ string) (*tenant_model.Tenant, error) { return &tenant_model.Tenant{ID: 1}, nil },
	}
}

func alwaysOKUserer() *fakeUserer {
	return &fakeUserer{
		createUserFn:      func(_ context.Context, _ user_service.CreateUserReq) (*user_model.User, error) { return &user_model.User{}, nil },
		createUserRolesFn: func(_ context.Context, _, _ uint64, _ []user_model.UserRole) error { return nil },
		createRoleFn:      func(_ context.Context, _ string) error { return nil },
	}
}

// --- tests ---

func TestNewStewardService_NoCredentials_Skips(t *testing.T) {
	conf := stewardConf("", "")
	tenanter := &fakeTenanter{
		createTenantFn:    func(_ context.Context, _ string) (*tenant_model.Tenant, error) { t.Fatal("unexpected call"); return nil, nil },
		getTenantByNameFn: func(_ context.Context, _ string) (*tenant_model.Tenant, error) { t.Fatal("unexpected call"); return nil, nil },
	}
	userer := &fakeUserer{
		createUserFn:      func(_ context.Context, _ user_service.CreateUserReq) (*user_model.User, error) { t.Fatal("unexpected call"); return nil, nil },
		createUserRolesFn: func(_ context.Context, _, _ uint64, _ []user_model.UserRole) error { t.Fatal("unexpected call"); return nil },
		createRoleFn:      func(_ context.Context, _ string) error { t.Fatal("unexpected call"); return nil },
	}

	_, err := NewStewardService(conf, tenanter, userer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewStewardService_UsernameOnly_Skips(t *testing.T) {
	conf := stewardConf("chorus", "")
	tenanter := &fakeTenanter{
		createTenantFn:    func(_ context.Context, _ string) (*tenant_model.Tenant, error) { t.Fatal("unexpected call"); return nil, nil },
		getTenantByNameFn: func(_ context.Context, _ string) (*tenant_model.Tenant, error) { t.Fatal("unexpected call"); return nil, nil },
	}
	userer := &fakeUserer{
		createUserFn:      func(_ context.Context, _ user_service.CreateUserReq) (*user_model.User, error) { t.Fatal("unexpected call"); return nil, nil },
		createUserRolesFn: func(_ context.Context, _, _ uint64, _ []user_model.UserRole) error { t.Fatal("unexpected call"); return nil },
		createRoleFn:      func(_ context.Context, _ string) error { t.Fatal("unexpected call"); return nil },
	}

	_, err := NewStewardService(conf, tenanter, userer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewStewardService_PasswordOnly_Skips(t *testing.T) {
	conf := stewardConf("", "password")
	tenanter := &fakeTenanter{
		createTenantFn:    func(_ context.Context, _ string) (*tenant_model.Tenant, error) { t.Fatal("unexpected call"); return nil, nil },
		getTenantByNameFn: func(_ context.Context, _ string) (*tenant_model.Tenant, error) { t.Fatal("unexpected call"); return nil, nil },
	}
	userer := &fakeUserer{
		createUserFn:      func(_ context.Context, _ user_service.CreateUserReq) (*user_model.User, error) { t.Fatal("unexpected call"); return nil, nil },
		createUserRolesFn: func(_ context.Context, _, _ uint64, _ []user_model.UserRole) error { t.Fatal("unexpected call"); return nil },
		createRoleFn:      func(_ context.Context, _ string) error { t.Fatal("unexpected call"); return nil },
	}

	_, err := NewStewardService(conf, tenanter, userer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewStewardService_TenantNameMissing_Skips(t *testing.T) {
	conf := stewardConf("chorus", "password")
	conf.Services.Steward.Tenant.Name = ""
	tenanter := &fakeTenanter{
		createTenantFn:    func(_ context.Context, _ string) (*tenant_model.Tenant, error) { t.Fatal("unexpected call"); return nil, nil },
		getTenantByNameFn: func(_ context.Context, _ string) (*tenant_model.Tenant, error) { t.Fatal("unexpected call"); return nil, nil },
	}
	userer := &fakeUserer{
		createUserFn:      func(_ context.Context, _ user_service.CreateUserReq) (*user_model.User, error) { t.Fatal("unexpected call"); return nil, nil },
		createUserRolesFn: func(_ context.Context, _, _ uint64, _ []user_model.UserRole) error { t.Fatal("unexpected call"); return nil },
		createRoleFn:      func(_ context.Context, _ string) error { t.Fatal("unexpected call"); return nil },
	}

	_, err := NewStewardService(conf, tenanter, userer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewStewardService_FirstRun_CreatesAll(t *testing.T) {
	conf := stewardConf("chorus", "password")

	_, err := NewStewardService(conf, alwaysOKTenanter(), alwaysOKUserer())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewStewardService_TenantAndUserAlreadyExist(t *testing.T) {
	conf := stewardConf("chorus", "password")

	userer := &fakeUserer{
		createUserFn:      func(_ context.Context, _ user_service.CreateUserReq) (*user_model.User, error) { return nil, &pq.Error{Code: "23505"} },
		createUserRolesFn: func(_ context.Context, _, _ uint64, _ []user_model.UserRole) error { t.Fatal("unexpected call"); return nil },
		createRoleFn:      func(_ context.Context, _ string) error { return nil },
	}

	_, err := NewStewardService(conf, tenantAlreadyExistsTenanter(), userer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewStewardService_NewUser_CreatesWithActualID(t *testing.T) {
	conf := stewardConf("chorus", "password")

	var capturedReq user_service.CreateUserReq
	var capturedRolesUserID uint64
	userer := &fakeUserer{
		createUserFn: func(_ context.Context, req user_service.CreateUserReq) (*user_model.User, error) {
			capturedReq = req
			return &user_model.User{ID: 7}, nil
		},
		createUserRolesFn: func(_ context.Context, _, userID uint64, _ []user_model.UserRole) error {
			capturedRolesUserID = userID
			return nil
		},
		createRoleFn: func(_ context.Context, _ string) error { return nil },
	}

	_, err := NewStewardService(conf, alwaysOKTenanter(), userer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedReq.User == nil {
		t.Fatal("expected CreateUser to be called")
	}
	if capturedReq.User.ID != 0 {
		t.Errorf("expected no explicit user ID, got %d", capturedReq.User.ID)
	}
	if capturedReq.User.Username != "chorus" {
		t.Errorf("expected username %q, got %q", "chorus", capturedReq.User.Username)
	}
	if capturedReq.User.Password != "password" {
		t.Errorf("expected password %q, got %q", "password", capturedReq.User.Password)
	}
	if len(capturedReq.User.Roles) != 0 {
		t.Errorf("expected no roles in CreateUser request, got %d", len(capturedReq.User.Roles))
	}
	if capturedRolesUserID != 7 {
		t.Errorf("expected CreateUserRoles called with user ID 7, got %d", capturedRolesUserID)
	}
}

func TestNewStewardService_CreateTenantError_ReturnsError(t *testing.T) {
	conf := stewardConf("chorus", "password")

	tenanter := &fakeTenanter{
		createTenantFn:    func(_ context.Context, _ string) (*tenant_model.Tenant, error) { return nil, errors.New("db error") },
		getTenantByNameFn: func(_ context.Context, _ string) (*tenant_model.Tenant, error) { return &tenant_model.Tenant{}, nil },
	}

	_, err := NewStewardService(conf, tenanter, alwaysOKUserer())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewStewardService_GetTenantByNameError_ReturnsError(t *testing.T) {
	conf := stewardConf("chorus", "password")

	tenanter := &fakeTenanter{
		createTenantFn:    func(_ context.Context, _ string) (*tenant_model.Tenant, error) { return nil, &pq.Error{Code: "23505"} },
		getTenantByNameFn: func(_ context.Context, _ string) (*tenant_model.Tenant, error) { return nil, errors.New("db error") },
	}

	_, err := NewStewardService(conf, tenanter, alwaysOKUserer())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewStewardService_UserExists_Skips(t *testing.T) {
	conf := stewardConf("chorus", "password")

	userer := &fakeUserer{
		createUserFn: func(_ context.Context, _ user_service.CreateUserReq) (*user_model.User, error) {
			return nil, &pq.Error{Code: "23505"}
		},
		createUserRolesFn: func(_ context.Context, _, _ uint64, _ []user_model.UserRole) error {
			t.Fatal("CreateUserRoles should not be called when user already exists")
			return nil
		},
		createRoleFn: func(_ context.Context, _ string) error { return nil },
	}

	_, err := NewStewardService(conf, alwaysOKTenanter(), userer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewStewardService_CreateUserError_ReturnsError(t *testing.T) {
	conf := stewardConf("chorus", "password")

	userer := &fakeUserer{
		createUserFn:      func(_ context.Context, _ user_service.CreateUserReq) (*user_model.User, error) { return nil, errors.New("insert failed") },
		createUserRolesFn: func(_ context.Context, _, _ uint64, _ []user_model.UserRole) error { return nil },
		createRoleFn:      func(_ context.Context, _ string) error { return nil },
	}

	_, err := NewStewardService(conf, alwaysOKTenanter(), userer)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewStewardService_CreateUserRolesError_ReturnsError(t *testing.T) {
	conf := stewardConf("chorus", "password")

	userer := &fakeUserer{
		createUserFn:      func(_ context.Context, _ user_service.CreateUserReq) (*user_model.User, error) { return &user_model.User{ID: 1}, nil },
		createUserRolesFn: func(_ context.Context, _, _ uint64, _ []user_model.UserRole) error { return errors.New("roles failed") },
		createRoleFn:      func(_ context.Context, _ string) error { return nil },
	}

	_, err := NewStewardService(conf, alwaysOKTenanter(), userer)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewStewardService_UsesConfigTenantName(t *testing.T) {
	conf := stewardConf("chorus", "password")
	conf.Services.Steward.Tenant.Name = "my-tenant"

	var capturedName string
	tenanter := &fakeTenanter{
		createTenantFn: func(_ context.Context, name string) (*tenant_model.Tenant, error) {
			capturedName = name
			return &tenant_model.Tenant{ID: 1, Name: name}, nil
		},
		getTenantByNameFn: func(_ context.Context, _ string) (*tenant_model.Tenant, error) { return &tenant_model.Tenant{}, nil },
	}

	_, err := NewStewardService(conf, tenanter, alwaysOKUserer())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedName != "my-tenant" {
		t.Errorf("expected tenant name %q, got %q", "my-tenant", capturedName)
	}
}

func newStewardSvc(conf config.Config, tenanter Tenanter, userer Userer) *StewardService {
	return &StewardService{conf: conf, tenanter: tenanter, userer: userer}
}

func TestInitializeNewTenant_ReturnsTenant(t *testing.T) {
	conf := stewardConf("chorus", "password")

	want := &tenant_model.Tenant{ID: 42, Name: "acme"}
	tenanter := &fakeTenanter{
		createTenantFn:    func(_ context.Context, _ string) (*tenant_model.Tenant, error) { return want, nil },
		getTenantByNameFn: func(_ context.Context, _ string) (*tenant_model.Tenant, error) { return &tenant_model.Tenant{}, nil },
	}

	svc := newStewardSvc(conf, tenanter, alwaysOKUserer())
	got, err := svc.InitializeNewTenant(context.Background(), "acme")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != want.ID || got.Name != want.Name {
		t.Errorf("expected tenant %+v, got %+v", want, got)
	}
}

func TestInitializeNewTenant_CreateRolesError_ReturnsError(t *testing.T) {
	conf := stewardConf("chorus", "password")

	userer := &fakeUserer{
		createUserFn:      func(_ context.Context, _ user_service.CreateUserReq) (*user_model.User, error) { return &user_model.User{}, nil },
		createUserRolesFn: func(_ context.Context, _, _ uint64, _ []user_model.UserRole) error { return nil },
		createRoleFn:      func(_ context.Context, _ string) error { return errors.New("role error") },
	}

	svc := newStewardSvc(conf, alwaysOKTenanter(), userer)
	_, err := svc.InitializeNewTenant(context.Background(), "acme")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestInitializeNewTenant_DuplicateTenant_ReturnsError(t *testing.T) {
	conf := stewardConf("chorus", "password")

	tenanter := &fakeTenanter{
		createTenantFn:    func(_ context.Context, _ string) (*tenant_model.Tenant, error) { return nil, &pq.Error{Code: "23505"} },
		getTenantByNameFn: func(_ context.Context, _ string) (*tenant_model.Tenant, error) { return &tenant_model.Tenant{}, nil },
	}

	svc := newStewardSvc(conf, tenanter, alwaysOKUserer())
	_, err := svc.InitializeNewTenant(context.Background(), "acme")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
