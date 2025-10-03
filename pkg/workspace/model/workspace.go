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

	IsMain bool

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (s Workspace) GetClusterName() string {
	return GetWorkspaceClusterName(s.ID)
}

func GetWorkspaceClusterName(id uint64) string {
	return fmt.Sprintf("workspace%v", id)
}

func GetIDFromClusterName(clusterName string) (uint64, error) {
	var id uint64
	_, err := fmt.Sscanf(clusterName, "workspace%d", &id)
	if err != nil {
		return 0, fmt.Errorf("unable to get workspace ID from cluster name %s: %w", clusterName, err)
	}
	return id, nil
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

func (Workspace) IsValidSortType(sortType string) bool {
	validSortTypes := map[string]bool{
		"id":          true,
		"name":        true,
		"shortname":   true,
		"description": true,
		"status":      true,
		"isMain":      true,
		"createdat":   true,
	}

	return validSortTypes[sortType]
}

type WorkspaceFile struct {
	Path string

	Name        string
	IsDirectory bool
	Size        int64
	MimeType    string

	UpdatedAt time.Time

	Content []byte
}
