package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"

	cerr "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	common "github.com/CHORUS-TRE/chorus-backend/pkg/common/storage"
	"github.com/CHORUS-TRE/chorus-backend/pkg/tenant/model"
)

type TenantStorage struct {
	db *sqlx.DB
}

func NewTenantStorage(db *sqlx.DB) *TenantStorage {
	return &TenantStorage{db: db}
}

func (s *TenantStorage) GetTenantByName(ctx context.Context, name string) (*model.Tenant, error) {
	const q = `SELECT * FROM tenants WHERE name = $1`
	t := &model.Tenant{}
	if err := s.db.GetContext(ctx, t, q, name); err != nil {
		return nil, fmt.Errorf("unable to get tenant by name: %w", err)
	}
	return t, nil
}

func (s *TenantStorage) CreateTenant(ctx context.Context, name string) (*model.Tenant, error) {
	const q = `
		INSERT INTO tenants(name, createdat, updatedat) VALUES($1, NOW(), NOW())
		RETURNING id, name, createdat, updatedat;
	`
	t := &model.Tenant{}
	if err := s.db.GetContext(ctx, t, q, name); err != nil {
		if common.IsDuplicateKey(err) {
			return nil, cerr.ErrDuplicateKey
		}
		return nil, fmt.Errorf("unable to create tenant: %w", err)
	}
	return t, nil
}
