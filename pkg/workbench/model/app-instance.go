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

	K8sState  K8sAppInstanceState
	K8sStatus K8sAppInstanceStatus
	Status    AppInstanceStatus

	InitialResolutionWidth  uint32
	InitialResolutionHeight uint32

	AppName                *string
	AppDockerImageRegistry *string
	AppDockerImageName     *string
	AppDockerImageTag      *string

	AppShmSize             *string
	AppKioskConfigURL      *string
	AppMaxCPU              *string
	AppMinCPU              *string
	AppMaxMemory           *string
	AppMinMemory           *string
	AppMaxEphemeralStorage *string
	AppMinEphemeralStorage *string
	AppIconURL             *string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type K8sAppInstanceState string

const (
	K8sAppInstanceStateRunning K8sAppInstanceState = "Running"
	K8sAppInstanceStateStopped K8sAppInstanceState = "Stopped"
	K8sAppInstanceStateKilled  K8sAppInstanceState = "Killed"
)

type K8sAppInstanceStatus string

const (
	K8sAppInstanceStatusUnknown     K8sAppInstanceStatus = "Unknown"
	K8sAppInstanceStatusRunning     K8sAppInstanceStatus = "Running"
	K8sAppInstanceStatusComplete    K8sAppInstanceStatus = "Complete"
	K8sAppInstanceStatusProgressing K8sAppInstanceStatus = "Progressing"
	K8sAppInstanceStatusFailed      K8sAppInstanceStatus = "Failed"
)

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
