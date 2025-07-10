package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	tenant_model "github.com/CHORUS-TRE/chorus-backend/pkg/tenant/model"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	user_service "github.com/CHORUS-TRE/chorus-backend/pkg/user/service"
	workspace_model "github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
)

type Tenanter interface {
	CreateTenant(ctx context.Context, tenantID uint64, name string) error
	GetTenant(ctx context.Context, tenantID uint64) (*tenant_model.Tenant, error)
}

type Userer interface {
	CreateUser(ctx context.Context, req user_service.CreateUserReq) (uint64, error)
	GetUser(ctx context.Context, req user_service.GetUserReq) (*user_model.User, error)
	CreateRole(ctx context.Context, role string) error
	GetRoles(ctx context.Context) ([]*user_model.Role, error)
}

type Workspaceer interface {
	GetWorkspace(ctx context.Context, tenantID, workspaceID uint64) (*workspace_model.Workspace, error)
	CreateWorkspace(ctx context.Context, workspace *workspace_model.Workspace) (uint64, error)
}

type Stewarder interface {
	InitializeNewTenant(ctx context.Context, tenantID uint64) error
}

type StewardService struct {
	conf        config.Config
	tenanter    Tenanter
	userer      Userer
	workspaceer Workspaceer
}

const DEFAULT_TENANT_ID = 1
const DEFAULT_USER_ID = 1
const DEFAULT_WORKSPACE_ID = 1

func NewStewardService(conf config.Config, tenanter Tenanter, userer Userer, worspaceer Workspaceer) *StewardService {
	stewardService := &StewardService{
		conf:        conf,
		tenanter:    tenanter,
		userer:      userer,
		workspaceer: worspaceer,
	}

	if conf.Services.Steward.Tenant.Enabled {
		// Initialize default tenant if it does not exist
		if err := stewardService.InitializeDefaultTenant(context.Background()); err != nil {
			fmt.Println(err)
		}

		if conf.Services.Steward.User.Enabled {
			// Create new tenant user with specified roles
			if err := stewardService.InitializeDefaultUser(context.Background()); err != nil {
				fmt.Println(err)
			}

			if conf.Services.Steward.Workspace.Enabled {
				// Create new tenant workspace
				if err := stewardService.InitializeDefaultWorkspace(context.Background()); err != nil {
					fmt.Println(err)
				}
			}
		}
	}

	return stewardService
}

func (s *StewardService) InitializeDefaultTenant(ctx context.Context) error {
	_, err := s.tenanter.GetTenant(ctx, DEFAULT_TENANT_ID)
	if err == nil {
		fmt.Println("default tenant already exists")
		return nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("unable to get default tenant %v: %w", DEFAULT_TENANT_ID, err)
	}

	// Create default tenant
	initErr := s.InitializeNewTenant(ctx, DEFAULT_TENANT_ID)
	if initErr != nil {
		return fmt.Errorf("unable to initialize default tenant %v: %w", DEFAULT_TENANT_ID, initErr)
	}

	fmt.Println("default tenant successfully initialized")
	return nil
}

func (s *StewardService) InitializeDefaultUser(ctx context.Context) error {
	_, err := s.userer.GetUser(ctx, user_service.GetUserReq{TenantID: DEFAULT_TENANT_ID, ID: DEFAULT_USER_ID})
	if err == nil {
		fmt.Println("default user already exists")
		return nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("unable to get default user %v: %w", DEFAULT_USER_ID, err)
	}

	// Map roles from config to UserRole array
	roles, roleErr := user_model.ToUserRoles(s.conf.Services.Steward.User.Roles)
	if roleErr != nil {
		return fmt.Errorf("unable to map user roles: %w", roleErr)
	}

	// Create default user
	_, createErr := s.userer.CreateUser(ctx, user_service.CreateUserReq{
		TenantID: DEFAULT_TENANT_ID,
		User: &user_service.UserReq{
			ID:        DEFAULT_USER_ID,
			FirstName: s.conf.Services.Steward.User.Username,
			LastName:  "default",
			Username:  s.conf.Services.Steward.User.Username,
			Source:    "internal",
			Password:  s.conf.Services.Steward.User.Password.PlainText(),
			Status:    user_model.UserActive,
			Roles:     roles,
		},
	})
	if createErr != nil {
		return fmt.Errorf("unable to initialize default user %v: %w", DEFAULT_USER_ID, createErr)
	}

	fmt.Println("default user successfully initialized")
	return nil
}

