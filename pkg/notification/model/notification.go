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
	Type               string                      `json:"type,omitempty"`
	SystemNotification *SystemNotification         `json:"system_notification,omitempty"`
	ApprovalRequest    *ApprovalRequestNotification `json:"approval_request,omitempty"`
}

type SystemNotification struct {
	RefreshJWTRequired bool `json:"refresh_jwt_required,omitempty"`
}

type ApprovalRequestNotification struct {
	ApprovalRequestID uint64 `json:"approval_request_id,omitempty"`
}
