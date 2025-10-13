package converter

import (
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
)

func UserFromBusiness(user *model.User) (*chorus.User, error) {
	ca, err := ToProtoTimestamp(user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert createdAt timestamp: %w", err)
	}
	ua, err := ToProtoTimestamp(user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("unable to convert updatedAt timestamp: %w", err)
	}

	roles := make([]*chorus.Role, len(user.Roles))
	rs := make([]string, len(user.Roles))
	for i, r := range user.Roles {
		role := RoleFromBusiness(r)
		roles[i] = &role
		rs[i] = role.Name
	}

	return &chorus.User{
		Id:              user.ID,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		Username:        user.Username,
		Source:          user.Source,
		Password:        user.Password,
		PasswordChanged: user.PasswordChanged,
		Status:          user.Status.String(),
		// Roles:           rs,
		RolesWithContext: roles,
		TotpEnabled:      user.TotpEnabled,
		CreatedAt:        ca,
		UpdatedAt:        ua,
	}, nil
}

func RoleFromBusiness(role model.UserRole) chorus.Role {
	c := make(map[string]string, len(role.Context))
	for k, v := range role.Context {
		c[k.String()] = v
	}
	return chorus.Role{
		Id:      role.ID,
		Name:    role.Name.String(),
		Context: c,
	}
}
