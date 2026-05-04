package authorization

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	"go.uber.org/zap"
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

type AuthorizationServiceInterface interface {
	GetAuthorizationSchema() *model.AuthorizationSchema
	SetAuthorizationSchema(schema *model.AuthorizationSchema) error
	GetUserPermissions(user []model.Role) ([]model.Permission, error)
	IsUserAllowed(user []model.Role, permission model.Permission) (bool, error)
	ExplainIsUserAllowed(user []model.Role, permission model.Permission) string
	GetContextListForPermission(user []model.Role, permissionName model.PermissionName) ([]model.Context, error)
}

type authStructures struct {
	PermissionMap           map[model.PermissionName]model.PermissionDefinition
	RoleMap                 map[model.RoleName]*model.RoleDefinition
	RolesGrantingPermission map[model.PermissionName][]model.RoleName
}

type authorizationService struct {
	AuthorizationSchema *model.AuthorizationSchema
	authStructures
}

type auth struct {
	policy              AuthorizationServiceInterface
	userPermissionStore UserPermissionStore
	authStructures
}

var _ AuthorizationServiceInterface = (*authorizationService)(nil)

func NewAuthorizationService(schema *model.AuthorizationSchema) (AuthorizationServiceInterface, error) {
	service := &authorizationService{}
	if err := service.SetAuthorizationSchema(schema); err != nil {
		return nil, err
	}
	return service, nil
}

func ExtractAuthorizationStructures(policy AuthorizationServiceInterface) (authStructures, error) {
	schema := policy.GetAuthorizationSchema()
	if schema == nil {
		return authStructures{}, fmt.Errorf("authorization schema is nil")
	}
	return extractAuthorizationStructures(schema)
}

func NewAuthorizer(policy AuthorizationServiceInterface, userPermissionStore UserPermissionStore) (Authorizer, error) {
	authStructures, err := ExtractAuthorizationStructures(policy)
	if err != nil {
		return nil, fmt.Errorf("failed to extract authorization structures: %w", err)
	}
	return &auth{
		policy:              policy,
		userPermissionStore: userPermissionStore,
		authStructures:      authStructures,
	}, nil
}

func (s *authorizationService) GetAuthorizationSchema() *model.AuthorizationSchema {
	return s.AuthorizationSchema
}

func (s *authorizationService) SetAuthorizationSchema(schema *model.AuthorizationSchema) error {
	if schema == nil {
		return fmt.Errorf("authorization schema is nil")
	}

	structures, err := extractAuthorizationStructures(schema)
	if err != nil {
		return err
	}

	s.AuthorizationSchema = schema
	s.authStructures = structures

	return nil
}

func (s *authorizationService) GetUserPermissions(user []model.Role) ([]model.Permission, error) {
	permissions := make([]model.Permission, 0)
	seen := map[model.PermissionName]bool{}
	for _, role := range user {
		if _, ok := s.RoleMap[role.Name]; !ok {
			return nil, fmt.Errorf("role %q not found in schema", role.Name)
		}
		definition := s.RoleMap[role.Name]
		for _, permissionName := range definition.Permissions {
			if seen[permissionName] {
				continue
			}
			seen[permissionName] = true
			permissions = append(permissions, model.Permission{Name: permissionName})
		}
	}
	return permissions, nil
}

func (s *authorizationService) getUserPermissionsWithContext(user []model.Role) ([]model.Permission, error) {
	permissions := make([]model.Permission, 0)
	for _, role := range user {
		if _, ok := s.RoleMap[role.Name]; !ok {
			return nil, fmt.Errorf("role %q not found in schema", role.Name)
		}
		definition := s.RoleMap[role.Name]
		for _, permissionName := range definition.Permissions {
			permissionDefinition := s.PermissionMap[permissionName]
			permission := model.Permission{
				Name:    permissionName,
				Context: make(model.Context, len(permissionDefinition.RequiredContextDimensions)),
			}
			for _, dimension := range permissionDefinition.RequiredContextDimensions {
				if actualValue, ok := role.Context[dimension]; ok {
					permission.Context[dimension] = actualValue
				}
			}
			permissions = append(permissions, permission)
		}
	}
	return permissions, nil
}

func (s *authorizationService) IsUserAllowed(user []model.Role, permission model.Permission) (bool, error) {
	permissions, err := s.getUserPermissionsWithContext(user)
	if err != nil {
		return false, err
	}

	for _, userPermission := range permissions {
		if isPermissionIdentical(userPermission, permission) {
			return true, nil
		}
	}
	return false, nil
}

func (s *authorizationService) ExplainIsUserAllowed(user []model.Role, permission model.Permission) string {
	permissions, err := s.getUserPermissionsWithContext(user)
	if err != nil {
		return fmt.Sprintf("error expanding user roles: %v", err)
	}

	explanations := ""
	for _, userPermission := range permissions {
		identical, explanation := explainIsPermissionIdentical(userPermission, permission)
		if identical {
			return explanation
		}
		explanations += explanation + "\n"
	}
	return explanations
}

