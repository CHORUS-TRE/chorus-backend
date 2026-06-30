package middleware

import (
	"context"

	"github.com/coocood/freecache"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/cache"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/terms-of-use/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/terms-of-use/service"
)

const (
	termsOfUseCacheSize    = 10 * 1024 * 1024 // 10MiB
	defaultCacheExpiration = 5                // seconds
	shortCacheExpiration   = 1                // seconds
)

func TermsOfUseCaching(log *logger.ContextLogger) func(service.TermsOfUseer) *Caching {
	return func(next service.TermsOfUseer) *Caching {
		return &Caching{
			cache: cache.NewCache(freecache.NewCache(termsOfUseCacheSize), log),
			next:  next,
		}
	}
}

type Caching struct {
	cache *cache.Cache
	next  service.TermsOfUseer
}

func (c *Caching) CreateTermsOfUseVersion(ctx context.Context, tenantID uint64, content string) (*model.TermsOfUseVersion, error) {
	return c.next.CreateTermsOfUseVersion(ctx, tenantID, content)
}

func (c *Caching) UpdateTermsOfUseVersion(ctx context.Context, tenantID, versionID uint64, content string) (*model.TermsOfUseVersion, error) {
	return c.next.UpdateTermsOfUseVersion(ctx, tenantID, versionID, content)
}

func (c *Caching) PublishTermsOfUseVersion(ctx context.Context, tenantID, versionID uint64) (*model.TermsOfUseVersion, error) {
	return c.next.PublishTermsOfUseVersion(ctx, tenantID, versionID)
}

func (c *Caching) GetTermsOfUseVersion(ctx context.Context, tenantID, versionID uint64) (reply *model.TermsOfUseVersion, err error) {
	entry := c.cache.NewEntry(cache.WithUint64(tenantID), cache.WithUint64(versionID))
	reply = &model.TermsOfUseVersion{}

	if ok := entry.Get(ctx, &reply); !ok {
		reply, err = c.next.GetTermsOfUseVersion(ctx, tenantID, versionID)
		if err == nil {
			entry.Set(ctx, defaultCacheExpiration, reply)
		}
	}

	return
}

func (c *Caching) ListTermsOfUseVersions(ctx context.Context, tenantID uint64, pagination *common_model.Pagination) ([]*model.TermsOfUseVersion, *common_model.PaginationResult, error) {
	return c.next.ListTermsOfUseVersions(ctx, tenantID, pagination)
}

func (c *Caching) GetCurrentTermsOfUseVersion(ctx context.Context, tenantID uint64) (reply *model.TermsOfUseVersion, err error) {
	entry := c.cache.NewEntry(cache.WithUint64(tenantID))
	reply = &model.TermsOfUseVersion{}

	if ok := entry.Get(ctx, &reply); !ok {
		reply, err = c.next.GetCurrentTermsOfUseVersion(ctx, tenantID)
		if err == nil {
			entry.Set(ctx, shortCacheExpiration, reply)
		}
	}

	return
}

func (c *Caching) ListTermsOfUseAcceptances(ctx context.Context, tenantID, userID uint64) ([]*model.TermsOfUseAcceptance, error) {
	return c.next.ListTermsOfUseAcceptances(ctx, tenantID, userID)
}

func (c *Caching) GetMyTermsOfUseStatus(ctx context.Context, tenantID, userID uint64) (reply bool, err error) {
	entry := c.cache.NewEntry(cache.WithUint64(tenantID), cache.WithUint64(userID))

	if ok := entry.Get(ctx, &reply); !ok {
		reply, err = c.next.GetMyTermsOfUseStatus(ctx, tenantID, userID)
		if err == nil {
			entry.Set(ctx, shortCacheExpiration, reply)
		}
	}

	return
}

func (c *Caching) AcceptTermsOfUse(ctx context.Context, tenantID, userID uint64) (*model.TermsOfUseAcceptance, error) {
	return c.next.AcceptTermsOfUse(ctx, tenantID, userID)
}
