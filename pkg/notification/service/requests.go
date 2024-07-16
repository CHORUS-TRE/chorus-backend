package service

import common "github.com/CHORUS-TRE/chorus-backend/pkg/common/model"

type GetNotificationsRequest struct {
	TenantID uint64 `validate:"required"`
	UserID   uint64 `validate:"required"`
	Query    string
	IsRead   *bool
	Offset   uint64
	Limit    uint64 `validate:"max=500"`
	Sort     Sort   `validate:"required"`
}

type Sort struct {
	SortOrder string `validate:"required,oneof=DESC ASC"`
	SortType  string `validate:"required,oneof=ID MESSAGE CREATEDAT"`
}

func (s Sort) ToBusinessSort() common.Sort {
	sort := common.Sort{
		SortOrder: s.SortOrder,
		SortType:  s.SortType,
	}
	return sort
}

type MarkNotificationsAsReadRequest struct {
	TenantID        uint64   `validate:"required"`
	UserID          uint64   `validate:"required"`
	NotificationIDs []string `validate:"required_without=MarkAll,dive,required"`
	MarkAll         bool
}

type CountUnreadNotificationRequest struct {
	TenantID uint64 `validate:"required"`
	UserID   uint64 `validate:"required"`
}
