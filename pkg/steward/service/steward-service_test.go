//go:build unit

package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	tenant_model "github.com/CHORUS-TRE/chorus-backend/pkg/tenant/model"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	user_service "github.com/CHORUS-TRE/chorus-backend/pkg/user/service"
)

// --- fakes ---

type fakeTenanter struct {
	getTenantFn    func(ctx context.Context, tenantID uint64) (*tenant_model.Tenant, error)
	createTenantFn func(ctx context.Context, name string) (*tenant_model.Tenant, error)
}

func (f *fakeTenanter) GetTenant(ctx context.Context, tenantID uint64) (*tenant_model.Tenant, error) {
	return f.getTenantFn(ctx, tenantID)
}

func (f *fakeTenanter) CreateTenant(ctx context.Context, name string) (*tenant_model.Tenant, error) {
	return f.createTenantFn(ctx, name)
}

type fakeUserer struct {
	getUserFn    func(ctx context.Context, req user_service.GetUserReq) (*user_model.User, error)
	createUserFn func(ctx context.Context, req user_service.CreateUserReq) (*user_model.User, error)
	createRoleFn func(ctx context.Context, role string) error
	getRolesFn   func(ctx context.Context) ([]*user_model.Role, error)
}

func (f *fakeUserer) GetUser(ctx context.Context, req user_service.GetUserReq) (*user_model.User, error) {
	return f.getUserFn(ctx, req)
}

func (f *fakeUserer) CreateUser(ctx context.Context, req user_service.CreateUserReq) (*user_model.User, error) {
	return f.createUserFn(ctx, req)
}

func (f *fakeUserer) CreateRole(ctx context.Context, role string) error {
	return f.createRoleFn(ctx, role)
}

func (f *fakeUserer) GetRoles(ctx context.Context) ([]*user_model.Role, error) {
	return f.getRolesFn(ctx)
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
		getTenantFn:    func(_ context.Context, _ uint64) (*tenant_model.Tenant, error) { return &tenant_model.Tenant{}, nil },
		createTenantFn: func(_ context.Context, _ string) (*tenant_model.Tenant, error) { return &tenant_model.Tenant{}, nil },
	}
}

func tenantNotFoundThenOK() *fakeTenanter {
	calls := 0
	return &fakeTenanter{
		getTenantFn: func(_ context.Context, _ uint64) (*tenant_model.Tenant, error) {
			if calls == 0 {
				calls++
				return nil, sql.ErrNoRows
			}
			return &tenant_model.Tenant{}, nil
		},
		createTenantFn: func(_ context.Context, _ string) (*tenant_model.Tenant, error) { return &tenant_model.Tenant{}, nil },
	}
}

func alwaysOKUserer() *fakeUserer {
	return &fakeUserer{
		getUserFn:    func(_ context.Context, _ user_service.GetUserReq) (*user_model.User, error) { return &user_model.User{}, nil },
		createUserFn: func(_ context.Context, _ user_service.CreateUserReq) (*user_model.User, error) { return &user_model.User{}, nil },
		createRoleFn: func(_ context.Context, _ string) error { return nil },
		getRolesFn:   func(_ context.Context) ([]*user_model.Role, error) { return nil, nil },
	}
}

// --- tests ---

