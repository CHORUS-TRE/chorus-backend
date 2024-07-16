package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/cache"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app-instance/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app-instance/service"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"

	"github.com/coocood/freecache"
)

const (
	appInstanceCacheSize   = 100 * 1024 * 1024 // Max 100MiB stored in memory
	defaultCacheExpiration = 5
	longCacheExpiration    = 60
)

func AppInstanceCaching(log *logger.ContextLogger) func(service.AppInstanceer) *Caching {
	return func(next service.AppInstanceer) *Caching {
		return &Caching{
			cache: cache.NewCache(freecache.NewCache(appInstanceCacheSize), log),
			next:  next,
		}
	}
}

type Caching struct {
	cache *cache.Cache
	next  service.AppInstanceer
}

func (c *Caching) ListAppInstances(ctx context.Context, tenantID uint64, pagination common_model.Pagination) (reply []*model.AppInstance, err error) {
	entry := c.cache.NewEntry(cache.WithUint64(tenantID), cache.WithInterface(pagination))
	reply = []*model.AppInstance{}

	if ok := entry.Get(ctx, &reply); !ok {
		reply, err = c.next.ListAppInstances(ctx, tenantID, pagination)
		if err == nil {
			entry.Set(ctx, defaultCacheExpiration, reply)
		}
	}

	return
}

func (c *Caching) GetAppInstance(ctx context.Context, tenantID, appInstanceID uint64) (reply *model.AppInstance, err error) {
	entry := c.cache.NewEntry(cache.WithUint64(tenantID), cache.WithUint64(appInstanceID))
	reply = &model.AppInstance{}

	if ok := entry.Get(ctx, &reply); !ok {
		reply, err = c.next.GetAppInstance(ctx, tenantID, appInstanceID)
		if err == nil {
			entry.Set(ctx, defaultCacheExpiration, reply)
		}
	}

	return
}

func (c *Caching) DeleteAppInstance(ctx context.Context, tenantID, appInstanceID uint64) error {
	return c.next.DeleteAppInstance(ctx, tenantID, appInstanceID)
}

func (c *Caching) UpdateAppInstance(ctx context.Context, appInstance *model.AppInstance) error {
	return c.next.UpdateAppInstance(ctx, appInstance)
}

func (c *Caching) CreateAppInstance(ctx context.Context, appInstance *model.AppInstance) (uint64, error) {
	return c.next.CreateAppInstance(ctx, appInstance)
}
