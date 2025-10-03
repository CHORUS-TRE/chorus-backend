package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/cache"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/service"

	"github.com/coocood/freecache"
)

const (
	workspaceCacheSize     = 100 * 1024 * 1024 // Max 100MiB stored in memory
	defaultCacheExpiration = 1
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

func (c *Caching) ListWorkspaces(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, filter service.WorkspaceFilter) (reply []*model.Workspace, paginationRes *common_model.PaginationResult, err error) {
	entry := c.cache.NewEntry(cache.WithUint64(tenantID), cache.WithInterface(pagination), cache.WithInterface(filter))
	reply = []*model.Workspace{}
	paginationRes = &common_model.PaginationResult{}

	if ok := entry.Get(ctx, &reply, &paginationRes); !ok {
		reply, paginationRes, err = c.next.ListWorkspaces(ctx, tenantID, pagination, filter)
		if err == nil {
			entry.Set(ctx, defaultCacheExpiration, reply, paginationRes)
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

func (c *Caching) UpdateWorkspace(ctx context.Context, workspace *model.Workspace) (*model.Workspace, error) {
	return c.next.UpdateWorkspace(ctx, workspace)
}

func (c *Caching) CreateWorkspace(ctx context.Context, workspace *model.Workspace) (*model.Workspace, error) {
	return c.next.CreateWorkspace(ctx, workspace)
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

func (c *Caching) ManageUserRoleInWorkspace(ctx context.Context, tenantID, userID uint64, role user_model.UserRole) error {
	return c.next.ManageUserRoleInWorkspace(ctx, tenantID, userID, role)
}

func (c *Caching) RemoveUserFromWorkspace(ctx context.Context, tenantID, userID uint64, workspaceID uint64) error {
	return c.next.RemoveUserFromWorkspace(ctx, tenantID, userID, workspaceID)
}
