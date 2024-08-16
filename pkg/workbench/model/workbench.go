package model

import (
	"fmt"
	"time"
)

// Workbench maps an entry in the 'workbenchs' database table.
type Workbench struct {
	ID uint64

	TenantID    uint64
	UserID      uint64
	WorkspaceID uint64

	Name        string
	ShortName   string
	Description string

	Status WorkbenchStatus

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// WorkbenchStatus represents the status of a workbench.
type WorkbenchStatus string

const (
	WorkbenchActive   WorkbenchStatus = "active"
	WorkbenchInactive WorkbenchStatus = "inactive"
	WorkbenchDeleted  WorkbenchStatus = "deleted"
)

func (s WorkbenchStatus) String() string {
	return string(s)
}

func ToWorkbenchStatus(status string) (WorkbenchStatus, error) {
	switch status {
	case WorkbenchActive.String():
		return WorkbenchActive, nil
	case WorkbenchInactive.String():
		return WorkbenchInactive, nil
	case WorkbenchDeleted.String():
		return WorkbenchDeleted, nil
	default:
		return "", fmt.Errorf("unexpected WorkbenchStatus: %s", status)
	}
}
