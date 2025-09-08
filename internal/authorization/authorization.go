package authorization

import (
	"fmt"

	gatekeeper_model "github.com/CHORUS-TRE/chorus-gatekeeper/pkg/authorization/model"
	gatekeeper_service "github.com/CHORUS-TRE/chorus-gatekeeper/pkg/authorization/service"
)

type Authorizer interface {
	IsUserAllowed(user []Role, permission Permission) (bool, error)
	GetUserPermissions(user []Role) ([]Permission, error)
}

type auth struct {
	gatekeeper    gatekeeper_service.AuthorizationServiceInterface
	permissionMap map[PermissionName]gatekeeper_model.Permission
	roleMap       map[RoleName]gatekeeper_model.Role
}

func NewAuthorizer(gatekeeper gatekeeper_service.AuthorizationServiceInterface) (Authorizer, error) {
	schema := gatekeeper.GetAuthorizationSchema()
	if schema == nil {
		return nil, fmt.Errorf("authorization schema is nil")
	}

	permissionMap := make(map[PermissionName]gatekeeper_model.Permission)
	for _, gkp := range schema.Permissions {
		p, err := ToPermissionName(gkp.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to convert gatekeeper permission %s: %w", gkp.Name, err)
		}

		permissionMap[p] = gkp
	}

	roleMap := make(map[RoleName]gatekeeper_model.Role)
	for _, gkr := range schema.Roles {
		r, err := ToRoleName(gkr.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to convert gatekeeper role %s: %w", gkr.Name, err)
		}
		roleMap[r] = *gkr
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
		role, ok := a.roleMap[r.Name]
		if !ok {
			return false, fmt.Errorf("unknown role: %s", r)
		}

		roleInstance := role
		roleInstance.Attributes = make(gatekeeper_model.Attributes)
		for k, v := range r.Context {
			roleInstance.Attributes[gatekeeper_model.ContextDimension(k)] = v
		}

		roles[i] = &roleInstance
	}

	p, ok := a.permissionMap[permission.Name]
	if !ok {
		return false, fmt.Errorf("unknown permission: %s", permission)
	}

	pInstance := p
	pInstance.Context = make(gatekeeper_model.Attributes)
	for k, v := range permission.Context {
		pInstance.Context[gatekeeper_model.ContextDimension(k)] = v
	}

	return a.gatekeeper.IsUserAllowed(gatekeeper_model.User{Roles: roles}, pInstance), nil
}

func (a *auth) GetUserPermissions(user []Role) ([]Permission, error) {
	roles := make([]*gatekeeper_model.Role, len(user))
	for i, r := range user {
		role, ok := a.roleMap[r.Name]
		if !ok {
			return nil, fmt.Errorf("unknown role: %s", r)
		}

		roleInstance := role
		roleInstance.Attributes = make(gatekeeper_model.Attributes)
		for k, v := range r.Context {
			roleInstance.Attributes[gatekeeper_model.ContextDimension(k)] = v
		}

		roles[i] = &roleInstance
	}

	gkPermissions := a.gatekeeper.GetUserPermissions(gatekeeper_model.User{Roles: roles})
	permissions := make([]Permission, len(gkPermissions))
	for i, p := range gkPermissions {
		cm := make(map[string]string)
		for k, v := range p.Context {
			cm[string(k)] = v
		}
		p, err := toPermission(p.Name, cm)
		if err != nil {
			return nil, fmt.Errorf("failed to convert gatekeeper permission %s: %w", p.Name, err)
		}
		permissions[i] = p
	}

	return permissions, nil
}
