package service

import (
	"context"
	"fmt"

	common "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/notification/model"
)

type Notificationer interface {
	CountUnreadNotifications(ctx context.Context, req CountUnreadNotificationRequest) (uint32, error)
	MarkNotificationsAsRead(ctx context.Context, req MarkNotificationsAsReadRequest) error
	GetNotifications(ctx context.Context, req GetNotificationsRequest) ([]*model.Notification, uint32, error)
}

// NotificationStore groups the database interface functions.
type NotificationStore interface {
	CreateNotification(ctx context.Context, notification *model.Notification, userIDs []uint64) error
	CountUnreadNotifications(ctx context.Context, tenantID, userID uint64) (uint32, error)
	MarkNotificationsAsRead(ctx context.Context, tenantID, userID uint64, notificationIDs []string, markAll bool) error
	GetNotifications(ctx context.Context, tenantID, userID uint64, query string, isRead *bool, offset, limit uint64, sort common.Sort) ([]*model.Notification, uint32, error)
}

type NotificationService struct {
	store NotificationStore
}

func NewNotificationService(store NotificationStore) *NotificationService {
	return &NotificationService{store: store}
}

func (s NotificationService) CountUnreadNotifications(ctx context.Context, req CountUnreadNotificationRequest) (uint32, error) {
	count, err := s.store.CountUnreadNotifications(ctx, req.TenantID, req.UserID)
	if err != nil {
		return 0, fmt.Errorf("unable to count unread notifications: %w", err)
	}
	return count, nil
}
func (s NotificationService) MarkNotificationsAsRead(ctx context.Context, req MarkNotificationsAsReadRequest) error {
	if err := s.store.MarkNotificationsAsRead(ctx, req.TenantID, req.UserID, req.NotificationIDs, req.MarkAll); err != nil {
		return fmt.Errorf("unable to mark notification as read: %w", err)
	}
	return nil
}
func (s NotificationService) GetNotifications(ctx context.Context, req GetNotificationsRequest) ([]*model.Notification, uint32, error) {
	notifications, count, err := s.store.GetNotifications(ctx, req.TenantID, req.UserID, req.Query, req.IsRead, req.Offset, req.Limit, req.Sort.ToBusinessSort())
	if err != nil {
		return nil, 0, fmt.Errorf("unable to get notifications: %w", err)
	}
	return notifications, count, nil
}
