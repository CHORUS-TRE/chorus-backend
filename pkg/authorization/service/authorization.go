package authorization

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	gatekeeper_model "github.com/CHORUS-TRE/chorus-gatekeeper/pkg/authorization/model"
	gatekeeper_service "github.com/CHORUS-TRE/chorus-gatekeeper/pkg/authorization/service"
)

type UserRoleStore interface {
	GetRoles(ctx context.Context) ([]*user_model.Role, error)
}

type UserPermissionStore interface {
	FindUsersWithPermission(ctx context.Context, tenantID uint64, filter model.FindUsersWithPermissionFilter) ([]uint64, error)
}

type Authorizer interface {
	IsUserAllowed(user []model.Role, permission model.Permission) (bool, error)
	ExplainIsUserAllowed(user []model.Role, permission model.Permission) string
	GetUserPermissions(user []model.Role) ([]model.Permission, error)
	GetContextListForPermission(user []model.Role, permissionName model.PermissionName) ([]model.Context, error)
	FindUsersWithPermission(ctx context.Context, tenantID uint64, filter model.FindUsersWithPermissionFilter) ([]uint64, error)
	GetRolesGrantingPermission(permissionName model.PermissionName) []model.RoleName
}

type authStructures struct {
	PermissionMap           map[model.PermissionName]gatekeeper_model.Permission
	RoleMap                 map[model.RoleName]gatekeeper_model.Role
	RolesGrantingPermission map[model.PermissionName][]model.RoleName
}

type auth struct {
	gatekeeper          gatekeeper_service.AuthorizationServiceInterface
	userPermissionStore UserPermissionStore
	authStructures
}

func ExtractAuthoizationStructures(gatekeeper gatekeeper_service.AuthorizationServiceInterface) (authStructures, error) {
	schema := gatekeeper.GetAuthorizationSchema()
	if schema == nil {
		return authStructures{}, fmt.Errorf("authorization schema is nil")
	}

	permissionMap := make(map[model.PermissionName]gatekeeper_model.Permission)
	for _, gkp := range schema.Permissions {
		p, err := model.ToPermissionName(gkp.Name)
		if err != nil {
			return authStructures{}, fmt.Errorf("failed to convert gatekeeper permission %s: %w", gkp.Name, err)
		}

		permissionMap[p] = gkp
	}

	roleMap := make(map[model.RoleName]gatekeeper_model.Role)
	rolesGrantingPermission := make(map[model.PermissionName][]model.RoleName)
	for _, gkr := range schema.Roles {
		r, err := model.ToRoleName(gkr.Name)
		if err != nil {
			return authStructures{}, fmt.Errorf("failed to convert gatekeeper role %s: %w", gkr.Name, err)
		}
		roleMap[r] = *gkr

		for _, perm := range gkr.Permissions {
			permName, err := model.ToPermissionName(perm.Name)
			if err != nil {
				continue
			}
			rolesGrantingPermission[permName] = append(rolesGrantingPermission[permName], r)
		}
	}

	return authStructures{
		PermissionMap:           permissionMap,
		RoleMap:                 roleMap,
		RolesGrantingPermission: rolesGrantingPermission,
	}, nil
}

func NewAuthorizer(gatekeeper gatekeeper_service.AuthorizationServiceInterface, userPermissionStore UserPermissionStore) (Authorizer, error) {
	authStructures, err := ExtractAuthoizationStructures(gatekeeper)
	if err != nil {
		return nil, fmt.Errorf("failed to extract authorization structures: %w", err)
	}
	return &auth{
		gatekeeper:          gatekeeper,
		userPermissionStore: userPermissionStore,
		authStructures:      authStructures,
	}, nil
}

