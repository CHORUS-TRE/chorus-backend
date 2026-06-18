package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/cache"
	tenant_model "github.com/CHORUS-TRE/chorus-backend/pkg/tenant/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/tenant/service"

	"github.com/coocood/freecache"
)

const (
	tenantCacheSize        = 100 * 1024 * 1024 // Max 100MiB stored in memory
	defaultCacheExpiration = 5
)

func TenantCaching(log *logger.ContextLogger) func(service.Tenanter) *Caching {
	return func(next service.Tenanter) *Caching {
		return &Caching{
			cache: cache.NewCache(freecache.NewCache(tenantCacheSize), log),
			next:  next,
		}
	}
}

type Caching struct {
	cache *cache.Cache
	next  service.Tenanter
}

func (c *Caching) CreateTenant(ctx context.Context, name string) (*tenant_model.Tenant, error) {
	return c.next.CreateTenant(ctx, name)
}

func (c *Caching) GetTenantByName(ctx context.Context, name string) (*tenant_model.Tenant, error) {
	return c.next.GetTenantByName(ctx, name)
}
