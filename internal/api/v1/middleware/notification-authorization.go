package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/golang/protobuf/ptypes/empty"
)

type notificationControllerAuthorization struct {
	authorization
	next chorus.NotificationServiceServer
}

func NotificationAuthorizing(logger *logger.ContextLogger, authorizedRoles []string) func(chorus.NotificationServiceServer) chorus.NotificationServiceServer {
	return func(next chorus.NotificationServiceServer) chorus.NotificationServiceServer {
		return &notificationControllerAuthorization{
			authorization: authorization{
				logger:          logger,
				authorizedRoles: authorizedRoles,
			},
			next: next,
		}
	}
}

func (c notificationControllerAuthorization) CountUnreadNotifications(ctx context.Context, empty *empty.Empty) (*chorus.CountUnreadNotificationsReply, error) {
	err := c.isAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}

	return c.next.CountUnreadNotifications(ctx, empty)
}
func (c notificationControllerAuthorization) MarkNotificationsAsRead(ctx context.Context, req *chorus.MarkNotificationsAsReadRequest) (*empty.Empty, error) {
	err := c.isAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}

	return c.next.MarkNotificationsAsRead(ctx, req)
}
func (c notificationControllerAuthorization) GetNotifications(ctx context.Context, req *chorus.GetNotificationsRequest) (*chorus.GetNotificationsReply, error) {
	err := c.isAuthenticatedAndAuthorized(ctx)
	if err != nil {
		return nil, err
	}

	return c.next.GetNotifications(ctx, req)
}
