package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/cache"
	"github.com/CHORUS-TRE/chorus-backend/pkg/devstore/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/devstore/service"

	"github.com/coocood/freecache"
)

const (
	devstoreCacheSize      = 300 * 1024 * 1024 // Max 300MiB stored in memory
	shortCacheExpiration   = 1
	defaultCacheExpiration = 5
	longCacheExpiration    = 60
)

func DevstoreCaching(log *logger.ContextLogger) func(service.Devstorer) *Caching {
	return func(next service.Devstorer) *Caching {
		return &Caching{
			cache: cache.NewCache(freecache.NewCache(devstoreCacheSize), log),
			next:  next,
		}
	}
}

type Caching struct {
	cache *cache.Cache
	next  service.Devstorer
}

func (s *Caching) ListEntries(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64) (reply []*model.DevstoreEntry, err error) {
	entry := s.cache.NewEntry(cache.WithUint64(tenantID), cache.WithString(string(scope)), cache.WithUint64(scopeID))
	reply = []*model.DevstoreEntry{}

	if ok := entry.Get(ctx, &reply); !ok {
		reply, err = s.next.ListEntries(ctx, tenantID, scope, scopeID)
		if err == nil {
			entry.Set(ctx, defaultCacheExpiration, reply)
		}
	}
	return
}

func (s *Caching) GetEntry(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64, key string) (reply *model.DevstoreEntry, err error) {
	entry := s.cache.NewEntry(cache.WithUint64(tenantID), cache.WithString(string(scope)), cache.WithUint64(scopeID), cache.WithString(key))
	reply = &model.DevstoreEntry{}

	if ok := entry.Get(ctx, &reply); !ok {
		reply, err = s.next.GetEntry(ctx, tenantID, scope, scopeID, key)
		if err == nil {
			entry.Set(ctx, shortCacheExpiration, reply)
		}
	}
	return
}

func (s *Caching) PutEntry(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64, key string, value string) (*model.DevstoreEntry, error) {
	return s.next.PutEntry(ctx, tenantID, scope, scopeID, key, value)
}

func (s *Caching) DeleteEntry(ctx context.Context, tenantID uint64, scope model.DevStoreScope, scopeID uint64, key string) error {
	return s.next.DeleteEntry(ctx, tenantID, scope, scopeID, key)
}
