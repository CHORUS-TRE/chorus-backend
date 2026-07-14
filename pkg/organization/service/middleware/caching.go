package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/cache"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/organization/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/organization/service"

	"github.com/coocood/freecache"
)

const (
	organizationCacheSize  = 100 * 1024 * 1024 // Max 100MiB stored in memory
	defaultCacheExpiration = 5
)

func OrganizationCaching(log *logger.ContextLogger) func(service.Organizationer) *Caching {
	return func(next service.Organizationer) *Caching {
		return &Caching{
			cache: cache.NewCache(freecache.NewCache(organizationCacheSize), log),
			next:  next,
		}
	}
}

type Caching struct {
	cache *cache.Cache
	next  service.Organizationer
}

func (c *Caching) ListOrganizations(ctx context.Context, tenantID uint64, pagination *common_model.Pagination) (reply []*model.Organization, paginationRes *common_model.PaginationResult, err error) {
	entry := c.cache.NewEntry(cache.WithUint64(tenantID), cache.WithInterface(pagination))
	reply = []*model.Organization{}
	paginationRes = &common_model.PaginationResult{}

	if ok := entry.Get(ctx, &reply, &paginationRes); !ok {
		reply, paginationRes, err = c.next.ListOrganizations(ctx, tenantID, pagination)
		if err == nil {
			entry.Set(ctx, defaultCacheExpiration, reply, paginationRes)
		}
	}

	return
}

func (c *Caching) GetOrganization(ctx context.Context, tenantID, id uint64) (reply *model.Organization, err error) {
	entry := c.cache.NewEntry(cache.WithUint64(tenantID), cache.WithUint64(id))
	reply = &model.Organization{}

	if ok := entry.Get(ctx, &reply); !ok {
		reply, err = c.next.GetOrganization(ctx, tenantID, id)
		if err == nil {
			entry.Set(ctx, defaultCacheExpiration, reply)
		}
	}

	return
}

func (c *Caching) GetOrganizationLogo(ctx context.Context, tenantID, id uint64) (*model.OrganizationLogo, error) {
	return c.next.GetOrganizationLogo(ctx, tenantID, id)
}

func (c *Caching) CreateOrganization(ctx context.Context, organization *model.Organization) (*model.Organization, error) {
	return c.next.CreateOrganization(ctx, organization)
}

func (c *Caching) UpdateOrganization(ctx context.Context, organization *model.Organization) (*model.Organization, error) {
	return c.next.UpdateOrganization(ctx, organization)
}

func (c *Caching) DeleteOrganization(ctx context.Context, tenantID, id uint64) error {
	return c.next.DeleteOrganization(ctx, tenantID, id)
}
