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

	DockerImageName string
	DockerImageTag  string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (a App) GetImage() string {
	if a.DockerImageTag == "" {
		return a.DockerImageName
	}

	return a.DockerImageName + ":" + a.DockerImageTag
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
