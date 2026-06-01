package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"

	val "github.com/go-playground/validator/v10"
)

var _ service.Authorizer = (*validation)(nil)

type validation struct {
	next     service.Authorizer
	validate *val.Validate
}

func Validation(validate *val.Validate) func(service.Authorizer) service.Authorizer {
	return func(next service.Authorizer) service.Authorizer {
		return &validation{
			next:     next,
			validate: validate,
		}
	}
}

func (v validation) GetAuthorizationSchema() *model.AuthorizationSchema {
	return v.next.GetAuthorizationSchema()
}

func (v validation) CreateDynamicRole(ctx context.Context, user []model.Role, role *model.RoleDefinition, validationContext model.Context) (*model.RoleDefinition, error) {
	if err := v.validate.Struct(role); err != nil {
		return nil, err
	}
	return v.next.CreateDynamicRole(ctx, user, role, validationContext)
}

func (v validation) IsUserAllowed(user []model.Role, permission model.Permission) (bool, error) {
	return v.next.IsUserAllowed(user, permission)
}

func (v validation) ExplainIsUserAllowed(user []model.Role, permission model.Permission) string {
	return v.next.ExplainIsUserAllowed(user, permission)
}

func (v validation) GetUserPermissions(user []model.Role) ([]model.Permission, error) {
	return v.next.GetUserPermissions(user)
}

func (v validation) GetContextListForPermission(user []model.Role, permissionName model.PermissionName) ([]model.Context, error) {
	return v.next.GetContextListForPermission(user, permissionName)
}

func (v validation) FindUsersWithPermission(ctx context.Context, tenantID uint64, filter model.FindUsersWithPermissionFilter) ([]uint64, error) {
	return v.next.FindUsersWithPermission(ctx, tenantID, filter)
}

func (v validation) GetRolesGrantingPermission(permissionName model.PermissionName) []model.RoleName {
	return v.next.GetRolesGrantingPermission(permissionName)
}

func (v validation) GetRoleDefinition(roleName model.RoleName) (*model.RoleDefinition, bool) {
	return v.next.GetRoleDefinition(roleName)
}

func (v validation) IsRoleInScope(roleName model.RoleName, scopes ...model.RoleScope) bool {
	return v.next.IsRoleInScope(roleName, scopes...)
}

func (v validation) CanAssignRole(user []model.Role, roleName model.RoleName, assignmentContext model.Context) (bool, error) {
	return v.next.CanAssignRole(user, roleName, assignmentContext)
}
