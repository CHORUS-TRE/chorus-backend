package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	authorization_model "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	tenant_model "github.com/CHORUS-TRE/chorus-backend/pkg/tenant/model"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	user_service "github.com/CHORUS-TRE/chorus-backend/pkg/user/service"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
	workspace_model "github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
)

type Tenanter interface {
	CreateTenant(ctx context.Context, tenantID uint64, name string) error
	GetTenant(ctx context.Context, tenantID uint64) (*tenant_model.Tenant, error)
}

type Userer interface {
	CreateUser(ctx context.Context, req user_service.CreateUserReq) (*user_model.User, error)
	GetUser(ctx context.Context, req user_service.GetUserReq) (*user_model.User, error)
	CreateRole(ctx context.Context, role string) error
	GetRoles(ctx context.Context) ([]*user_model.Role, error)
}

type Workspaceer interface {
	ListWorkspaces(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, filter workspace_model.WorkspaceFilter) ([]*model.Workspace, *common_model.PaginationResult, error)
	CreateWorkspace(ctx context.Context, workspace *workspace_model.Workspace) (*workspace_model.Workspace, error)
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

func NewStewardService(conf config.Config, tenanter Tenanter, userer Userer, workspaceer Workspaceer) (*StewardService, error) {
	stewardService := &StewardService{
		conf:        conf,
		tenanter:    tenanter,
		userer:      userer,
		workspaceer: workspaceer,
	}

	if conf.Services.Steward.InitTenant.Enabled {
		// Initialize default tenant if it does not exist
		if err := stewardService.InitializeDefaultTenant(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to initialize default tenant: %w", err)
		}

		if conf.Services.Steward.InitUser.Enabled {
			// Create new tenant user with specified roles
			if err := stewardService.InitializeDefaultUser(context.Background()); err != nil {
				return nil, fmt.Errorf("failed to initialize default user: %w", err)
			}

			if conf.Services.Steward.InitWorkspace.Enabled {
				// Create new tenant workspace
				if err := stewardService.InitializeDefaultWorkspace(context.Background()); err != nil {
					return nil, fmt.Errorf("failed to initialize default workspace: %w", err)
				}
			}
		}
	}

	return stewardService, nil
}

func (s *StewardService) InitializeDefaultTenant(ctx context.Context) error {
	_, err := s.tenanter.GetTenant(ctx, s.conf.Services.Steward.InitTenant.TenantID)
	if err == nil {
		logger.TechLog.Info(ctx, "default tenant already exists")
		return nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("unable to get default tenant %v: %w", s.conf.Services.Steward.InitTenant.TenantID, err)
	}

	// Create default tenant
	initErr := s.InitializeNewTenant(ctx, s.conf.Services.Steward.InitTenant.TenantID)
	if initErr != nil {
		return fmt.Errorf("unable to initialize default tenant %v: %w", s.conf.Services.Steward.InitTenant.TenantID, initErr)
	}

	logger.TechLog.Info(ctx, "default tenant successfully initialized")
	return nil
}

func (s *StewardService) InitializeDefaultUser(ctx context.Context) error {
	_, err := s.userer.GetUser(ctx, user_service.GetUserReq{TenantID: s.conf.Services.Steward.InitTenant.TenantID, ID: s.conf.Services.Steward.InitUser.UserID})
	if err == nil {
		logger.TechLog.Info(ctx, "default user already exists")
		return nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("unable to get default user %v: %w", s.conf.Services.Steward.InitUser.UserID, err)
	}

	roles := make([]user_model.UserRole, len(s.conf.Services.Steward.InitUser.Roles))
	for i, r := range s.conf.Services.Steward.InitUser.Roles {
		role, err := authorization_model.ToRole(r.Name, r.Context)
		if err != nil {
			return fmt.Errorf("unable to convert role %v: %w", r, err)
		}
		roles[i] = user_model.UserRole{Role: role}
	}

	// Create default user
	_, createErr := s.userer.CreateUser(ctx, user_service.CreateUserReq{
		TenantID: s.conf.Services.Steward.InitTenant.TenantID,
		User: &user_service.UserReq{
			ID:        s.conf.Services.Steward.InitUser.UserID,
			FirstName: s.conf.Services.Steward.InitUser.Username,
			LastName:  "default",
			Username:  s.conf.Services.Steward.InitUser.Username,
			Source:    "internal",
			Password:  s.conf.Services.Steward.InitUser.Password.PlainText(),
			Status:    user_model.UserActive,
			Roles:     roles,
		},
	})
	if createErr != nil {
		return fmt.Errorf("unable to initialize default user %v: %w", s.conf.Services.Steward.InitUser.UserID, createErr)
	}

	logger.TechLog.Info(ctx, "default user successfully initialized")
	return nil
}

func (s *StewardService) InitializeDefaultWorkspace(ctx context.Context) error {
	workspaces, _, err := s.workspaceer.ListWorkspaces(ctx, s.conf.Services.Steward.InitTenant.TenantID, &common_model.Pagination{}, workspace_model.WorkspaceFilter{})
	if err == nil {
		for _, workspace := range workspaces {
			if workspace.UserID == s.conf.Services.Steward.InitUser.UserID && workspace.Name == s.conf.Services.Steward.InitWorkspace.Name {
				logger.TechLog.Info(ctx, "default workspace already exists")
				return nil
			}
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("unable to get default workspace %v: %w", s.conf.Services.Steward.InitWorkspace.WorkspaceID, err)
	}

	// Create default workspace
	_, createErr := s.workspaceer.CreateWorkspace(ctx, &workspace_model.Workspace{
		ID:          s.conf.Services.Steward.InitWorkspace.WorkspaceID,
		UserID:      s.conf.Services.Steward.InitUser.UserID,
		TenantID:    s.conf.Services.Steward.InitTenant.TenantID,
		Name:        s.conf.Services.Steward.InitWorkspace.Name,
		ShortName:   fmt.Sprintf("ws-%d", s.conf.Services.Steward.InitWorkspace.WorkspaceID),
		Description: fmt.Sprintf("Default workspace for user %v", s.conf.Services.Steward.InitUser.Username),
		Status:      workspace_model.WorkspaceActive,
	})
	if createErr != nil {
		return fmt.Errorf("unable to create default workspace %v: %w", s.conf.Services.Steward.InitWorkspace.WorkspaceID, createErr)
	}

	logger.TechLog.Info(ctx, "default workspace successfully initialized")
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
	allRoles := authorization_model.GetAllRoles()

	for _, r := range allRoles {
		if err := s.userer.CreateRole(ctx, r.String()); err != nil {
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
