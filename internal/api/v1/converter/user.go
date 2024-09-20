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

	var roles []string
	for _, role := range user.Roles {
		roles = append(roles, role.String())
	}

	return &chorus.User{
		Id:              user.ID,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		Username:        user.Username,
		Password:        user.Password,
		PasswordChanged: user.PasswordChanged,
		Status:          user.Status.String(),
		Roles:           roles,
		TotpEnabled:     user.TotpEnabled,
		CreatedAt:       ca,
		UpdatedAt:       ua,
	}, nil
}
