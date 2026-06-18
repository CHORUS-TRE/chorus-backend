package service

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/pkg/tenant/model"
)

type Tenanter interface {
	CreateTenant(ctx context.Context, name string) (*model.Tenant, error)
	GetTenant(ctx context.Context, tenantID uint64) (*model.Tenant, error)
}

type TenantStore interface {
	GetTenant(ctx context.Context, tenantID uint64) (*model.Tenant, error)
	CreateTenant(ctx context.Context, name string) (*model.Tenant, error)
}

type TenantService struct {
	store TenantStore
	conf  config.Config
}

func NewTenantService(store TenantStore, conf config.Config) *TenantService {
	return &TenantService{store: store, conf: conf}
}

func (s *TenantService) CreateTenant(ctx context.Context, name string) (*model.Tenant, error) {
	return s.store.CreateTenant(ctx, name)
}

func (s *TenantService) GetTenant(ctx context.Context, tenantID uint64) (*model.Tenant, error) {
	return s.store.GetTenant(ctx, tenantID)
}
