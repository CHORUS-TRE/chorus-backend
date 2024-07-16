package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/cache"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/service"

	"github.com/coocood/freecache"
)

const (
	workspaceCacheSize     = 100 * 1024 * 1024 // Max 100MiB stored in memory
	defaultCacheExpiration = 5
	longCacheExpiration    = 60
)

func WorkspaceCaching(log *logger.ContextLogger) func(service.Workspaceer) *Caching {
	return func(next service.Workspaceer) *Caching {
		return &Caching{
			cache: cache.NewCache(freecache.NewCache(workspaceCacheSize), log),
			next:  next,
		}
	}
}

type Caching struct {
	cache *cache.Cache
	next  service.Workspaceer
}

func (c *Caching) ListWorkspaces(ctx context.Context, tenantID uint64, pagination common_model.Pagination) (reply []*model.Workspace, err error) {
	entry := c.cache.NewEntry(cache.WithUint64(tenantID), cache.WithInterface(pagination))
	reply = []*model.Workspace{}

	if ok := entry.Get(ctx, &reply); !ok {
		reply, err = c.next.ListWorkspaces(ctx, tenantID, pagination)
		if err == nil {
			entry.Set(ctx, defaultCacheExpiration, reply)
		}
	}

	return
}

func (c *Caching) GetWorkspace(ctx context.Context, tenantID, workspaceID uint64) (reply *model.Workspace, err error) {
	entry := c.cache.NewEntry(cache.WithUint64(tenantID), cache.WithUint64(workspaceID))
	reply = &model.Workspace{}

	if ok := entry.Get(ctx, &reply); !ok {
		reply, err = c.next.GetWorkspace(ctx, tenantID, workspaceID)
		if err == nil {
			entry.Set(ctx, defaultCacheExpiration, reply)
		}
	}

	return
}

func (c *Caching) DeleteWorkspace(ctx context.Context, tenantID, workspaceID uint64) error {
	return c.next.DeleteWorkspace(ctx, tenantID, workspaceID)
}

func (c *Caching) UpdateWorkspace(ctx context.Context, workspace *model.Workspace) error {
	return c.next.UpdateWorkspace(ctx, workspace)
}

func (c *Caching) CreateWorkspace(ctx context.Context, workspace *model.Workspace) (uint64, error) {
	return c.next.CreateWorkspace(ctx, workspace)
}
