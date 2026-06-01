package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/cache"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"

	"github.com/coocood/freecache"
)

var _ service.Authorizer = (*Caching)(nil)

const (
	authorizationCacheSize = 10 * 1024 * 1024 // Max 10MiB stored in memory
	defaultCacheExpiration = 5
)

func AuthorizationCaching(log *logger.ContextLogger) func(service.Authorizer) *Caching {
	return func(next service.Authorizer) *Caching {
		return &Caching{
			cache: cache.NewCache(freecache.NewCache(authorizationCacheSize), log),
			next:  next,
		}
	}
}

type Caching struct {
	cache *cache.Cache
	next  service.Authorizer
}

func (c *Caching) GetAuthorizationSchema() *model.AuthorizationSchema {
	// cpu bound no cache
	return c.next.GetAuthorizationSchema()
}

func (c *Caching) CreateDynamicRole(ctx context.Context, user []model.Role, role *model.RoleDefinition, validationContext model.Context) (*model.RoleDefinition, error) {
	return c.next.CreateDynamicRole(ctx, user, role, validationContext)
}

func (c *Caching) IsUserAllowed(user []model.Role, permission model.Permission) (bool, error) {
	// cpu bound no cache
	return c.next.IsUserAllowed(user, permission)
}

func (c *Caching) ExplainIsUserAllowed(user []model.Role, permission model.Permission) string {
	// cpu bound no cache
	return c.next.ExplainIsUserAllowed(user, permission)
}

func (c *Caching) GetUserPermissions(user []model.Role) ([]model.Permission, error) {
	// cpu bound no cache
	return c.next.GetUserPermissions(user)
}

func (c *Caching) GetContextListForPermission(user []model.Role, permissionName model.PermissionName) ([]model.Context, error) {
	// cpu bound no cache
	return c.next.GetContextListForPermission(user, permissionName)
}

func (c *Caching) FindUsersWithPermission(ctx context.Context, tenantID uint64, filter model.FindUsersWithPermissionFilter) ([]uint64, error) {
	entry := c.cache.NewEntry(cache.WithInterface(filter), cache.WithUint64(tenantID))

	userIDs := []uint64{}
	if ok := entry.Get(ctx, &userIDs); !ok {
		var err error
		userIDs, err = c.next.FindUsersWithPermission(ctx, tenantID, filter)
		if err == nil {
			entry.Set(ctx, defaultCacheExpiration, userIDs)
		}
	}

	return userIDs, nil
}

func (c *Caching) GetRolesGrantingPermission(permissionName model.PermissionName) []model.RoleName {
	entry := c.cache.NewEntry(cache.WithInterface(permissionName))
	roleNames := []model.RoleName{}

	if ok := entry.Get(context.Background(), &roleNames); !ok {
		roleNames = c.next.GetRolesGrantingPermission(permissionName)
		entry.Set(context.Background(), defaultCacheExpiration, roleNames)
	}

	return roleNames
}

func (c *Caching) GetRoleDefinition(roleName model.RoleName) (*model.RoleDefinition, bool) {
	entry := c.cache.NewEntry(cache.WithInterface(roleName))
	roleDef := &model.RoleDefinition{}
	found := false

	if ok := entry.Get(context.Background(), &roleDef, &found); !ok {
		roleDef, found = c.next.GetRoleDefinition(roleName)
		entry.Set(context.Background(), defaultCacheExpiration, roleDef, found)
	}

	return roleDef, found
}

func (c *Caching) IsRoleInScope(roleName model.RoleName, scopes ...model.RoleScope) bool {
	// cpu bound no cache
	return c.next.IsRoleInScope(roleName, scopes...)
}

func (c *Caching) CanAssignRole(user []model.Role, roleName model.RoleName, assignmentContext model.Context) (bool, error) {
	return c.next.CanAssignRole(user, roleName, assignmentContext)
}
