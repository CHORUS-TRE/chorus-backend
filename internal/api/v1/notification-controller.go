package v1

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/converter"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils"
	"github.com/CHORUS-TRE/chorus-backend/pkg/notification/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/notification/service"
)

type NotificationController struct {
	notification service.Notificationer
}

func NewNotificationController(notification service.Notificationer) NotificationController {
	return NotificationController{notification: notification}
}

func (c NotificationController) CountUnreadNotifications(ctx context.Context, empty *empty.Empty) (*chorus.CountUnreadNotificationsReply, error) {
	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	userID, err := jwt_model.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract user id from jwt-token")
	}

	count, err := c.notification.CountUnreadNotifications(ctx, service.CountUnreadNotificationRequest{TenantID: tenantID, UserID: userID})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to call 'CountUnreadNotifications': %v", err.Error())
	}

	return &chorus.CountUnreadNotificationsReply{Result: count}, nil
}

func (c NotificationController) MarkNotificationsAsRead(ctx context.Context, req *chorus.MarkNotificationsAsReadRequest) (*empty.Empty, error) {
	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	userID, err := jwt_model.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract user id from jwt-token")
	}

	if err = c.notification.MarkNotificationsAsRead(ctx, service.MarkNotificationsAsReadRequest{
		TenantID:        tenantID,
		UserID:          userID,
		NotificationIDs: req.NotificationIds,
		MarkAll:         req.MarkAll,
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to call 'MarkNotificationsAsRead': %v", err.Error())
	}

	return &empty.Empty{}, nil
}
func (c NotificationController) GetNotifications(ctx context.Context, req *chorus.GetNotificationsRequest) (*chorus.GetNotificationsReply, error) {
	tenantID, err := jwt_model.ExtractTenantID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract tenant id from jwt-token")
	}

	userID, err := jwt_model.ExtractUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "could not extract user id from jwt-token")
	}

	res, count, err := c.notification.GetNotifications(ctx, c.getNotificationToServiceRequest(tenantID, userID, req))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to call 'GetNotifications': %v", err.Error())
	}

	var notifications []*chorus.Notification
	for _, r := range res {
		notification, err := notificationFromBusiness(r)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "conversion error: %v", err.Error())
		}
		notifications = append(notifications, notification)
	}

	return &chorus.GetNotificationsReply{Result: notifications, TotalItems: count}, nil
}

func (c NotificationController) getNotificationToServiceRequest(tenantID, userID uint64, r *chorus.GetNotificationsRequest) service.GetNotificationsRequest {
	if r.Pagination == nil {
		r.Pagination = &chorus.PaginationQuery{}
	}
	if r.Pagination.Limit == 0 {
		r.Pagination.Limit = 20
	}
	if r.Pagination.Sort == nil {
		r.Pagination.Sort = &chorus.Sort{Type: "CREATEDAT", Order: "DESC"}
	}

	return service.GetNotificationsRequest{
		TenantID: tenantID,
		UserID:   userID,
		// Query:    r.Pagination.Query,
		IsRead: converter.FromProtoBoolValue(r.IsRead),
		Offset: uint64(r.Pagination.Offset),
		Limit:  uint64(r.Pagination.Limit),
		Sort: service.Sort{
			SortOrder: strings.ToUpper(r.Pagination.Sort.Order),
			SortType:  strings.ToUpper(r.Pagination.Sort.Type),
		},
	}
}

func notificationFromBusiness(r *model.Notification) (*chorus.Notification, error) {
	ca, err := converter.ToProtoTimestamp(r.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert createdAt timestamp: %w", err)
	}
	ra, err := converter.ToProtoTimestamp(utils.ToTime(r.ReadAt))
	if err != nil {
		return nil, fmt.Errorf("unable to convert createdAt timestamp: %w", err)
	}
	return &chorus.Notification{
		Id:        r.ID,
		TenantId:  r.TenantID,
		Message:   r.Message,
		CreatedAt: ca,
		ReadAt:    ra,
	}, nil
}
