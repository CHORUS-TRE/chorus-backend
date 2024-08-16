package model

import (
	"errors"
	"time"
)

// User maps an entry in the 'user' database table.
// Nullable fields have pointer types.
type User struct {
	ID       uint64
	TenantID uint64

	FirstName       string
	LastName        string
	Username        string
	Password        string
	PasswordChanged bool
	Status          UserStatus

	TotpEnabled       bool
	TotpSecret        *string
	TotpRecoveryCodes []string

	Roles []UserRole

	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserStatus string

const (
	UserActive   UserStatus = "active"
	UserDisabled UserStatus = "disabled"
	UserDeleted  UserStatus = "deleted"
)

func (s UserStatus) String() string {
	return string(s)
}

func ToUserStatus(status string) (UserStatus, error) {
	switch status {
	case UserActive.String():
		return UserActive, nil
	case UserDisabled.String():
		return UserDisabled, nil
	case UserDeleted.String():
		return UserDeleted, nil
	default:
		return "", errors.New("unexpected UserStatus: " + status)
	}
}
