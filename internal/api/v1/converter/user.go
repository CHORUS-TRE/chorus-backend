package converter

import (
	"fmt"
	"strconv"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	authorization_model "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
)

func UserFromBusiness(user *model.User, uidOffset, gidOffset uint64) (*chorus.User, error) {
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
		role := UserRoleFromBusiness(r)
		roles[i] = &role
		rs[i] = role.Name
	}

	seen := make(map[uint64]struct{})
	var gids []uint64
	for _, r := range user.Roles {
		wsIDStr, ok := r.Context[authorization_model.RoleContextWorkspace]
		if !ok || wsIDStr == "" {
			continue
		}
		wsID, err := strconv.ParseUint(wsIDStr, 10, 64)
		if err != nil {
			continue
		}
		gid := wsID + gidOffset
		if _, exists := seen[gid]; exists {
			continue
		}
		seen[gid] = struct{}{}
		gids = append(gids, gid)
	}

	return &chorus.User{
		Id:              user.ID,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		Username:        user.Username,
		Email:           user.Email,
		Source:          user.Source,
		Password:        user.Password,
		PasswordChanged: user.PasswordChanged,
		Status:          user.Status.String(),
		// Roles:           rs,
		RolesWithContext: roles,
		TotpEnabled:      user.TotpEnabled,
		Namespaces:       user.Namespaces(),
		CreatedAt:        ca,
		UpdatedAt:        ua,
		Uid:              user.ID + uidOffset,
		Gids:             gids,
	}, nil
}

func UserRoleFromBusiness(role model.UserRole) chorus.Role {
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
