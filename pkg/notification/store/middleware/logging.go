package middleware

import (
	"context"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/common"
	common_model "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/notification/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/notification/service"

	"go.uber.org/zap"
)

type notificationStorageLogging struct {
	logger *logger.ContextLogger
	next   service.NotificationStore
}

func Logging(log *logger.ContextLogger) func(service.NotificationStore) service.NotificationStore {
	l := logger.With(log, zap.String("layer", "store"))
	return func(next service.NotificationStore) service.NotificationStore {
		return &notificationStorageLogging{
			logger: l,
			next:   next,
		}
	}
}

func (s *notificationStorageLogging) CreateNotification(ctx context.Context, notification *model.Notification, userIDs []uint64) error {
	log := logger.With(s.logger,
		zap.String("service", "CreateNotification"),
		zap.Uint64("tenant_id", notification.TenantID),
		zap.Any("notification", notification),
	)
	s.logger.Debug(ctx, logger.LoggerMessageRequestStarted)

	now := time.Now()

	err := s.next.CreateNotification(ctx, notification, userIDs)
	return common.LogErrorIfAny(err, ctx, now, log)
}

func (s *notificationStorageLogging) CountUnreadNotifications(ctx context.Context, tenantID, userID uint64) (uint32, error) {
	log := logger.With(s.logger,
		zap.String("service", "CountUnreadNotifications"),
		zap.Uint64("tenant_id", tenantID),
		zap.Uint64("user_id", tenantID),
	)
	s.logger.Debug(ctx, logger.LoggerMessageRequestStarted)

	now := time.Now()

	res, err := s.next.CountUnreadNotifications(ctx, tenantID, userID)
	return res, common.LogErrorIfAny(err, ctx, now, logger.With(log, zap.Any("result", res)))
}

func (s *notificationStorageLogging) MarkNotificationsAsRead(ctx context.Context, tenantID, userID uint64, notificationIDs []string, markAll bool) error {
	log := logger.With(s.logger,
		zap.String("service", "MarkNotificationsAsRead"),
		zap.Uint64("tenant_id", tenantID),
		zap.Uint64("user_id", tenantID),
		zap.Strings("notification-ids", notificationIDs),
		zap.Bool("mark_all", markAll),
	)
	s.logger.Debug(ctx, logger.LoggerMessageRequestStarted)

	now := time.Now()

	err := s.next.MarkNotificationsAsRead(ctx, tenantID, userID, notificationIDs, markAll)
	return common.LogErrorIfAny(err, ctx, now, log)
}

func (s *notificationStorageLogging) GetNotifications(ctx context.Context, tenantID, userID uint64, query string, isRead *bool, offset, limit uint64, sort common_model.Sort) ([]*model.Notification, uint32, error) {
	log := logger.With(s.logger,
		zap.String("service", "CountUnreadNotifications"),
		zap.Uint64("tenant_id", tenantID),
		zap.Uint64("user_id", userID),
		zap.String("query", query),
		zap.Boolp("is_read", isRead),
		zap.Uint64("offset", offset),
		zap.Uint64("limit", limit),
		zap.Any("sort", sort),
	)
	s.logger.Debug(ctx, logger.LoggerMessageRequestStarted)

	now := time.Now()

	res, count, err := s.next.GetNotifications(ctx, tenantID, userID, query, isRead, offset, limit, sort)
	return res, count, common.LogErrorIfAny(err, ctx, now, logger.With(log, zap.Any("result", res)))
}
