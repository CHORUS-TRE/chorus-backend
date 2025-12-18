package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/blockstore"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/cache"
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

func (c *Caching) GetWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) (*blockstore.File, error) {
	return c.next.GetWorkspaceFile(ctx, workspaceID, filePath)
}

func (c *Caching) ListWorkspaceFiles(ctx context.Context, workspaceID uint64, filePath string) ([]*blockstore.File, error) {
	return c.next.ListWorkspaceFiles(ctx, workspaceID, filePath)
}

func (c *Caching) CreateWorkspaceFile(ctx context.Context, workspaceID uint64, file *blockstore.File) (*blockstore.File, error) {
	return c.next.CreateWorkspaceFile(ctx, workspaceID, file)
}

func (c *Caching) UpdateWorkspaceFile(ctx context.Context, workspaceID uint64, oldPath string, file *blockstore.File) (*blockstore.File, error) {
	return c.next.UpdateWorkspaceFile(ctx, workspaceID, oldPath, file)
}

func (c *Caching) DeleteWorkspaceFile(ctx context.Context, workspaceID uint64, filePath string) error {
	return c.next.DeleteWorkspaceFile(ctx, workspaceID, filePath)
}

func (c *Caching) InitiateWorkspaceFileUpload(ctx context.Context, workspaceID uint64, filePath string, file *blockstore.File) (*blockstore.FileUploadInfo, error) {
	return c.next.InitiateWorkspaceFileUpload(ctx, workspaceID, filePath, file)
}

func (c *Caching) UploadWorkspaceFilePart(ctx context.Context, workspaceID uint64, filePath string, uploadID string, part *blockstore.FilePart) (*blockstore.FilePart, error) {
	return c.next.UploadWorkspaceFilePart(ctx, workspaceID, filePath, uploadID, part)
}

func (c *Caching) CompleteWorkspaceFileUpload(ctx context.Context, workspaceID uint64, filePath string, uploadID string, parts []*blockstore.FilePart) (*blockstore.File, error) {
	return c.next.CompleteWorkspaceFileUpload(ctx, workspaceID, filePath, uploadID, parts)
}

func (c *Caching) AbortWorkspaceFileUpload(ctx context.Context, workspaceID uint64, filePath string, uploadID string) error {
	return c.next.AbortWorkspaceFileUpload(ctx, workspaceID, filePath, uploadID)
}
