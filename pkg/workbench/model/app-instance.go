package model

import (
	"errors"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/k8s"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils"
)

// AppInstance maps an entry in the 'app_instances' database table.
type AppInstance struct {
	ID uint64

	TenantID    uint64
	UserID      uint64
	AppID       uint64
	WorkspaceID uint64
	WorkbenchID uint64

	Status    AppInstanceStatus
	K8sStatus K8sAppInstanceStatus
	K8sState  K8sAppInstanceState

	InitialResolutionWidth  uint32
	InitialResolutionHeight uint32

	KioskConfigJWTToken string

	AppName                *string
	AppDockerImageRegistry *string
	AppDockerImageName     *string
	AppDockerImageTag      *string

	AppShmSize             *string
	AppKioskConfigURL      *string
	AppKioskConfigJWTURL   *string
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

func (a *AppInstance) ToK8sAppInstance() k8s.AppInstance {
	return k8s.AppInstance{
		ID:      a.ID,
		AppName: utils.ToString(a.AppName),

		AppRegistry: utils.ToString(a.AppDockerImageRegistry),
		AppImage:    utils.ToString(a.AppDockerImageName),
		AppTag:      utils.ToString(a.AppDockerImageTag),

		K8sState: a.K8sState.String(),

		KioskConfigURL:      utils.ToString(a.AppKioskConfigURL),
		KioskConfigJWTURL:   utils.ToString(a.AppKioskConfigJWTURL),
		KioskConfigJWTToken: a.KioskConfigJWTToken,

		ShmSize:             utils.ToString(a.AppShmSize),
		MaxCPU:              utils.ToString(a.AppMaxCPU),
		MinCPU:              utils.ToString(a.AppMinCPU),
		MaxMemory:           utils.ToString(a.AppMaxMemory),
		MinMemory:           utils.ToString(a.AppMinMemory),
		MaxEphemeralStorage: utils.ToString(a.AppMaxEphemeralStorage),
		MinEphemeralStorage: utils.ToString(a.AppMinEphemeralStorage),
	}
}

type K8sAppInstanceState string

const (
	K8sAppInstanceStateRunning K8sAppInstanceState = "Running"
	K8sAppInstanceStateStopped K8sAppInstanceState = "Stopped"
	K8sAppInstanceStateKilled  K8sAppInstanceState = "Killed"
)

func (s K8sAppInstanceState) ToAppInstanceStatus() AppInstanceStatus {
	switch s {
	case K8sAppInstanceStateRunning:
		return AppInstanceActive
	case K8sAppInstanceStateStopped:
		return AppInstanceInactive
	case K8sAppInstanceStateKilled:
		return AppInstanceDeleted
	default:
		return AppInstanceActive
	}
}

func (s K8sAppInstanceState) String() string {
	return string(s)
}

type K8sAppInstanceStatus string

const (
	K8sAppInstanceStatusUnknown     K8sAppInstanceStatus = "Unknown"
	K8sAppInstanceStatusRunning     K8sAppInstanceStatus = "Running"
	K8sAppInstanceStatusComplete    K8sAppInstanceStatus = "Complete"
	K8sAppInstanceStatusProgressing K8sAppInstanceStatus = "Progressing"
	K8sAppInstanceStatusFailed      K8sAppInstanceStatus = "Failed"
)

func (s K8sAppInstanceStatus) ToAppInstanceStatus() AppInstanceStatus {
	switch s {
	case K8sAppInstanceStatusUnknown:
		return AppInstanceInactive
	case K8sAppInstanceStatusRunning:
		return AppInstanceActive
	case K8sAppInstanceStatusComplete:
		return AppInstanceDeleted
	case K8sAppInstanceStatusFailed:
		return AppInstanceDeleted
	case K8sAppInstanceStatusProgressing:
		return AppInstanceInactive
	default:
		return AppInstanceInactive
	}
}

func (s K8sAppInstanceStatus) String() string {
	return string(s)
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

func (AppInstance) IsValidSortType(sortType string) bool {
	validSortTypes := map[string]bool{
		"id":                     true,
		"appid":                  true,
		"workspaceid":            true,
		"workbenchid":            true,
		"status":                 true,
		"createdat":              true,
		"appname":                true,
		"appdockerimageregistry": true,
		"appdockerimagename":     true,
	}

	return validSortTypes[sortType]
}
