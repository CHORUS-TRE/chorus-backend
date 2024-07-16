package model

import (
	"time"
)

type Notification struct {
	ID        string
	TenantID  uint64
	Message   string
	CreatedAt time.Time
	ReadAt    *time.Time
}

var NotificationSortTypeToString = map[string]string{
	"ID":        "n.id",
	"MESSAGE":   "n.message",
	"CREATEDAT": "n.createdat",
}
