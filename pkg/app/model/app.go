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
	PrettyName  string
	Description string
	Status      AppStatus

	DockerImageRegistry string
	DockerImageName     string
	DockerImageTag      string

	ShmSize             string
	MaxCPU              string
	MinCPU              string
	MaxMemory           string
	MinMemory           string
	MaxEphemeralStorage string
	MinEphemeralStorage string
	KioskConfigURL      string
	IconURL             string

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

func (App) IsValidSortType(sortType string) bool {
	validSortTypes := map[string]bool{
		"id":                  true,
		"name":                true,
		"prettyname":          true,
		"status":              true,
		"dockerimageregistry": true,
		"dockerimagename":     true,
		"createdat":           true,
	}

	return validSortTypes[sortType]
}
