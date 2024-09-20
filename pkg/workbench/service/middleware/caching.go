package middleware

import (
	"context"
	"net/http"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/cache"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/service"

	"github.com/coocood/freecache"
)

const (
	workbenchCacheSize     = 100 * 1024 * 1024 // Max 100MiB stored in memory
	defaultCacheExpiration = 5
	longCacheExpiration    = 60
)

func WorkbenchCaching(log *logger.ContextLogger) func(service.Workbencher) *Caching {
	return func(next service.Workbencher) *Caching {
		return &Caching{
			cache: cache.NewCache(freecache.NewCache(workbenchCacheSize), log),
			next:  next,
		}
	}
}

type Caching struct {
	cache *cache.Cache
	next  service.Workbencher
}

func (c *Caching) ListWorkbenchs(ctx context.Context, tenantID uint64, pagination common_model.Pagination) (reply []*model.Workbench, err error) {
	entry := c.cache.NewEntry(cache.WithUint64(tenantID), cache.WithInterface(pagination))
	reply = []*model.Workbench{}

	if ok := entry.Get(ctx, &reply); !ok {
		reply, err = c.next.ListWorkbenchs(ctx, tenantID, pagination)
		if err == nil {
			entry.Set(ctx, defaultCacheExpiration, reply)
		}
	}

	return
}

func (c *Caching) GetWorkbench(ctx context.Context, tenantID, workbenchID uint64) (reply *model.Workbench, err error) {
	entry := c.cache.NewEntry(cache.WithUint64(tenantID), cache.WithUint64(workbenchID))
	reply = &model.Workbench{}

	if ok := entry.Get(ctx, &reply); !ok {
		reply, err = c.next.GetWorkbench(ctx, tenantID, workbenchID)
		if err == nil {
			entry.Set(ctx, defaultCacheExpiration, reply)
		}
	}

	return
}

func (c *Caching) ProxyWorkbench(ctx context.Context, tenantID, workbenchID uint64, w http.ResponseWriter, r *http.Request) error {
	return c.next.ProxyWorkbench(ctx, tenantID, workbenchID, w, r)

}

func (c *Caching) DeleteWorkbench(ctx context.Context, tenantID, workbenchID uint64) error {
	return c.next.DeleteWorkbench(ctx, tenantID, workbenchID)
}

func (c *Caching) UpdateWorkbench(ctx context.Context, workbench *model.Workbench) error {
	return c.next.UpdateWorkbench(ctx, workbench)
}

func (c *Caching) CreateWorkbench(ctx context.Context, workbench *model.Workbench) (uint64, error) {
	return c.next.CreateWorkbench(ctx, workbench)
}
