package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/cache"
	authorization_model "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
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

func (c *Caching) ListWorkspaces(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, filter model.WorkspaceFilter) (reply []*model.Workspace, paginationRes *common_model.PaginationResult, err error) {
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

func (c *Caching) ManageUserRoleInWorkspace(ctx context.Context, tenantID, userID uint64, role user_model.UserRole) error {
	return c.next.ManageUserRoleInWorkspace(ctx, tenantID, userID, role)
}

func (c *Caching) RemoveUserRoleInWorkspace(ctx context.Context, tenantID, userID, workspaceID uint64, roleName authorization_model.RoleName) error {
	return c.next.RemoveUserRoleInWorkspace(ctx, tenantID, userID, workspaceID, roleName)
}

func (c *Caching) RemoveUserFromWorkspace(ctx context.Context, tenantID, userID uint64, workspaceID uint64) error {
	return c.next.RemoveUserFromWorkspace(ctx, tenantID, userID, workspaceID)
}

func (c *Caching) GetWorkspaceSvc(ctx context.Context, tenantID, workspaceSvcID uint64) (*model.WorkspaceSvc, error) {
	return c.next.GetWorkspaceSvc(ctx, tenantID, workspaceSvcID)
}

func (c *Caching) ListWorkspaceSvcs(ctx context.Context, tenantID uint64, pagination *common_model.Pagination, filter service.WorkspaceSvcFilter) ([]*model.WorkspaceSvc, *common_model.PaginationResult, error) {
	return c.next.ListWorkspaceSvcs(ctx, tenantID, pagination, filter)
}

func (c *Caching) CreateWorkspaceSvc(ctx context.Context, svc *model.WorkspaceSvc) (*model.WorkspaceSvc, error) {
	return c.next.CreateWorkspaceSvc(ctx, svc)
}

func (c *Caching) UpdateWorkspaceSvc(ctx context.Context, svc *model.WorkspaceSvc) (*model.WorkspaceSvc, error) {
	return c.next.UpdateWorkspaceSvc(ctx, svc)
}

func (c *Caching) DeleteWorkspaceSvc(ctx context.Context, tenantID, workspaceSvcID uint64) error {
	return c.next.DeleteWorkspaceSvc(ctx, tenantID, workspaceSvcID)
}
