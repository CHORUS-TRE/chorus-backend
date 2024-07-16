package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/cache"
	"github.com/CHORUS-TRE/chorus-backend/pkg/notification/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/notification/service"

	"github.com/coocood/freecache"
)

const (
	notificationCacheSize  = 100 * 1024 * 1024 // Max 100MiB stored in memory
	defaultCacheExpiration = 2
)

func NotificationCaching(log *logger.ContextLogger) func(service.Notificationer) *Caching {
	return func(next service.Notificationer) *Caching {
		return &Caching{
			cache: cache.NewCache(freecache.NewCache(notificationCacheSize), log),
			next:  next,
		}
	}
}

type Caching struct {
	cache *cache.Cache
	next  service.Notificationer
}

func (c *Caching) CountUnreadNotifications(ctx context.Context, req service.CountUnreadNotificationRequest) (reply uint32, err error) {
	entry := c.cache.NewEntry(cache.WithInterface(req))

	if ok := entry.Get(ctx, &reply); !ok {
		reply, err = c.next.CountUnreadNotifications(ctx, req)
		if err == nil {
			entry.Set(ctx, defaultCacheExpiration, reply)
		}
	}

	return
}

func (c *Caching) MarkNotificationsAsRead(ctx context.Context, req service.MarkNotificationsAsReadRequest) error {
	return c.next.MarkNotificationsAsRead(ctx, req)
}

func (c *Caching) GetNotifications(ctx context.Context, req service.GetNotificationsRequest) (reply []*model.Notification, count uint32, err error) {
	entry := c.cache.NewEntry(cache.WithInterface(req))

	if ok := entry.Get(ctx, &reply, &count); !ok {
		reply, count, err = c.next.GetNotifications(ctx, req)
		if err == nil {
			entry.Set(ctx, defaultCacheExpiration, reply, count)
		}
	}

	return
}
