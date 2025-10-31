package service

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/pkg/devstore/model"
)

type Devstorer interface {
	ListEntries(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64) ([]*model.DevstoreEntry, error)
	GetEntry(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64, key string) (*model.DevstoreEntry, error)
	PutEntry(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64, key string, value string) (*model.DevstoreEntry, error)
	DeleteEntry(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64, key string) error
}

type DevstoreStore interface {
	ListEntries(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64) ([]*model.DevstoreEntry, error)
	GetEntry(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64, key string) (*model.DevstoreEntry, error)
	PutEntry(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64, key string, value string) (*model.DevstoreEntry, error)
	DeleteEntry(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64, key string) error
}

type DevstoreService struct {
	store DevstoreStore
}

func NewDevstoreService(config config.Config, store DevstoreStore) *DevstoreService {
	return &DevstoreService{
		store: store,
	}
}

func (s DevstoreService) ListEntries(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64) ([]*model.DevstoreEntry, error) {
	return s.store.ListEntries(ctx, tenantID, scope, scopeID)
}

func (s DevstoreService) GetEntry(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64, key string) (*model.DevstoreEntry, error) {
	return s.store.GetEntry(ctx, tenantID, scope, scopeID, key)
}

func (s DevstoreService) PutEntry(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64, key string, value string) (*model.DevstoreEntry, error) {
	return s.store.PutEntry(ctx, tenantID, scope, scopeID, key, value)
}

func (s DevstoreService) DeleteEntry(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64, key string) error {
	return s.store.DeleteEntry(ctx, tenantID, scope, scopeID, key)
}
