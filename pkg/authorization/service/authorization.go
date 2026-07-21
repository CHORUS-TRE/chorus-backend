package service

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
)

// Store persists role definitions and answers user-permission lookups.
// It is the only direct collaborator of the authorization service.
type Store interface {
	SyncSystemRoles(ctx context.Context, roles []*model.RoleDefinition) error
	ListRoles(ctx context.Context) ([]*model.RoleDefinition, error)
	CreateDynamicRole(ctx context.Context, role *model.RoleDefinition) error
	FindUsersWithPermission(ctx context.Context, tenantID uint64, filter model.FindUsersWithPermissionFilter, rolesGranting []model.RoleName) ([]uint64, error)
}

type Authorizer interface {
	GetAuthorizationSchema() *model.AuthorizationSchema
	CreateDynamicRole(ctx context.Context, user []model.Role, role *model.RoleDefinition, validationContext model.Context) (*model.RoleDefinition, error)

	GetUserPermissions(user []model.Role) ([]model.Permission, error)
	IsUserAllowed(user []model.Role, permission model.Permission) (bool, error)
	ExplainIsUserAllowed(user []model.Role, permission model.Permission) string
	GetContextListForPermission(user []model.Role, permissionName model.PermissionName) ([]model.Context, error)
	GetRoleDefinition(roleName model.RoleName) (*model.RoleDefinition, bool)
	GetRolesGrantingPermission(permissionName model.PermissionName) []model.RoleName
	IsRoleInScope(roleName model.RoleName, scopes ...model.RoleScope) bool
	CanAssignRole(user []model.Role, roleName model.RoleName, assignmentContext model.Context) (bool, error)
	FindUsersWithPermission(ctx context.Context, tenantID uint64, filter model.FindUsersWithPermissionFilter) ([]uint64, error)
}