func (a *auth) IsUserAllowed(user []model.Role, permission model.Permission) (bool, error) {
	roles, err := a.userToGatekeeperRoles(user)
	if err != nil {
		return false, fmt.Errorf("failed to convert user roles to gatekeeper roles: %w", err)
	}

	p, ok := a.PermissionMap[permission.Name]
	if !ok {
		return false, fmt.Errorf("unknown permission: %s", permission)
	}

	pInstance := p
	pInstance.Context = make(gatekeeper_model.Attributes)
	for k, v := range permission.Context {
		pInstance.Context[gatekeeper_model.ContextDimension(k)] = v
	}

	allowed := a.gatekeeper.IsUserAllowed(gatekeeper_model.User{Roles: roles}, pInstance)

	if !allowed {
		explanation := a.gatekeeper.ExplainIsUserAllowed(gatekeeper_model.User{Roles: roles}, pInstance)
		fmt.Println("explanation:", explanation)
	}

	return allowed, nil
}

func (a *auth) ExplainIsUserAllowed(user []model.Role, permission model.Permission) string {
	roles, err := a.userToGatekeeperRoles(user)
	if err != nil {
		return fmt.Sprintf("failed to convert user roles to gatekeeper roles: %v", err)
	}

	p, ok := a.PermissionMap[permission.Name]
	if !ok {
		return fmt.Sprintf("unknown permission: %s", permission)
	}

	pInstance := p
	pInstance.Context = make(gatekeeper_model.Attributes)
	for k, v := range permission.Context {
		pInstance.Context[gatekeeper_model.ContextDimension(k)] = v
	}

	return a.gatekeeper.ExplainIsUserAllowed(gatekeeper_model.User{Roles: roles}, pInstance)
}

func (a *auth) userToGatekeeperRoles(user []model.Role) ([]*gatekeeper_model.Role, error) {
	roles := make([]*gatekeeper_model.Role, len(user))
	for i, r := range user {
		role, ok := a.RoleMap[r.Name]
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
	return roles, nil
}

func (a *auth) GetUserPermissions(user []model.Role) ([]model.Permission, error) {
	roles, err := a.userToGatekeeperRoles(user)
	if err != nil {
		return nil, fmt.Errorf("failed to convert user roles to gatekeeper roles: %w", err)
	}

	gkPermissions := a.gatekeeper.GetUserPermissions(gatekeeper_model.User{Roles: roles})
	permissions := make([]model.Permission, len(gkPermissions))
	for i, p := range gkPermissions {
		cm := make(map[string]string)
		for k, v := range p.Context {
			cm[string(k)] = v
		}
		p, err := model.ToPermission(p.Name, cm)
		if err != nil {
			return nil, fmt.Errorf("failed to convert gatekeeper permission %s: %w", p.Name, err)
		}
		permissions[i] = p
	}

	return permissions, nil
}

func (a *auth) GetContextListForPermission(user []model.Role, permissionName model.PermissionName) ([]model.Context, error) {
	roles, err := a.userToGatekeeperRoles(user)
	if err != nil {
		return nil, fmt.Errorf("failed to convert user roles to gatekeeper roles: %w", err)
	}

	_, ok := a.PermissionMap[permissionName]
	if !ok {
		return nil, fmt.Errorf("unknown permission: %s", permissionName)
	}

	contextList := a.gatekeeper.GetContextListForPermission(gatekeeper_model.User{Roles: roles}, string(permissionName))

	result := make([]model.Context, len(contextList))
	for i, c := range contextList {
		cm := make(map[model.ContextDimension]string)
		for k, v := range c {
			contextDim, err := model.ToRoleContext(string(k))
			if err != nil {
				return nil, fmt.Errorf("failed to convert context dimension %s: %w", k, err)
			}
			cm[contextDim] = v
		}
		result[i] = cm
	}

	return result, nil
}

func (a *auth) FindUsersWithPermission(ctx context.Context, tenantID uint64, filter model.FindUsersWithPermissionFilter) ([]uint64, error) {
	return a.userPermissionStore.FindUsersWithPermission(ctx, tenantID, filter)
}

func (a *auth) GetRolesGrantingPermission(permissionName model.PermissionName) []model.RoleName {
	return a.RolesGrantingPermission[permissionName]
}
