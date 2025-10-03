package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/authorization"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	empty "google.golang.org/protobuf/types/known/emptypb"
)

var _ chorus.NotificationServiceServer = (*notificationControllerAuthorization)(nil)

type notificationControllerAuthorization struct {
	Authorization
	next chorus.NotificationServiceServer
}

func NotificationAuthorizing(logger *logger.ContextLogger, authorizer authorization.Authorizer) func(chorus.NotificationServiceServer) chorus.NotificationServiceServer {
	return func(next chorus.NotificationServiceServer) chorus.NotificationServiceServer {
		return &notificationControllerAuthorization{
			Authorization: Authorization{
				logger:     logger,
				authorizer: authorizer,
			},
			next: next,
		}
	}
}

func (c notificationControllerAuthorization) CountUnreadNotifications(ctx context.Context, empty *empty.Empty) (*chorus.CountUnreadNotificationsReply, error) {
	userID, err := c.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	err = c.IsAuthorized(ctx, authorization.PermissionCountUnreadNotifications, authorization.WithUser(userID))
	if err != nil {
		return nil, err
	}

	return c.next.CountUnreadNotifications(ctx, empty)
}
func (c notificationControllerAuthorization) MarkNotificationsAsRead(ctx context.Context, req *chorus.MarkNotificationsAsReadRequest) (*empty.Empty, error) {
	userID, err := c.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	err = c.IsAuthorized(ctx, authorization.PermissionMarkNotificationAsRead, authorization.WithUser(userID))
	if err != nil {
		return nil, err
	}

	return c.next.MarkNotificationsAsRead(ctx, req)
}
func (c notificationControllerAuthorization) GetNotifications(ctx context.Context, req *chorus.GetNotificationsRequest) (*chorus.GetNotificationsReply, error) {
	userID, err := c.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	err = c.IsAuthorized(ctx, authorization.PermissionListNotifications, authorization.WithUser(userID))
	if err != nil {
		return nil, err
	}

	return c.next.GetNotifications(ctx, req)
}
