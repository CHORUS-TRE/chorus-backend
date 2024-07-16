package middleware

import (
	"context"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/common"
	"github.com/CHORUS-TRE/chorus-backend/pkg/notification/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/notification/service"

	"go.uber.org/zap"
)

type notificationServiceLogging struct {
	logger *logger.ContextLogger
	next   service.Notificationer
}

func Logging(log *logger.ContextLogger) func(service.Notificationer) service.Notificationer {
	l := logger.With(log, zap.String("layer", "service"))
	return func(next service.Notificationer) service.Notificationer {
		return &notificationServiceLogging{
			logger: l,
			next:   next,
		}
	}
}

func (c notificationServiceLogging) CountUnreadNotifications(ctx context.Context, req service.CountUnreadNotificationRequest) (uint32, error) {
	log := logger.With(c.logger,
		zap.String("service", "CountUnreadNotifications"),
		zap.Uint64("tenant_id", req.TenantID),
		zap.Uint64("user_id", req.UserID),
	)
	c.logger.Debug(ctx, logger.LoggerMessageRequestStarted)

	now := time.Now()

	res, err := c.next.CountUnreadNotifications(ctx, req)
	return res, common.LogErrorIfAny(err, ctx, now, logger.With(log, zap.Any("result", res)))
}
func (c notificationServiceLogging) MarkNotificationsAsRead(ctx context.Context, req service.MarkNotificationsAsReadRequest) error {
	log := logger.With(c.logger,
		zap.String("service", "MarkNotificationsAsRead"),
		zap.Uint64("tenant_id", req.TenantID),
		zap.Uint64("user_id", req.UserID),
		zap.Strings("notifications", req.NotificationIDs),
		zap.Bool("mark_all", req.MarkAll),
	)
	c.logger.Debug(ctx, logger.LoggerMessageRequestStarted)

	now := time.Now()

	err := c.next.MarkNotificationsAsRead(ctx, req)
	return common.LogErrorIfAny(err, ctx, now, log)
}
func (c notificationServiceLogging) GetNotifications(ctx context.Context, req service.GetNotificationsRequest) ([]*model.Notification, uint32, error) {
	log := logger.With(c.logger,
		zap.String("service", "CountUnreadNotifications"),
		zap.Uint64("tenant_id", req.TenantID),
		zap.Uint64("user_id", req.UserID),
		zap.String("query", req.Query),
		zap.Boolp("is_read", req.IsRead),
		zap.Uint64("offset", req.Offset),
		zap.Uint64("limit", req.Limit),
		zap.Any("sort", req.Sort),
	)
	c.logger.Debug(ctx, logger.LoggerMessageRequestStarted)

	now := time.Now()

	res, count, err := c.next.GetNotifications(ctx, req)
	return res, count, common.LogErrorIfAny(err, ctx, now, logger.With(log, zap.Any("result", res)))
}
