package model

import (
	"fmt"
	"time"
)

// Workspace maps an entry in the 'workspaces' database table.
type Workspace struct {
	ID uint64

	TenantID uint64
	UserID   uint64

	Name        string
	ShortName   string
	Description string

	Status WorkspaceStatus

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// WorkspaceStatus represents the status of a workspace.
type WorkspaceStatus string

const (
	WorkspaceActive   WorkspaceStatus = "active"
	WorkspaceInactive WorkspaceStatus = "inactive"
	WorkspaceDeleted  WorkspaceStatus = "deleted"
)

func (s WorkspaceStatus) String() string {
	return string(s)
}

func ToWorkspaceStatus(status string) (WorkspaceStatus, error) {
	switch status {
	case WorkspaceActive.String():
		return WorkspaceActive, nil
	case WorkspaceInactive.String():
		return WorkspaceInactive, nil
	case WorkspaceDeleted.String():
		return WorkspaceDeleted, nil
	default:
		return "", fmt.Errorf("unexpected WorkspaceStatus: %s", status)
	}
}
