package model

import (
	"time"
)

type Notification struct {
	ID        string
	TenantID  uint64
	UserID    uint64
	Message   string
	Content   NotificationContent
	CreatedAt time.Time
	ReadAt    *time.Time
}

var NotificationSortTypeToString = map[string]string{
	"ID":        "n.id",
	"MESSAGE":   "n.message",
	"CREATEDAT": "n.createdat",
}

type NotificationContent struct {
	Type               string
	SystemNotification SystemNotification
	ApprovalRequest    ApprovalRequestNotification
}

type SystemNotification struct {
	RefreshJWTRequired bool
}

type ApprovalRequestNotification struct {
	ApprovalRequestID uint64
}
