package model

import (
	"errors"
	"time"
)

// App maps an entry in the 'apps' database table.
type App struct {
	ID uint64

	TenantID uint64
	UserID   uint64

	Name        string
	Description string
	Status      AppStatus

	DockerImageRegistry string
	DockerImageName     string
	DockerImageTag      string

	ShmSize                    string
	MaxCPU                     string
	MinCPU                     string
	MaxMemory                  string
	MinMemory                  string
	MaxEphemeralStorage        string
	MinEphemeralStorage        string
	KioskConfigURL             string
	KioskConfigJWTURL          string
	KioskConfigJWTOIDCClientID string
	IconURL                    string
	IconBackgroundColor        string
	StabilityStatus            AppStabilityStatus

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// AppStatus represents the status of an app.
type AppStatus string

const (
	AppActive   AppStatus = "active"
	AppInactive AppStatus = "inactive"
	AppDeleted  AppStatus = "deleted"
)

func (s AppStatus) String() string {
	return string(s)
}

func ToAppStatus(status string) (AppStatus, error) {
	switch status {
	case AppActive.String():
		return AppActive, nil
	case AppInactive.String():
		return AppInactive, nil
	case AppDeleted.String():
		return AppDeleted, nil
	default:
		return "", errors.New("unexpected AppStatus: " + status)
	}
}

// AppStabilityStatus represents the stability status of an app (technical field).
type AppStabilityStatus string

const (
	AppStabilityStatusReady AppStabilityStatus = "ready"
	AppStabilityStatusBeta  AppStabilityStatus = "beta"
	AppStabilityStatusAlpha AppStabilityStatus = "alpha"
	AppStabilityStatusOff   AppStabilityStatus = "off"
)

func (s AppStabilityStatus) String() string {
	return string(s)
}

func ToAppStabilityStatus(status string) AppStabilityStatus {
	switch status {
	case AppStabilityStatusReady.String():
		return AppStabilityStatusReady
	case AppStabilityStatusBeta.String():
		return AppStabilityStatusBeta
	case AppStabilityStatusAlpha.String():
		return AppStabilityStatusAlpha
	case AppStabilityStatusOff.String():
		return AppStabilityStatusOff
	default:
		return AppStabilityStatusOff
	}
}

func (App) IsValidSortType(sortType string) bool {
	validSortTypes := map[string]bool{
		"id":                  true,
		"name":                true,
		"status":              true,
		"dockerimageregistry": true,
		"dockerimagename":     true,
		"createdat":           true,
	}

	return validSortTypes[sortType]
}
