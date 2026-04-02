package model

import (
	"errors"
	"fmt"
	"time"

	authorization_model "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
)

// User maps an entry in the 'user' database table.
// Nullable fields have pointer types.
type User struct {
	ID       uint64
	TenantID uint64

	FirstName       string
	LastName        string
	Username        string
	Email           string
	Source          string
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

type UserRole struct {
	authorization_model.Role
	ID uint64
}

func (u User) Namespaces() []string {
	seen := make(map[string]struct{})
	var namespaces []string
	for _, r := range u.Roles {
		if wsID, ok := r.Context[authorization_model.RoleContextWorkspace]; ok && wsID != "" {
			ns := fmt.Sprintf("workspace%s", wsID)
			if _, exists := seen[ns]; !exists {
				seen[ns] = struct{}{}
				namespaces = append(namespaces, ns)
			}
		}
	}
	return namespaces
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

func (User) IsValidSortType(sortType string) bool {
	validSortTypes := map[string]bool{
		"id":        true,
		"firstname": true,
		"lastname":  true,
		"username":  true,
		"email":     true,
		"status":    true,
		"createdat": true,
	}

	return validSortTypes[sortType]
}
