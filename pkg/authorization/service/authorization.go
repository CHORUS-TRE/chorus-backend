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

type DynamicRoleStore interface {
	ListDynamicRoles(ctx context.Context) ([]*model.RoleDefinition, error)
	CreateDynamicRole(ctx context.Context, role *model.RoleDefinition) error
}

type Authorizer interface {
	IsUserAllowed(user []model.Role, permission model.Permission) (bool, error)
	ExplainIsUserAllowed(user []model.Role, permission model.Permission) string
	GetUserPermissions(user []model.Role) ([]model.Permission, error)
	GetContextListForPermission(user []model.Role, permissionName model.PermissionName) ([]model.Context, error)
	FindUsersWithPermission(ctx context.Context, tenantID uint64, filter model.FindUsersWithPermissionFilter) ([]uint64, error)
	GetRolesGrantingPermission(permissionName model.PermissionName) []model.RoleName
	GetRoleDefinition(roleName model.RoleName) (*model.RoleDefinition, bool)
	IsRoleInScope(roleName model.RoleName, scopes ...model.RoleScope) bool
	CanAssignRole(user []model.Role, roleName model.RoleName, assignmentContext model.Context) (bool, error)
}

type AuthorizationServiceInterface interface {
	GetAuthorizationSchema() *model.AuthorizationSchema
	SetAuthorizationSchema(schema *model.AuthorizationSchema) error
	SetDynamicRoleStore(store DynamicRoleStore)
	LoadDynamicRoles(ctx context.Context) error
	CreateDynamicRole(ctx context.Context, user []model.Role, role *model.RoleDefinition, validationContext model.Context) (*model.RoleDefinition, error)
	GetUserPermissions(user []model.Role) ([]model.Permission, error)
	IsUserAllowed(user []model.Role, permission model.Permission) (bool, error)
	ExplainIsUserAllowed(user []model.Role, permission model.Permission) string
	GetContextListForPermission(user []model.Role, permissionName model.PermissionName) ([]model.Context, error)
	GetRoleDefinition(roleName model.RoleName) (*model.RoleDefinition, bool)
	IsRoleInScope(roleName model.RoleName, scopes ...model.RoleScope) bool
	CanAssignRole(user []model.Role, roleName model.RoleName, assignmentContext model.Context) (bool, error)
}

type authStructures struct {
	PermissionMap           map[model.PermissionName]model.PermissionDefinition
	RoleMap                 map[model.RoleName]*model.RoleDefinition
	RolesGrantingPermission map[model.PermissionName][]model.RoleName
}

