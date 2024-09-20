package model

import "fmt"

// Role maps an entry in the 'roles' database table.
type Role struct {
	ID   uint64
	Name string
}

type UserRole string

const (
	RoleAdmin          UserRole = "admin"
	RoleAuthenticated  UserRole = "authenticated"
	RoleChorus         UserRole = "chorus"
	RoleFileUploader   UserRole = "fileuploader"
	RoleFileDownloader UserRole = "filedownloader"
)

var ValidRoles = map[UserRole]struct{}{RoleAdmin: {}, RoleAuthenticated: {}}

func (r UserRole) String() string {
	return string(r)
}

func ToUserRole(role string) (UserRole, error) {
	switch role {
	case RoleAdmin.String():
		return RoleAdmin, nil
	case RoleAuthenticated.String():
		return RoleAuthenticated, nil
	case RoleChorus.String():
		return RoleChorus, nil
	default:
		return "", fmt.Errorf("unexpected role %v", role)
	}
}

func ToUserRoles(roles []string) ([]UserRole, error) {
	var userRoles []UserRole
	for _, role := range roles {
		userRole, err := ToUserRole(role)
		if err != nil {
			return nil, err
		}
		userRoles = append(userRoles, userRole)
	}
	return userRoles, nil
}
