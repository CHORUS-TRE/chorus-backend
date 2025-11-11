package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/cache"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace-file/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace-file/service"

	"github.com/coocood/freecache"
)

const (
	workspaceCacheSize     = 100 * 1024 * 1024 // Max 100MiB stored in memory
	defaultCacheExpiration = 1
	longCacheExpiration    = 60
)

func WorkspaceCaching(log *logger.ContextLogger) func(service.WorkspaceFiler) *Caching {
	return func(next service.WorkspaceFiler) *Caching {
		return &Caching{
			cache: cache.NewCache(freecache.NewCache(workspaceCacheSize), log),
			next:  next,
		}
	}
}

type Caching struct {
	cache *cache.Cache
	next  service.WorkspaceFiler
}

func (c *Caching) GetWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) (*model.WorkspaceFile, error) {
	return c.next.GetWorkspaceFile(ctx, workspaceID, filePath)
}

func (c *Caching) ListWorkspaceFiles(ctx context.Context, workspaceID uint64, filePath string) ([]*model.WorkspaceFile, error) {
	return c.next.ListWorkspaceFiles(ctx, workspaceID, filePath)
}

func (c *Caching) CreateWorkspaceFile(ctx context.Context, workspaceID uint64, file *model.WorkspaceFile) (*model.WorkspaceFile, error) {
	return c.next.CreateWorkspaceFile(ctx, workspaceID, file)
}

func (c *Caching) UpdateWorkspaceFile(ctx context.Context, workspaceID uint64, oldPath string, file *model.WorkspaceFile) (*model.WorkspaceFile, error) {
	return c.next.UpdateWorkspaceFile(ctx, workspaceID, oldPath, file)
}

func (c *Caching) DeleteWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) error {
	return c.next.DeleteWorkspaceFile(ctx, workspaceID, filePath)
}
