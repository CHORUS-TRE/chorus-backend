package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/CHORUS-TRE/chorus-backend/pkg/tenant/model"
)

type TenantStorage struct {
	db *sqlx.DB
}

func NewTenantStorage(db *sqlx.DB) *TenantStorage {
	return &TenantStorage{db: db}
}

func (s *TenantStorage) GetTenant(ctx context.Context, tenantID uint64) (*model.Tenant, error) {
	const q = `SELECT * FROM tenants where id = $1`
	t := &model.Tenant{}
	if err := s.db.Get(t, q, tenantID); err != nil {
		return nil, fmt.Errorf("unable to get tenant: %w", err)
	}
	return t, nil
}

func (s *TenantStorage) CreateTenant(ctx context.Context, tenantID uint64, name string) error {
	ins := `
		INSERT INTO tenants(id, name, createdat, updatedat) VALUES($1, $2, $3, $3);
	`
	_, err := s.db.ExecContext(ctx, ins, tenantID, name, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("unable to create tenant: %w", err)
	}

	return nil
}
