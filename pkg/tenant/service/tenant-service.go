package service

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/pkg/tenant/model"
)

type Tenanter interface {
	CreateTenant(ctx context.Context, name string) (*model.Tenant, error)
	GetTenantByName(ctx context.Context, name string) (*model.Tenant, error)
}

type TenantStore interface {
	GetTenantByName(ctx context.Context, name string) (*model.Tenant, error)
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
	tenant, err := s.store.CreateTenant(ctx, name)
	if err != nil {
		return nil, err
	}
	return tenant, nil
}

func (s *TenantService) GetTenantByName(ctx context.Context, name string) (*model.Tenant, error) {
	return s.store.GetTenantByName(ctx, name)
}
