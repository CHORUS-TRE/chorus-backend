package middleware

import (
	"context"

	"github.com/coocood/freecache"

	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/cache"
	"github.com/CHORUS-TRE/chorus-backend/pkg/platform-settings/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/platform-settings/service"
)

const (
	platformSettingsCacheSize       = 100 * 1024 * 1024 // 100MiB
	platformSettingsCacheExpiration = 5                 // seconds
)

func PlatformSettingsCaching(log *logger.ContextLogger) func(service.PlatformSettingser) *Caching {
	return func(next service.PlatformSettingser) *Caching {
		return &Caching{
			cache: cache.NewCache(freecache.NewCache(platformSettingsCacheSize), log),
			next:  next,
		}
	}
}

type Caching struct {
	cache *cache.Cache
	next  service.PlatformSettingser
}

func (c *Caching) GetPlatformSettings(ctx context.Context) (reply *model.PlatformSettings, err error) {
	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		tenantID = 1
	}

	entry := c.cache.NewEntry(cache.WithUint64(tenantID))
	reply = &model.PlatformSettings{}

	if ok := entry.Get(ctx, &reply); !ok {
		reply, err = c.next.GetPlatformSettings(ctx)
		if err == nil {
			entry.Set(ctx, platformSettingsCacheExpiration, reply)
		}
	}

	return
}

func (c *Caching) UpdatePlatformSettings(ctx context.Context, settings *model.PlatformSettings) (*model.PlatformSettings, error) {
	return c.next.UpdatePlatformSettings(ctx, settings)
}
