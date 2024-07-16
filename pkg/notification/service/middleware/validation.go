package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/pkg/notification/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/notification/service"
	"github.com/pkg/errors"

	val "github.com/go-playground/validator/v10"
)

type validation struct {
	next     service.Notificationer
	validate *val.Validate
}

func Validation(validate *val.Validate) func(service.Notificationer) service.Notificationer {
	return func(next service.Notificationer) service.Notificationer {
		return &validation{
			next:     next,
			validate: validate,
		}
	}
}

func (v validation) CountUnreadNotifications(ctx context.Context, req service.CountUnreadNotificationRequest) (uint32, error) {
	if err := v.validate.Struct(req); err != nil {
		return 0, err
	}
	return v.next.CountUnreadNotifications(ctx, req)
}
func (v validation) MarkNotificationsAsRead(ctx context.Context, req service.MarkNotificationsAsReadRequest) error {
	if err := v.validate.Struct(req); err != nil {
		return errors.Wrap(err, "unable to mark notification as read")
	}
	return v.next.MarkNotificationsAsRead(ctx, req)
}
func (v validation) GetNotifications(ctx context.Context, req service.GetNotificationsRequest) ([]*model.Notification, uint32, error) {
	if err := v.validate.Struct(req); err != nil {
		return nil, 0, err
	}
	return v.next.GetNotifications(ctx, req)
}
