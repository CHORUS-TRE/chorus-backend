package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	authorization_model "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	tenant_model "github.com/CHORUS-TRE/chorus-backend/pkg/tenant/model"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	user_service "github.com/CHORUS-TRE/chorus-backend/pkg/user/service"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Tenanter interface {
	CreateTenant(ctx context.Context, name string) (*tenant_model.Tenant, error)
	GetTenantByName(ctx context.Context, name string) (*tenant_model.Tenant, error)
}

type Userer interface {
	CreateUser(ctx context.Context, req user_service.CreateUserReq) (*user_model.User, error)
	CreateUserRoles(ctx context.Context, tenantID, userID uint64, roles []user_model.UserRole) error
	CreateRole(ctx context.Context, role string) error
}

type Stewarder interface {
	InitializeNewTenant(ctx context.Context, name string) (*tenant_model.Tenant, error)
}

type StewardService struct {
	conf     config.Config
	tenanter Tenanter
	userer   Userer
}

func NewStewardService(conf config.Config, tenanter Tenanter, userer Userer) (*StewardService, error) {
	stewardService := &StewardService{
		conf:     conf,
		tenanter: tenanter,
		userer:   userer,
	}

	if conf.Services.Steward.Tenant.Name != "" && conf.Services.Steward.User.Username != "" && conf.Services.Steward.User.Password.IsSet() {
		tenantID, err := stewardService.InitializeDefaultTenant(context.Background())
		if err != nil {
			return nil, fmt.Errorf("failed to initialize default tenant: %w", err)
		}
		if err := stewardService.InitializeDefaultUser(context.Background(), tenantID); err != nil {
			return nil, fmt.Errorf("failed to initialize default user: %w", err)
		}
	}

	return stewardService, nil
}

func (s *StewardService) InitializeDefaultTenant(ctx context.Context) (uint64, error) {
	if err := s.createDefaultRoles(ctx); err != nil {
		return 0, fmt.Errorf("unable to create default roles: %w", err)
	}

	tenant, err := s.createTenant(ctx, s.conf.Services.Steward.Tenant.Name)
	if err != nil {
		if !strings.Contains(err.Error(), "duplicate key") {
			return 0, fmt.Errorf("unable to initialize default tenant: %w", err)
		}
		existing, err := s.tenanter.GetTenantByName(ctx, s.conf.Services.Steward.Tenant.Name)
		if err != nil {
			return 0, fmt.Errorf("unable to get existing default tenant %q: %w", s.conf.Services.Steward.Tenant.Name, err)
		}
		return existing.ID, nil
	}

	return tenant.ID, nil
}

func (s *StewardService) InitializeDefaultUser(ctx context.Context, tenantID uint64) error {
	user, err := s.userer.CreateUser(ctx, user_service.CreateUserReq{
		TenantID: tenantID,
		User: &user_service.UserReq{
			FirstName: cases.Title(language.English).String(s.conf.Services.Steward.User.Username),
			LastName:  "Default",
			Username:  s.conf.Services.Steward.User.Username,
			Source:    "internal",
			Password:  s.conf.Services.Steward.User.Password.PlainText(),
			Status:    user_model.UserActive,
		},
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil
		}
		return fmt.Errorf("unable to initialize default user: %w", err)
	}

	if err := s.userer.CreateUserRoles(ctx, tenantID, user.ID, bootstrapRolesFor(user.ID)); err != nil {
		return fmt.Errorf("unable to assign bootstrap roles to default user %v: %w", user.ID, err)
	}

	return nil
}

func bootstrapRolesFor(userID uint64) []user_model.UserRole {
	return []user_model.UserRole{
		{Role: authorization_model.NewRole(authorization_model.RoleAuthenticated, authorization_model.WithUser(userID))},
		{Role: authorization_model.NewRole(authorization_model.RolePlatformSettingsManager, authorization_model.WithUser(userID))},
		{Role: authorization_model.NewRole(authorization_model.RolePlateformUserManager, authorization_model.WithUser(userID))},
		{Role: authorization_model.NewRole(authorization_model.RoleAppStoreAdmin, authorization_model.WithUser(userID))},
	}
}

func (s *StewardService) InitializeNewTenant(ctx context.Context, name string) (*tenant_model.Tenant, error) {
	if err := s.createDefaultRoles(ctx); err != nil {
		return nil, fmt.Errorf("unable to create default roles: %w", err)
	}

	tenant, err := s.createTenant(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("unable to create tenant %q: %w", name, err)
	}

	return tenant, nil
}

func (s *StewardService) createDefaultRoles(ctx context.Context) error {
	allRoles := authorization_model.GetAllRoles()

	for _, r := range allRoles {
		if err := s.userer.CreateRole(ctx, r.String()); err != nil {
			return fmt.Errorf("unable to create '%v' role: %w", r, err)
		}
	}

	return nil
}

func (s *StewardService) createTenant(ctx context.Context, name string) (*tenant_model.Tenant, error) {
	tenant, err := s.tenanter.CreateTenant(ctx, name)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return nil, fmt.Errorf("tenant %q already exists: %w", name, err)
		}
		return nil, err
	}
	return tenant, nil
}