func TestNewStewardService_NoCredentials_Skips(t *testing.T) {
	conf := stewardConf("", "")
	tenanter := &fakeTenanter{
		getTenantFn:    func(_ context.Context, _ uint64) (*tenant_model.Tenant, error) { t.Fatal("unexpected call"); return nil, nil },
		createTenantFn: func(_ context.Context, _ string) (*tenant_model.Tenant, error) { t.Fatal("unexpected call"); return nil, nil },
	}
	userer := &fakeUserer{
		getUserFn:    func(_ context.Context, _ user_service.GetUserReq) (*user_model.User, error) { t.Fatal("unexpected call"); return nil, nil },
		createUserFn: func(_ context.Context, _ user_service.CreateUserReq) (*user_model.User, error) { t.Fatal("unexpected call"); return nil, nil },
		createRoleFn: func(_ context.Context, _ string) error { t.Fatal("unexpected call"); return nil },
		getRolesFn:   func(_ context.Context) ([]*user_model.Role, error) { t.Fatal("unexpected call"); return nil, nil },
	}

	_, err := NewStewardService(conf, tenanter, userer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewStewardService_UsernameOnly_Skips(t *testing.T) {
	conf := stewardConf("chorus", "")
	tenanter := &fakeTenanter{
		getTenantFn:    func(_ context.Context, _ uint64) (*tenant_model.Tenant, error) { t.Fatal("unexpected call"); return nil, nil },
		createTenantFn: func(_ context.Context, _ string) (*tenant_model.Tenant, error) { t.Fatal("unexpected call"); return nil, nil },
	}
	userer := &fakeUserer{
		getUserFn:    func(_ context.Context, _ user_service.GetUserReq) (*user_model.User, error) { t.Fatal("unexpected call"); return nil, nil },
		createUserFn: func(_ context.Context, _ user_service.CreateUserReq) (*user_model.User, error) { t.Fatal("unexpected call"); return nil, nil },
		createRoleFn: func(_ context.Context, _ string) error { t.Fatal("unexpected call"); return nil },
		getRolesFn:   func(_ context.Context) ([]*user_model.Role, error) { t.Fatal("unexpected call"); return nil, nil },
	}

	_, err := NewStewardService(conf, tenanter, userer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewStewardService_TenantAndUserAlreadyExist(t *testing.T) {
	conf := stewardConf("chorus", "password")

	_, err := NewStewardService(conf, alwaysOKTenanter(), alwaysOKUserer())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewStewardService_TenantNotFound_Creates(t *testing.T) {
	conf := stewardConf("chorus", "password")

	tenanter := tenantNotFoundThenOK()
	userer := alwaysOKUserer()

	_, err := NewStewardService(conf, tenanter, userer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewStewardService_UserNotFound_Creates(t *testing.T) {
	conf := stewardConf("chorus", "password")

	var capturedReq user_service.CreateUserReq
	userer := &fakeUserer{
		getUserFn: func(_ context.Context, _ user_service.GetUserReq) (*user_model.User, error) {
			return nil, sql.ErrNoRows
		},
		createUserFn: func(_ context.Context, req user_service.CreateUserReq) (*user_model.User, error) {
			capturedReq = req
			return &user_model.User{}, nil
		},
		createRoleFn: func(_ context.Context, _ string) error { return nil },
		getRolesFn:   func(_ context.Context) ([]*user_model.Role, error) { return nil, nil },
	}

	_, err := NewStewardService(conf, alwaysOKTenanter(), userer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedReq.User == nil {
		t.Fatal("expected CreateUser to be called")
	}
	if capturedReq.User.ID != defaultUserID {
		t.Errorf("expected user ID %d, got %d", defaultUserID, capturedReq.User.ID)
	}
	if capturedReq.User.Username != "chorus" {
		t.Errorf("expected username %q, got %q", "chorus", capturedReq.User.Username)
	}
	if capturedReq.User.Password != "password" {
		t.Errorf("expected password %q, got %q", "password", capturedReq.User.Password)
	}
	if len(capturedReq.User.Roles) != len(defaultBootstrapRoles) {
		t.Errorf("expected %d roles, got %d", len(defaultBootstrapRoles), len(capturedReq.User.Roles))
	}
}

func TestNewStewardService_GetTenantError_ReturnsError(t *testing.T) {
	conf := stewardConf("chorus", "password")

	tenanter := &fakeTenanter{
		getTenantFn:    func(_ context.Context, _ uint64) (*tenant_model.Tenant, error) { return nil, errors.New("db error") },
		createTenantFn: func(_ context.Context, _ string) (*tenant_model.Tenant, error) { return &tenant_model.Tenant{}, nil },
	}

	_, err := NewStewardService(conf, tenanter, alwaysOKUserer())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewStewardService_GetUserError_ReturnsError(t *testing.T) {
	conf := stewardConf("chorus", "password")

	userer := &fakeUserer{
		getUserFn:    func(_ context.Context, _ user_service.GetUserReq) (*user_model.User, error) { return nil, errors.New("db error") },
		createUserFn: func(_ context.Context, _ user_service.CreateUserReq) (*user_model.User, error) { return &user_model.User{}, nil },
		createRoleFn: func(_ context.Context, _ string) error { return nil },
		getRolesFn:   func(_ context.Context) ([]*user_model.Role, error) { return nil, nil },
	}

	_, err := NewStewardService(conf, alwaysOKTenanter(), userer)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewStewardService_CreateUserError_ReturnsError(t *testing.T) {
	conf := stewardConf("chorus", "password")

	userer := &fakeUserer{
		getUserFn:    func(_ context.Context, _ user_service.GetUserReq) (*user_model.User, error) { return nil, sql.ErrNoRows },
		createUserFn: func(_ context.Context, _ user_service.CreateUserReq) (*user_model.User, error) { return nil, errors.New("insert failed") },
		createRoleFn: func(_ context.Context, _ string) error { return nil },
		getRolesFn:   func(_ context.Context) ([]*user_model.Role, error) { return nil, nil },
	}

	_, err := NewStewardService(conf, alwaysOKTenanter(), userer)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestNewStewardService_TenantNotFound_UsesConfigName(t *testing.T) {
	conf := stewardConf("chorus", "password")
	conf.Services.Steward.Tenant.Name = "my-tenant"

	var capturedName string
	tenanter := &fakeTenanter{
		getTenantFn: func(_ context.Context, _ uint64) (*tenant_model.Tenant, error) { return nil, sql.ErrNoRows },
		createTenantFn: func(_ context.Context, name string) (*tenant_model.Tenant, error) {
			capturedName = name
			return &tenant_model.Tenant{ID: 1, Name: name}, nil
		},
	}

	_, err := NewStewardService(conf, tenanter, alwaysOKUserer())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedName != "my-tenant" {
		t.Errorf("expected tenant name %q, got %q", "my-tenant", capturedName)
	}
}

func TestInitializeNewTenant_ReturnsTenant(t *testing.T) {
	conf := stewardConf("chorus", "password")

	want := &tenant_model.Tenant{ID: 42, Name: "acme"}
	tenanter := &fakeTenanter{
		getTenantFn:    func(_ context.Context, _ uint64) (*tenant_model.Tenant, error) { return &tenant_model.Tenant{}, nil },
		createTenantFn: func(_ context.Context, _ string) (*tenant_model.Tenant, error) { return want, nil },
	}

	svc, err := NewStewardService(conf, tenanter, alwaysOKUserer())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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
		getUserFn:    func(_ context.Context, _ user_service.GetUserReq) (*user_model.User, error) { return &user_model.User{}, nil },
		createUserFn: func(_ context.Context, _ user_service.CreateUserReq) (*user_model.User, error) { return &user_model.User{}, nil },
		createRoleFn: func(_ context.Context, _ string) error { return errors.New("role error") },
		getRolesFn:   func(_ context.Context) ([]*user_model.Role, error) { return nil, nil },
	}

	svc, err := NewStewardService(conf, alwaysOKTenanter(), userer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.InitializeNewTenant(context.Background(), "acme")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestInitializeNewTenant_DuplicateTenant_ReturnsError(t *testing.T) {
	conf := stewardConf("chorus", "password")

	tenanter := &fakeTenanter{
		getTenantFn:    func(_ context.Context, _ uint64) (*tenant_model.Tenant, error) { return &tenant_model.Tenant{}, nil },
		createTenantFn: func(_ context.Context, _ string) (*tenant_model.Tenant, error) { return nil, errors.New("duplicate key") },
	}

	svc, err := NewStewardService(conf, tenanter, alwaysOKUserer())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.InitializeNewTenant(context.Background(), "acme")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
