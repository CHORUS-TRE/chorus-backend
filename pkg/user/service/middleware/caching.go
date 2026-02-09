package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/cache"
	common "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/service"

	"github.com/coocood/freecache"
)

const (
	userCacheSize          = 100 * 1024 * 1024 // Max 100MiB stored in memory
	defaultCacheExpiration = 1
	longCacheExpiration    = 60
)

func UserCaching(log *logger.ContextLogger) func(service.Userer) *Caching {
	return func(next service.Userer) *Caching {
		return &Caching{
			cache: cache.NewCache(freecache.NewCache(userCacheSize), log),
			next:  next,
		}
	}
}

type Caching struct {
	cache *cache.Cache
	next  service.Userer
}

func (c *Caching) CreateRole(ctx context.Context, role string) error {
	return c.next.CreateRole(ctx, role)
}

func (c *Caching) CreateUserRoles(ctx context.Context, tenantID, userID uint64, roles []model.UserRole) error {
	return c.next.CreateUserRoles(ctx, tenantID, userID, roles)
}

func (c *Caching) RemoveUserRoles(ctx context.Context, tenantID, userID uint64, userRoleIDs []uint64) error {
	return c.next.RemoveUserRoles(ctx, tenantID, userID, userRoleIDs)
}

func (c *Caching) GetRoles(ctx context.Context) (reply []*model.Role, err error) {
	entry := c.cache.NewEntry()
	reply = []*model.Role{}

	if ok := entry.Get(ctx, &reply); !ok {
		reply, err = c.next.GetRoles(ctx)
		if err == nil {
			entry.Set(ctx, longCacheExpiration, reply)
		}
	}

	return
}

func (c *Caching) ListUsers(ctx context.Context, req service.ListUsersReq) (users []*model.User, pagination *common.PaginationResult, err error) {
	entry := c.cache.NewEntry(cache.WithInterface(req))
	users = []*model.User{}
	pagination = &common.PaginationResult{}

	if ok := entry.Get(ctx, &users, &pagination); !ok {
		users, pagination, err = c.next.ListUsers(ctx, req)
		if err == nil {
			entry.Set(ctx, defaultCacheExpiration, users, pagination)
		}
	}

	return
}

func (c *Caching) GetUser(ctx context.Context, req service.GetUserReq) (reply *model.User, err error) {
	entry := c.cache.NewEntry(cache.WithInterface(req))
	reply = &model.User{}

	if !req.SkipCache {
		if ok := entry.Get(ctx, &reply); !ok {
			reply, err = c.next.GetUser(ctx, req)
			if err == nil {
				entry.Set(ctx, defaultCacheExpiration, reply)
			}
		}
	} else {
		reply, err = c.next.GetUser(ctx, req)
		if err == nil {
			entry.Set(ctx, defaultCacheExpiration, reply)
		}
	}

	return
}

func (c *Caching) SoftDeleteUser(ctx context.Context, req service.DeleteUserReq) error {
	return c.next.SoftDeleteUser(ctx, req)
}

func (c *Caching) UpdateUser(ctx context.Context, req service.UpdateUserReq) (*model.User, error) {
	return c.next.UpdateUser(ctx, req)
}

func (c *Caching) CreateUser(ctx context.Context, req service.CreateUserReq) (*model.User, error) {
	return c.next.CreateUser(ctx, req)
}

func (c *Caching) UpdateUserPassword(ctx context.Context, req service.UpdateUserPasswordReq) error {
	return c.next.UpdateUserPassword(ctx, req)
}

func (c *Caching) EnableUserTotp(ctx context.Context, req service.EnableTotpReq) error {
	return c.next.EnableUserTotp(ctx, req)
}

func (c *Caching) ResetUserTotp(ctx context.Context, req service.ResetTotpReq) (string, []string, error) {
	return c.next.ResetUserTotp(ctx, req)
}

func (c *Caching) ResetUserPassword(ctx context.Context, req service.ResetUserPasswordReq) error {
	return c.next.ResetUserPassword(ctx, req)
}

func (c *Caching) GetTotpRecoveryCodes(ctx context.Context, tenantID, userID uint64) ([]*model.TotpRecoveryCode, error) {
	return c.next.GetTotpRecoveryCodes(ctx, tenantID, userID)
}

func (c *Caching) DeleteTotpRecoveryCode(ctx context.Context, req *service.DeleteTotpRecoveryCodeReq) error {
	return c.next.DeleteTotpRecoveryCode(ctx, req)
}

func (c *Caching) UpsertGrants(ctx context.Context, grants []model.UserGrant) error {
	return c.next.UpsertGrants(ctx, grants)
}

func (c *Caching) DeleteGrants(ctx context.Context, tenantID uint64, userID uint64, clientID string) error {
	return c.next.DeleteGrants(ctx, tenantID, userID, clientID)
}

func (c *Caching) GetUserGrants(ctx context.Context, tenantID uint64, userID uint64, clientID string) ([]model.UserGrant, error) {
	return c.next.GetUserGrants(ctx, tenantID, userID, clientID)
}