func (s *authorizationService) GetContextListForPermission(user []model.Role, permissionName model.PermissionName) ([]model.Context, error) {
	if _, ok := s.PermissionMap[permissionName]; !ok {
		return nil, fmt.Errorf("unknown permission: %s", permissionName)
	}

	permissions, err := s.getUserPermissionsWithContext(user)
	if err != nil {
		return nil, err
	}

	contexts := []model.Context{}
	seenContext := map[string]bool{}
	for _, permission := range permissions {
		if permission.Name != permissionName {
			continue
		}
		contextKey := permission.Context.String()
		if seenContext[contextKey] {
			continue
		}
		seenContext[contextKey] = true
		contexts = append(contexts, permission.Context)
	}
	return contexts, nil
}

func (a *auth) IsUserAllowed(user []model.Role, permission model.Permission) (bool, error) {
	if _, ok := a.PermissionMap[permission.Name]; !ok {
		return false, fmt.Errorf("unknown permission: %s", permission)
	}

	allowed, err := a.policy.IsUserAllowed(user, permission)
	if err != nil {
		return false, err
	}

	if !allowed {
		logger.TechLog.Info(context.Background(), "no role grants the required permission",
			zap.String("required_permission", string(permission.Name)),
			zap.Any("roles_granting_permission", a.RolesGrantingPermission[permission.Name]),
		)
	}

	return allowed, nil
}

func (a *auth) ExplainIsUserAllowed(user []model.Role, permission model.Permission) string {
	if _, ok := a.PermissionMap[permission.Name]; !ok {
		return fmt.Sprintf("unknown permission: %s", permission)
	}
	return a.policy.ExplainIsUserAllowed(user, permission)
}

func (a *auth) GetUserPermissions(user []model.Role) ([]model.Permission, error) {
	return a.policy.GetUserPermissions(user)
}

func (a *auth) GetContextListForPermission(user []model.Role, permissionName model.PermissionName) ([]model.Context, error) {
	return a.policy.GetContextListForPermission(user, permissionName)
}

func (a *auth) FindUsersWithPermission(ctx context.Context, tenantID uint64, filter model.FindUsersWithPermissionFilter) ([]uint64, error) {
	return a.userPermissionStore.FindUsersWithPermission(ctx, tenantID, filter)
}

func (a *auth) GetRolesGrantingPermission(permissionName model.PermissionName) []model.RoleName {
	return a.RolesGrantingPermission[permissionName]
}

func extractAuthorizationStructures(schema *model.AuthorizationSchema) (authStructures, error) {
	permissionMap := make(map[model.PermissionName]model.PermissionDefinition, len(schema.Permissions))
	for _, permission := range schema.Permissions {
		if _, ok := permissionMap[permission.Name]; ok {
			return authStructures{}, fmt.Errorf("duplicate permission name %s", permission.Name)
		}
		permissionMap[permission.Name] = permission
	}

	roleMap := make(map[model.RoleName]*model.RoleDefinition, len(schema.Roles))
	for _, role := range schema.Roles {
		if _, ok := roleMap[role.Name]; ok {
			return authStructures{}, fmt.Errorf("duplicate role name %s", role.Name)
		}
		roleMap[role.Name] = role
	}

	rolesGrantingPermission := make(map[model.PermissionName][]model.RoleName)
	for _, role := range schema.Roles {
		for _, permission := range role.Permissions {
			if _, ok := permissionMap[permission]; !ok {
				return authStructures{}, fmt.Errorf("role %s has unknown permission %s", role.Name, permission)
			}
			rolesGrantingPermission[permission] = append(rolesGrantingPermission[permission], role.Name)
		}
	}

	return authStructures{
		PermissionMap:           permissionMap,
		RoleMap:                 roleMap,
		RolesGrantingPermission: rolesGrantingPermission,
	}, nil
}

func isPermissionIdentical(userPermission, permission model.Permission) bool {
	if userPermission.Name != permission.Name {
		return false
	}

	for dimension, value := range userPermission.Context {
		if value != model.Wildcard && value != permission.Context[dimension] {
			return false
		}
	}
	return true
}

func explainIsPermissionIdentical(userPermission, permission model.Permission) (bool, string) {
	format := func(res bool) (bool, string) {
		comparison := "=="
		if !res {
			comparison = "!="
		}
		return res, fmt.Sprintf("%s %s %s", userPermission.String(), comparison, permission.String())
	}

	if userPermission.Name != permission.Name {
		return format(false)
	}

	for dimension, value := range userPermission.Context {
		if value != model.Wildcard && value != permission.Context[dimension] {
			return format(false)
		}
	}
	return format(true)
}
