package service

import (
	"context"
	"database/sql"
	"errors"
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

const (
	defaultTenantID = uint64(1)
	defaultUserID   = uint64(1)
)

var defaultBootstrapRoles = []user_model.UserRole{
	{Role: authorization_model.NewRole(authorization_model.RoleAuthenticated, authorization_model.WithUser(defaultUserID))},
	{Role: authorization_model.NewRole(authorization_model.RolePlatformSettingsManager, authorization_model.WithUser(defaultUserID))},
	{Role: authorization_model.NewRole(authorization_model.RolePlateformUserManager, authorization_model.WithUser(defaultUserID))},
	{Role: authorization_model.NewRole(authorization_model.RoleAppStoreAdmin, authorization_model.WithUser(defaultUserID))},
}

type Tenanter interface {
	CreateTenant(ctx context.Context, name string) (*tenant_model.Tenant, error)
	GetTenant(ctx context.Context, tenantID uint64) (*tenant_model.Tenant, error)
}

type Userer interface {
	CreateUser(ctx context.Context, req user_service.CreateUserReq) (*user_model.User, error)
	GetUser(ctx context.Context, req user_service.GetUserReq) (*user_model.User, error)
	CreateRole(ctx context.Context, role string) error
	GetRoles(ctx context.Context) ([]*user_model.Role, error)
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

	if conf.Services.Steward.User.Username != "" && conf.Services.Steward.User.Password.IsSet() {
		if err := stewardService.InitializeDefaultTenant(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to initialize default tenant: %w", err)
		}
		if err := stewardService.InitializeDefaultUser(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to initialize default user: %w", err)
		}
	}

	return stewardService, nil
}

func (s *StewardService) InitializeDefaultTenant(ctx context.Context) error {
	_, err := s.tenanter.GetTenant(ctx, defaultTenantID)
	if err == nil {
		return nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("unable to get default tenant %v: %w", defaultTenantID, err)
	}

	if err := s.createDefaultRoles(ctx); err != nil {
		return fmt.Errorf("unable to create default roles: %w", err)
	}

	if _, err := s.createTenant(ctx, s.conf.Services.Steward.Tenant.Name); err != nil {
		return fmt.Errorf("unable to initialize default tenant: %w", err)
	}

	return nil
}

func (s *StewardService) InitializeDefaultUser(ctx context.Context) error {
	_, err := s.userer.GetUser(ctx, user_service.GetUserReq{TenantID: defaultTenantID, ID: defaultUserID})
	if err == nil {
		return nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("unable to get default user %v: %w", defaultUserID, err)
	}

	_, createErr := s.userer.CreateUser(ctx, user_service.CreateUserReq{
		TenantID: defaultTenantID,
		User: &user_service.UserReq{
			ID:        defaultUserID,
			FirstName: cases.Title(language.English).String(s.conf.Services.Steward.User.Username),
			LastName:  "Default",
			Username:  s.conf.Services.Steward.User.Username,
			Source:    "internal",
			Password:  s.conf.Services.Steward.User.Password.PlainText(),
			Status:    user_model.UserActive,
			Roles:     defaultBootstrapRoles,
		},
	})
	if createErr != nil {
		return fmt.Errorf("unable to initialize default user %v: %w", defaultUserID, createErr)
	}

	return nil
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
