package model

import (
	"errors"
	"time"
)

// AppInstance maps an entry in the 'app_instances' database table.
type AppInstance struct {
	ID uint64

	TenantID    uint64
	UserID      uint64
	AppID       uint64
	WorkspaceID uint64
	WorkbenchID uint64

	Status AppInstanceStatus

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// AppInstanceStatus represents the status of an app instance.
type AppInstanceStatus string

const (
	AppInstanceActive   AppInstanceStatus = "active"
	AppInstanceInactive AppInstanceStatus = "inactive"
	AppInstanceDeleted  AppInstanceStatus = "deleted"
)

func (s AppInstanceStatus) String() string {
	return string(s)
}

func ToAppInstanceStatus(status string) (AppInstanceStatus, error) {
	switch status {
	case AppInstanceActive.String():
		return AppInstanceActive, nil
	case AppInstanceInactive.String():
		return AppInstanceInactive, nil
	case AppInstanceDeleted.String():
		return AppInstanceDeleted, nil
	default:
		return "", errors.New("unexpected AppInstanceStatus: " + status)
	}
}
