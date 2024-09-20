package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	tenant_model "github.com/CHORUS-TRE/chorus-backend/pkg/tenant/model"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
)

type Tenanter interface {
	CreateTenant(ctx context.Context, tenantID uint64, name string) error
	GetTenant(ctx context.Context, tenantID uint64) (*tenant_model.Tenant, error)
}

type Userer interface {
	CreateRole(ctx context.Context, role string) error
	GetRoles(ctx context.Context) ([]*user_model.Role, error)
}

type Stewarder interface {
	InitializeNewTenant(ctx context.Context, tenantID uint64) error
}

type StewardService struct {
	conf     config.Config
	tenanter Tenanter
	userer   Userer
}

func NewStewardService(conf config.Config, tenanter Tenanter, userer Userer) *StewardService {
	return &StewardService{conf: conf, tenanter: tenanter, userer: userer}
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