// authStructures is the schema in lookup form, rebuilt on every schema
// reload and handed to the policy kernel for decisions.
type authStructures struct {
	PermissionMap           map[model.PermissionName]model.PermissionDefinition
	RoleMap                 map[model.RoleName]*model.RoleDefinition
	RolesGrantingPermission map[model.PermissionName][]model.RoleName
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

type authorizationService struct {
	schema *model.AuthorizationSchema
	store  Store
	authStructures
}

var _ Authorizer = (*authorizationService)(nil)

// NewAuthorizationService syncs the code-defined system roles to the store,
// reloads the full role set from the store, and builds the in-memory schema.
// The store is the single source of truth for roles after construction.
func NewAuthorizationService(ctx context.Context, cfg config.Config, store Store) (Authorizer, error) {
	codeSchema := model.GetDefaultSchema()

	if cfg.Services.AuthorizationService.WorkspaceAdminCanAssignDataManager {
		for i, role := range codeSchema.Roles {
			if role.Name == model.RoleWorkspaceAdmin {
				codeSchema.Roles[i].Permissions = append(codeSchema.Roles[i].Permissions, model.PermissionManageUsersDataRoleInWorkspace)
				break
			}
		}
	}

	if store == nil {
		return nil, fmt.Errorf("authorization store is nil")
	}

	if err := store.SyncSystemRoles(ctx, codeSchema.Roles); err != nil {
		return nil, fmt.Errorf("sync system roles: %w", err)
	}

	s := &authorizationService{store: store}
	if err := s.reloadSchema(ctx, codeSchema.Permissions); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *authorizationService) reloadSchema(ctx context.Context, permissions []model.PermissionDefinition) error {
	roles, err := s.store.ListRoles(ctx)
	if err != nil {
		return fmt.Errorf("list roles: %w", err)
	}
	schema := &model.AuthorizationSchema{
		Permissions: permissions,
		Roles:       roles,
	}
	structures, err := extractAuthorizationStructures(schema)
	if err != nil {
		return err
	}
	s.schema = schema
	s.authStructures = structures
	return nil
}

func (s *authorizationService) GetAuthorizationSchema() *model.AuthorizationSchema {
	return s.schema
}

func (s *authorizationService) CreateDynamicRole(ctx context.Context, user []model.Role, role *model.RoleDefinition, validationContext model.Context) (*model.RoleDefinition, error) {
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
	if err := s.store.CreateDynamicRole(ctx, normalizedRole); err != nil {
		return nil, err
	}
	if err := s.reloadSchema(ctx, s.schema.Permissions); err != nil {
		return nil, err
	}
	return normalizedRole, nil
}

// GetUserPermissions returns the deduplicated permission names the user's
// roles grant, without context.
func (s *authorizationService) GetUserPermissions(user []model.Role) ([]model.Permission, error) {
	permissions := make([]model.Permission, 0)
	seen := map[model.PermissionName]bool{}
	for _, role := range user {
		definition, ok := s.RoleMap[role.Name]
		if !ok {
			return nil, fmt.Errorf("role %q not found in schema", role.Name)
		}
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

func (s *authorizationService) IsUserAllowed(user []model.Role, permission model.Permission) (bool, error) {
	return isUserAllowed(s.RoleMap, s.PermissionMap, user, permission)
}

func (s *authorizationService) ExplainIsUserAllowed(user []model.Role, permission model.Permission) string {
	if _, ok := s.PermissionMap[permission.Name]; !ok {
		return fmt.Sprintf("unknown permission: %s", permission)
	}
	permissions, err := expandUserPermissions(s.RoleMap, s.PermissionMap, user)
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

// GetContextListForPermission returns the distinct contexts in which the
// user holds the given permission.
func (s *authorizationService) GetContextListForPermission(user []model.Role, permissionName model.PermissionName) ([]model.Context, error) {
	if _, ok := s.PermissionMap[permissionName]; !ok {
		return nil, fmt.Errorf("unknown permission: %s", permissionName)
	}
	permissions, err := expandUserPermissions(s.RoleMap, s.PermissionMap, user)
	if err != nil {
		return nil, err
	}
	contexts := []model.Context{}
	seenContext := map[string]bool{}
	for _, permission := range permissions {
		if permission.Name != permissionName {
			continue
		}
		key := permission.Context.String()
		if seenContext[key] {
			continue
		}
		seenContext[key] = true
		contexts = append(contexts, permission.Context)
	}
	return contexts, nil
}

func (s *authorizationService) GetRoleDefinition(roleName model.RoleName) (*model.RoleDefinition, bool) {
	definition, ok := s.RoleMap[roleName]
	return definition, ok
}

func (s *authorizationService) GetRolesGrantingPermission(permissionName model.PermissionName) []model.RoleName {
	return s.RolesGrantingPermission[permissionName]
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

	// User-role managers may delegate any non-system role
	if definition.Scope != model.RoleScopeSystem && s.canManageUserRoles(user) {
		return true, nil
	}

	if err := s.ensureUserCanGrantPermissions(user, definition.Permissions, assignmentContext); err != nil {
		return false, err
	}
	return true, nil
}

func (s *authorizationService) canManageUserRoles(user []model.Role) bool {
	allowed, err := s.IsUserAllowed(user, model.Permission{Name: model.PermissionManageUserRoles})
	return err == nil && allowed
}

func (s *authorizationService) FindUsersWithPermission(ctx context.Context, tenantID uint64, filter model.FindUsersWithPermissionFilter) ([]uint64, error) {
	return s.store.FindUsersWithPermission(ctx, tenantID, filter, s.RolesGrantingPermission[filter.PermissionName])
}

// ensureUserCanGrantPermissions verifies the user holds, in the assignment
// context, every permission they are about to grant.
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

		permission := permissionForContext(permissionDefinition, assignmentContext)

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

// ensurePermissionsFitScope verifies every permission of the role only
// requires context dimensions the role scope provides.
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

// permissionForContext builds the permission check for a definition, keeping
// only the context values the permission requires.
func permissionForContext(definition model.PermissionDefinition, assignmentContext model.Context) model.Permission {
	permission := model.Permission{
		Name:    definition.Name,
		Context: make(model.Context, len(definition.RequiredContextDimensions)),
	}
	for _, dimension := range definition.RequiredContextDimensions {
		if value, ok := assignmentContext[dimension]; ok && value != "" {
			permission.Context[dimension] = value
		}
	}
	return permission
}

func requiredContextForScope(scope model.RoleScope) map[model.ContextDimension]model.ContextQuantifier {
	switch scope {
	case model.RoleScopeWorkspace:
		return map[model.ContextDimension]model.ContextQuantifier{model.ContextWorkspace: model.ContextQuantifierOne}
	default:
		return nil
	}
}
