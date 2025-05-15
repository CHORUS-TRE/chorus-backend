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

	InitialResolutionWidth  uint32
	InitialResolutionHeight uint32

	K8sStatus K8sWorkbenchStatus
	Status    WorkbenchStatus

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (s Workbench) GetClusterName() string {
	return GetWorkbenchClusterName(s.ID)
}

func GetWorkbenchClusterName(id uint64) string {
	return fmt.Sprintf("workbench%v", id)
}

func GetIDFromClusterName(clusterName string) (uint64, error) {
	var id uint64
	_, err := fmt.Sscanf(clusterName, "workbench%v", &id)
	if err != nil {
		return 0, fmt.Errorf("unable to get workbench ID from cluster name %s: %w", clusterName, err)
	}
	return id, nil
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

type K8sWorkbenchStatus string

const (
	K8sWorkbenchStatusRunning     K8sWorkbenchStatus = "Running"
	K8sWorkbenchStatusProgressing K8sWorkbenchStatus = "Progressing"
	K8sWorkbenchStatusFailed      K8sWorkbenchStatus = "Failed"
)

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