type authorizationService struct {
	AuthorizationSchema *model.AuthorizationSchema
	dynamicRoleStore    DynamicRoleStore
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

func (s *authorizationService) SetDynamicRoleStore(store DynamicRoleStore) {
	s.dynamicRoleStore = store
}

func (s *authorizationService) LoadDynamicRoles(ctx context.Context) error {
	if s.dynamicRoleStore == nil {
		return nil
	}

	dynamicRoles, err := s.dynamicRoleStore.ListDynamicRoles(ctx)
	if err != nil {
		return err
	}

	schema := cloneSchemaWithoutDynamicRoles(s.AuthorizationSchema)
	schema.Roles = append(schema.Roles, dynamicRoles...)
	return s.SetAuthorizationSchema(&schema)
}

func (s *authorizationService) CreateDynamicRole(ctx context.Context, user []model.Role, role *model.RoleDefinition, validationContext model.Context) (*model.RoleDefinition, error) {
	if s.dynamicRoleStore == nil {
		return nil, fmt.Errorf("dynamic role store is nil")
	}
	if role == nil {
		return nil, fmt.Errorf("role is nil")
	}
	if model.IsSystemRole(role.Name) {
		return nil, fmt.Errorf("cannot create or update system role %q", role.Name)
	}
	if _, exists := s.RoleMap[role.Name]; exists {
		return nil, fmt.Errorf("role %q already exists", role.Name)
	}
	if role.Scope != model.RoleScopePlatform && role.Scope != model.RoleScopeWorkspace {
		return nil, fmt.Errorf("dynamic role scope must be %q or %q", model.RoleScopePlatform, model.RoleScopeWorkspace)
	}
	if len(role.Permissions) == 0 {
		return nil, fmt.Errorf("dynamic role must grant at least one permission")
	}

	normalizedRole := &model.RoleDefinition{
		Name:                      role.Name,
		Description:               role.Description,
		Scope:                     role.Scope,
		Dynamic:                   true,
		RequiredContextDimensions: requiredContextForScope(role.Scope),
		Permissions:               append([]model.PermissionName(nil), role.Permissions...),
	}

	if err := s.ensurePermissionsFitScope(normalizedRole); err != nil {
		return nil, err
	}

	if err := s.ensureUserCanGrantPermissions(user, normalizedRole.Permissions, validationContext); err != nil {
		return nil, err
	}

	if err := s.dynamicRoleStore.CreateDynamicRole(ctx, normalizedRole); err != nil {
		return nil, err
	}

	schema := cloneSchemaWithoutDynamicRoles(s.AuthorizationSchema)
	schema.Roles = append(schema.Roles, normalizedRole)
	if err := s.SetAuthorizationSchema(&schema); err != nil {
		return nil, err
	}
	return normalizedRole, nil
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

func (s *authorizationService) GetRoleDefinition(roleName model.RoleName) (*model.RoleDefinition, bool) {
	definition, ok := s.RoleMap[roleName]
	return definition, ok
}

func (s *authorizationService) IsRoleInScope(roleName model.RoleName, scopes ...model.RoleScope) bool {
	definition, ok := s.GetRoleDefinition(roleName)
	if !ok {
		return false
	}
	for _, scope := range scopes {
		if definition.Scope == scope {
			return true
		}
	}
	return false
}

func (s *authorizationService) CanAssignRole(user []model.Role, roleName model.RoleName, assignmentContext model.Context) (bool, error) {
	definition, ok := s.GetRoleDefinition(roleName)
	if !ok {
		return false, fmt.Errorf("role %q not found in schema", roleName)
	}
	if err := s.ensureUserCanGrantPermissions(user, definition.Permissions, assignmentContext); err != nil {
		return false, err
	}
	return true, nil
}

func (s *authorizationService) ensureUserCanGrantPermissions(user []model.Role, permissions []model.PermissionName, assignmentContext model.Context) error {
	seen := map[model.PermissionName]bool{}
	for _, permissionName := range permissions {
		if seen[permissionName] {
			continue
		}
		seen[permissionName] = true

		permissionDefinition, ok := s.PermissionMap[permissionName]
		if !ok {
			return fmt.Errorf("unknown permission: %s", permissionName)
		}

		permission, err := permissionForContext(permissionDefinition, assignmentContext)
		if err != nil {
			return err
		}

		allowed, err := s.IsUserAllowed(user, permission)
		if err != nil {
			return err
		}
		if !allowed {
			return fmt.Errorf("caller cannot grant permission %s", permission.String())
		}
	}
	return nil
}

func (s *authorizationService) ensurePermissionsFitScope(role *model.RoleDefinition) error {
	roleDimensions := map[model.ContextDimension]bool{}
	for dimension := range role.RequiredContextDimensions {
		roleDimensions[dimension] = true
	}

	for _, permissionName := range role.Permissions {
		permissionDefinition, ok := s.PermissionMap[permissionName]
		if !ok {
			return fmt.Errorf("unknown permission: %s", permissionName)
		}
		for _, dimension := range permissionDefinition.RequiredContextDimensions {
			if !roleDimensions[dimension] {
				return fmt.Errorf("permission %s requires context %q outside %s role scope", permissionName, dimension, role.Scope)
			}
		}
	}
	return nil
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

func (a *auth) GetRoleDefinition(roleName model.RoleName) (*model.RoleDefinition, bool) {
	return a.policy.GetRoleDefinition(roleName)
}

func (a *auth) IsRoleInScope(roleName model.RoleName, scopes ...model.RoleScope) bool {
	return a.policy.IsRoleInScope(roleName, scopes...)
}

func (a *auth) CanAssignRole(user []model.Role, roleName model.RoleName, assignmentContext model.Context) (bool, error) {
	return a.policy.CanAssignRole(user, roleName, assignmentContext)
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

func permissionForContext(permissionDefinition model.PermissionDefinition, assignmentContext model.Context) (model.Permission, error) {
	permission := model.Permission{
		Name:    permissionDefinition.Name,
		Context: make(model.Context, len(permissionDefinition.RequiredContextDimensions)),
	}
	for _, dimension := range permissionDefinition.RequiredContextDimensions {
		if value, ok := assignmentContext[dimension]; ok && value != "" {
			permission.Context[dimension] = value
		}
	}
	return permission, nil
}

func requiredContextForScope(scope model.RoleScope) map[model.ContextDimension]model.ContextQuantifier {
	switch scope {
	case model.RoleScopeWorkspace:
		return map[model.ContextDimension]model.ContextQuantifier{model.RoleContextWorkspace: model.ContextQuantifierOne}
	case model.RoleScopePlatform:
		return nil
	default:
		return nil
	}
}

func cloneSchemaWithoutDynamicRoles(schema *model.AuthorizationSchema) model.AuthorizationSchema {
	if schema == nil {
		return model.AuthorizationSchema{}
	}
	clone := model.AuthorizationSchema{
		Permissions: append([]model.PermissionDefinition(nil), schema.Permissions...),
		Roles:       make([]*model.RoleDefinition, 0, len(schema.Roles)),
	}
	for _, role := range schema.Roles {
		if role.Dynamic {
			continue
		}
		roleCopy := *role
		roleCopy.Permissions = append([]model.PermissionName(nil), role.Permissions...)
		if role.RequiredContextDimensions != nil {
			roleCopy.RequiredContextDimensions = make(map[model.ContextDimension]model.ContextQuantifier, len(role.RequiredContextDimensions))
			for dimension, quantifier := range role.RequiredContextDimensions {
				roleCopy.RequiredContextDimensions[dimension] = quantifier
			}
		}
		clone.Roles = append(clone.Roles, &roleCopy)
	}
	return clone
}
