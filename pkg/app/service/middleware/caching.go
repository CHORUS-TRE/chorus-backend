package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/cache"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app/service"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"

	"github.com/coocood/freecache"
)

const (
	appCacheSize           = 100 * 1024 * 1024 // Max 100MiB stored in memory
	defaultCacheExpiration = 5
	longCacheExpiration    = 60
)

func AppCaching(log *logger.ContextLogger) func(service.Apper) *Caching {
	return func(next service.Apper) *Caching {
		return &Caching{
			cache: cache.NewCache(freecache.NewCache(appCacheSize), log),
			next:  next,
		}
	}
}

type Caching struct {
	cache *cache.Cache
	next  service.Apper
}

func (c *Caching) ListApps(ctx context.Context, tenantID uint64, pagination common_model.Pagination) (reply []*model.App, err error) {
	entry := c.cache.NewEntry(cache.WithUint64(tenantID), cache.WithInterface(pagination))
	reply = []*model.App{}

	if ok := entry.Get(ctx, &reply); !ok {
		reply, err = c.next.ListApps(ctx, tenantID, pagination)
		if err == nil {
			entry.Set(ctx, defaultCacheExpiration, reply)
		}
	}

	return
}

func (c *Caching) GetApp(ctx context.Context, tenantID, appID uint64) (reply *model.App, err error) {
	entry := c.cache.NewEntry(cache.WithUint64(tenantID), cache.WithUint64(appID))
	reply = &model.App{}

	if ok := entry.Get(ctx, &reply); !ok {
		reply, err = c.next.GetApp(ctx, tenantID, appID)
		if err == nil {
			entry.Set(ctx, defaultCacheExpiration, reply)
		}
	}

	return
}

func (c *Caching) DeleteApp(ctx context.Context, tenantID, appID uint64) error {
	return c.next.DeleteApp(ctx, tenantID, appID)
}

func (c *Caching) UpdateApp(ctx context.Context, app *model.App) error {
	return c.next.UpdateApp(ctx, app)
}

func (c *Caching) CreateApp(ctx context.Context, app *model.App) (uint64, error) {
	return c.next.CreateApp(ctx, app)
}
