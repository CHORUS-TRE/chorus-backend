package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/pkg/devstore/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/devstore/service"

	val "github.com/go-playground/validator/v10"
)

type validation struct {
	next     service.Devstorer
	validate *val.Validate
}

func Validation(validate *val.Validate) func(service.Devstorer) service.Devstorer {
	return func(next service.Devstorer) service.Devstorer {
		return &validation{
			next:     next,
			validate: validate,
		}
	}
}

func (v validation) ListEntries(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64) ([]*model.DevstoreEntry, error) {
	return v.next.ListEntries(ctx, tenantID, scope, scopeID)
}

func (v validation) GetEntry(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64, key string) (*model.DevstoreEntry, error) {
	return v.next.GetEntry(ctx, tenantID, scope, scopeID, key)
}

func (v validation) PutEntry(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64, key string, value string) (*model.DevstoreEntry, error) {
	return v.next.PutEntry(ctx, tenantID, scope, scopeID, key, value)
}

func (v validation) DeleteEntry(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64, key string) error {
	return v.next.DeleteEntry(ctx, tenantID, scope, scopeID, key)
}
