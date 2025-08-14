package authorization

import (
	"fmt"

	"github.com/CHORUS-TRE/chorus-gatekeeper/pkg/authorization/model"
	gatekeeper_model "github.com/CHORUS-TRE/chorus-gatekeeper/pkg/authorization/model"
	gatekeeper_service "github.com/CHORUS-TRE/chorus-gatekeeper/pkg/authorization/service"
)

type Authorizer interface {
	IsUserAllowed(user []Role, permission Permission) (bool, error)
}

type auth struct {
	gatekeeper    gatekeeper_service.AuthorizationServiceInterface
	permissionMap map[Permission]model.Permission
	roleMap       map[Role]*model.Role
}

func NewAuthorizer(gatekeeper gatekeeper_service.AuthorizationServiceInterface) (Authorizer, error) {
	schema := gatekeeper.GetAuthorizationSchema()
	if schema == nil {
		return nil, fmt.Errorf("authorization schema is nil")
	}

	permissionMap := make(map[Permission]model.Permission)
	for _, gkp := range schema.Permissions {
		p, err := ToPermission(gkp.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to convert gatekeeper permission %s: %w", gkp.Name, err)
		}

		permissionMap[p] = gkp
	}

	roleMap := make(map[Role]*model.Role)
	for _, gkr := range schema.Roles {
		r, err := ToRole(gkr.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to convert gatekeeper role %s: %w", gkr.Name, err)
		}
		roleMap[r] = gkr
	}

	return &auth{
		gatekeeper:    gatekeeper,
		permissionMap: permissionMap,
		roleMap:       roleMap,
	}, nil
}

func (a *auth) IsUserAllowed(user []Role, permission Permission) (bool, error) {
	roles := make([]*gatekeeper_model.Role, len(user))
	for i, r := range user {
		role, ok := a.roleMap[r]
		if !ok {
			return false, fmt.Errorf("unknown role: %s", r)
		}
		roles[i] = role
	}

	p, ok := a.permissionMap[permission]
	if !ok {
		return false, fmt.Errorf("unknown permission: %s", permission)
	}

	return a.gatekeeper.IsUserAllowed(gatekeeper_model.User{Roles: roles}, p), nil
}