func (s *StewardService) InitializeDefaultWorkspace(ctx context.Context) error {
	_, err := s.workspaceer.GetWorkspace(ctx, DEFAULT_TENANT_ID, DEFAULT_WORKSPACE_ID)
	if err == nil {
		fmt.Println("default workspace already exists")
		return nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("unable to get default workspace %v: %w", DEFAULT_WORKSPACE_ID, err)
	}

	// Create default workspace
	_, createErr := s.workspaceer.CreateWorkspace(ctx, &workspace_model.Workspace{
		ID:          DEFAULT_WORKSPACE_ID,
		UserID:      DEFAULT_USER_ID,
		TenantID:    DEFAULT_TENANT_ID,
		Name:        s.conf.Services.Steward.Workspace.Name,
		ShortName:   fmt.Sprintf("ws-%d", DEFAULT_WORKSPACE_ID),
		Description: fmt.Sprintf("Default workspace for user %v", s.conf.Services.Steward.User.Username),
		Status:      workspace_model.WorkspaceActive,
	})
	if createErr != nil {
		return fmt.Errorf("unable to create default workspace %v: %w", DEFAULT_WORKSPACE_ID, createErr)
	}

	fmt.Println("default workspace successfully initialized")
	return nil
}

func (s *StewardService) InitializeNewTenant(ctx context.Context, tenantID uint64) error {

	if tenantID == s.conf.Daemon.TenantID {
		return fmt.Errorf("tenant %v is reserved for technical users and cannot be initialized manually", tenantID)
	}

	// 1) ensure that default roles exist
	if err := s.createDefaultRoles(ctx); err != nil {
		return fmt.Errorf("unable to create default roles: %w", err)
	}

	// 2) ensure that technical tenant is created with required users
	if err := s.createTechnicalTenant(ctx); err != nil {
		return fmt.Errorf("unable to create technical tenant: %w", err)
	}

	// 3) Create tenant
	if err := s.createTenant(ctx, tenantID); err != nil {
		return fmt.Errorf("unable to create tenant: %v: %w", tenantID, err)
	}

	return nil
}

func (s *StewardService) createDefaultRoles(ctx context.Context) error {

	for _, r := range []string{user_model.RoleAuthenticated.String(), user_model.RoleAdmin.String(), user_model.RoleChorus.String()} {
		if err := s.userer.CreateRole(ctx, r); err != nil {
			return fmt.Errorf("unable to create '%v' role: %w", r, err)
		}
	}

	return nil
}
func (s *StewardService) createTechnicalTenant(ctx context.Context) error {

	err := s.tenanter.CreateTenant(ctx, s.conf.Daemon.TenantID, fmt.Sprintf("CHORUS-TECHNICAL-TENANT-%v", s.conf.Daemon.TenantID))
	if err != nil && !strings.Contains(err.Error(), "duplicate key") {
		return fmt.Errorf("unable to create technical tenant: %v: %w", s.conf.Daemon.TenantID, err)
	}

	return nil
}

func (s *StewardService) createTenant(ctx context.Context, tenantID uint64) error {

	name := fmt.Sprintf("CHORUS-TENANT-%v", tenantID)

	err := s.tenanter.CreateTenant(ctx, tenantID, name)
	if err != nil {

		if strings.Contains(err.Error(), "duplicate key") {
			return fmt.Errorf("tenant %v already exists: %w", tenantID, err)
		}

		return err
	}

	return nil
}
