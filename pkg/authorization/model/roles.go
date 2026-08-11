package model

import "fmt"

// roleDefinitions is the single source of every role's definition, populated by
// the role declarations below.
var roleDefinitions []*RoleDefinition

func registerRole(name RoleName, description string, scope RoleScope, required map[ContextDimension]ContextQuantifier, permissions []PermissionName) *RoleDefinition {
	def := &RoleDefinition{
		Name:                      name,
		Description:               description,
		Scope:                     scope,
		RequiredContextDimensions: required,
		Permissions:               append([]PermissionName(nil), permissions...),
	}
	roleDefinitions = append(roleDefinitions, def)
	return def
}

// RoleDefinitions returns every declared role definition.
func RoleDefinitions() []*RoleDefinition { return roleDefinitions }

// RoleDefinitionsMap returns every declared role definition as a map keyed by role name.
func RoleDefinitionsMap() map[RoleName]*RoleDefinition {
	m := make(map[RoleName]*RoleDefinition, len(roleDefinitions))
	for _, def := range roleDefinitions {
		if _, exists := m[def.Name]; exists {
			panic(fmt.Sprintf("duplicate role definition: %s", def.Name))
		}
		m[def.Name] = def
	}
	return m
}

// -------------------------------------------------------------------
// Role factories
// -------------------------------------------------------------------

// A role factory embeds its definition (so a child role can inherit a parent's
// permissions via `.Permissions`) and bakes in wildcard dimensions, exposing a
// typed .For for the exact-bound value. Declare with roleWithNoContext /
// roleWithOneContext.

// roleFactoryNoContext is the factory for a role that binds no exact context
// value. It may still bake in wildcard dimensions (granted for any value).
type roleFactoryNoContext struct {
	*RoleDefinition
	wildcards []ContextDimension
}

func (r roleFactoryNoContext) For() Role {
	ctx := make(Context, len(r.wildcards))
	for _, dim := range r.wildcards {
		ctx[dim] = Wildcard
	}
	return Role{Name: r.Name, Context: ctx}
}

// roleFactoryOneContext is the factory for a role that binds one exact context value.
type roleFactoryOneContext[A contextID] struct {
	*RoleDefinition
	wildcards []ContextDimension
}

func (r roleFactoryOneContext[A]) For(a A) Role {
	ctx := make(Context, len(r.wildcards)+1)
	for _, dim := range r.wildcards {
		ctx[dim] = Wildcard
	}
	ctx[a.dimension()] = fmt.Sprint(a)
	return Role{Name: r.Name, Context: ctx}
}

func requiredContext(exact, wildcards []ContextDimension) map[ContextDimension]ContextQuantifier {
	if len(exact) == 0 && len(wildcards) == 0 {
		return nil
	}
	required := make(map[ContextDimension]ContextQuantifier, len(exact)+len(wildcards))
	for _, dim := range exact {
		required[dim] = ContextQuantifierOne
	}
	for _, dim := range wildcards {
		required[dim] = ContextQuantifierAny
	}
	return required
}

func roleWithNoContext(name RoleName, description string, scope RoleScope, permissions []PermissionName, wildcards ...ContextDimension) roleFactoryNoContext {
	return roleFactoryNoContext{registerRole(name, description, scope, requiredContext(nil, wildcards), permissions), wildcards}
}

func roleWithOneContext[A contextID](name RoleName, description string, scope RoleScope, permissions []PermissionName, wildcards ...ContextDimension) roleFactoryOneContext[A] {
	var a A
	return roleFactoryOneContext[A]{registerRole(name, description, scope, requiredContext([]ContextDimension{a.dimension()}, wildcards), permissions), wildcards}
}

// -------------------------------------------------------------------
// Permission grants
// -------------------------------------------------------------------

// grant lists the names of the given permission factories.
func grant(perms ...permFactory) []PermissionName {
	out := make([]PermissionName, len(perms))
	for i, p := range perms {
		out[i] = p.name()
	}
	return out
}

// merge unions permission-name groups.
func merge(groups ...[]PermissionName) []PermissionName {
	result := make([]PermissionName, 0)
	seen := make(map[PermissionName]bool)
	for _, group := range groups {
		for _, name := range group {
			if seen[name] {
				continue
			}
			seen[name] = true
			result = append(result, name)
		}
	}
	return result
}

// -------------------------------------------------------------------
// Roles
// -------------------------------------------------------------------

// One factory per role — name, description, scope, granted permissions, and the
// context binding (exact via the id type parameter, wildcard via trailing args).
// Each role's permissions live here: a role inherits its parent's set via
// `Parent.Permissions` and adds its own with `grant(...)`.
var (
	RolePublic = roleWithNoContext(
		"Public",
		"Public users can authenticate and read public platform settings",
		RoleScopePlatform,
		grant(
			PermAuthenticate,
			PermGetListOfPossibleWayToAuthenticate,
			PermAuthenticateUsingAuth2_0,
			PermAuthenticateRedirectUsingAuth2_0,
			PermGetPlatformSettings,
		),
	)
	RoleAuthenticated = roleWithOneContext[UserID](
		"Authenticated",
		"Authenticated users can manage their own session, profile, notifications, and base resources",
		RoleScopePlatform,
		merge(RolePublic.Permissions, grant(
			PermListNotifications,
			PermCountUnreadNotifications,
			PermMarkNotificationAsRead,
			PermLogout,
			PermRefreshToken,
			PermUpdateUser,
			PermGetMyOwnUser,
			PermUpdatePassword,
			PermEnableTotp,
			PermResetTotp,
			PermResetPassword,
			PermListWorkspaces,
			PermListPublicWorkspaces,
			PermListWorkbenchs,
			PermListApps,
			PermListAppInstances,
			PermListMyRequests,
			PermAuditUser,
			PermGetCurrentTermsOfUseVersion,
			PermGetMyTermsOfUseStatus,
			PermAcceptTermsOfUse,
			PermListOrganizations,
			PermGetOrganization,
		)),
	)
	RoleWorkspaceGuest = roleWithOneContext[WorkspaceID](
		"WorkspaceGuest",
		"Workspace guests can view workspace metadata and create requests",
		RoleScopeWorkspace,
		merge(RoleAuthenticated.Permissions, grant(
			PermListWorkspaces,
			PermGetWorkspace,
			PermListUsers,
			PermCreateRequest,
			PermListWorkspaceServiceInstances,
		)),
	)
	RoleWorkspaceMember = roleWithOneContext[WorkspaceID](
		"WorkspaceMember",
		"Workspace members can create workbenches and list workspace files",
		RoleScopeWorkspace,
		merge(RoleWorkspaceGuest.Permissions, grant(
			PermCreateWorkbench,
			PermListFilesInWorkspace,
			PermCreateRequest,
			PermGetWorkspaceServiceInstance,
			PermGetWorkspaceServiceInstanceSecret,
		)),
	)
	RoleWorkspaceMaintainer = roleWithOneContext[WorkspaceID](
		"WorkspaceMaintainer",
		"Workspace maintainers can update workspace metadata and manage workspace files",
		RoleScopeWorkspace,
		merge(RoleWorkspaceMember.Permissions, grant(
			PermUpdateWorkspace,
			PermUploadFilesToWorkspace,
			PermModifyFilesInWorkspace,
			PermSearchUsers,
			PermCreateRequest,
		)),
	)
	RoleWorkspaceDataManager = roleWithOneContext[WorkspaceID](
		"WorkspaceDataManager",
		"Workspace data managers can manage workspace files and data-manager assignments",
		RoleScopeWorkspace,
		merge(RoleWorkspaceMember.Permissions, grant(
			PermUploadFilesToWorkspace,
			PermModifyFilesInWorkspace,
			PermDownloadFilesFromWorkspace,
			PermManageUsersDataRoleInWorkspace,
			PermCreateRequest,
			PermListRequests,
		)),
	)
	RoleWorkspaceAdmin = roleWithOneContext[WorkspaceID](
		"WorkspaceAdmin",
		"Workspace admins can administer workspace users, requests, workbenches, files, and services",
		RoleScopeWorkspace,
		merge(RoleWorkspaceMaintainer.Permissions, grant(
			PermListAppInstances,
			PermListWorkbenchs,
			PermUpdateWorkbench,
			PermGetWorkbench,
			PermStreamWorkbench,
			PermDeleteWorkbench,
			PermAuditWorkbench,
			PermManageUsersInWorkbench,
			PermDeleteWorkspace,
			PermAuditWorkspace,
			PermManageUsersInWorkspace,
			PermListRequests,
			PermGetRequest,
			PermApproveRequest,
			PermDeleteRequest,
			PermCreateWorkspaceServiceInstance,
			PermUpdateWorkspaceServiceInstance,
			PermDeleteWorkspaceServiceInstance,
		)),
	)
	RoleWorkbenchViewer = roleWithOneContext[WorkbenchID](
		"WorkbenchViewer",
		"Workbench viewers can view and stream workbenches",
		RoleScopeWorkbench,
		merge(RoleAuthenticated.Permissions, grant(
			PermListAppInstances,
			PermListWorkbenchs,
			PermGetWorkbench,
			PermStreamWorkbench,
			PermListUsers,
		)),
	)
	RoleWorkbenchMember = roleWithOneContext[WorkbenchID](
		"WorkbenchMember",
		"Workbench members can update workbenches and manage app instances",
		RoleScopeWorkbench,
		merge(RoleWorkbenchViewer.Permissions, grant(
			PermCreateAppInstance,
			PermUpdateAppInstance,
			PermGetAppInstance,
			PermDeleteAppInstance,
			PermUpdateWorkbench,
		)),
	)
	RoleWorkbenchAdmin = roleWithOneContext[WorkbenchID](
		"WorkbenchAdmin",
		"Workbench admins can administer workbenches and their users",
		RoleScopeWorkbench,
		merge(RoleWorkbenchMember.Permissions, grant(
			PermDeleteWorkbench,
			PermManageUsersInWorkbench,
			PermSearchUsers,
			PermAuditWorkbench,
		)),
	)
	RoleHealthchecker = roleWithNoContext(
		"Healthchecker",
		"Healthcheckers can read healthcheck status",
		RoleScopePlatform,
		grant(
			PermGetHealthCheck,
		),
		ContextUser,
	)
	RolePlatformSettingsManager = roleWithNoContext(
		"PlatformSettingsManager",
		"Platform settings managers can manage platform settings",
		RoleScopePlatform,
		merge(RoleAuthenticated.Permissions, grant(
			PermSetPlatformSettings,
			PermListTermsOfUseVersions,
			PermGetTermsOfUseVersion,
			PermCreateTermsOfUseVersion,
			PermUpdateTermsOfUseVersion,
			PermPublishTermsOfUseVersion,
		)),
		ContextUser,
	)
	RolePlatformUserManager = roleWithNoContext(
		"PlatformUserManager",
		"Platform user managers can administer platform users and their roles",
		RoleScopePlatform,
		merge(RoleAuthenticated.Permissions, grant(
			PermListUsers,
			PermCreateUser,
			PermUpdateUser,
			PermManageUserRoles,
			PermManageDynamicRoles,
			PermGetUser,
			PermDeleteUser,
			PermResetPassword,
			PermListTermsOfUseAcceptances,
		)),
		ContextUser,
	)
	RolePlatformOrganizationManager = roleWithNoContext(
		"PlatformOrganizationManager",
		"Platform organization managers can manage organizations",
		RoleScopePlatform,
		merge(RoleAuthenticated.Permissions, grant(
			PermCreateOrganization,
			PermUpdateOrganization,
			PermDeleteOrganization,
		)),
		ContextUser,
	)
	RolePlatformAuditor = roleWithNoContext(
		"PlatformAuditor",
		"Platform auditors can audit the platform",
		RoleScopePlatform,
		merge(RoleAuthenticated.Permissions, grant(
			PermAuditPlatform,
		)),
		ContextUser,
	)
	RolePlatformWorkspaceManager = roleWithNoContext(
		"PlatformWorkspaceManager",
		"Platform workspace managers can create, update, and delete any workspace",
		RoleScopePlatform,
		merge(RoleAuthenticated.Permissions, grant(
			PermCreateWorkspace,
			PermGetWorkspace,
			PermUpdateWorkspace,
			PermDeleteWorkspace,
		)),
		ContextWorkspace,
	)
	RoleAppStoreAdmin = roleWithNoContext(
		"AppStoreAdmin",
		"App store admins can administer apps",
		RoleScopePlatform,
		merge(RoleAuthenticated.Permissions, grant(
			PermListApps,
			PermCreateApp,
			PermUpdateApp,
			PermGetApp,
			PermDeleteApp,
		)),
		ContextUser,
	)
	RolePlatformDataManager = roleWithNoContext(
		"PlatformDataManager",
		"Data managers can manage workspace data across workspaces",
		RoleScopePlatform,
		merge(RoleAuthenticated.Permissions, grant(
			PermListWorkspaces,
			PermGetWorkspace,
			PermListFilesInWorkspace,
			PermUploadFilesToWorkspace,
			PermModifyFilesInWorkspace,
			PermDownloadFilesFromWorkspace,
		)),
		ContextWorkspace,
	)
	RoleSuperAdmin = roleWithNoContext(
		"SuperAdmin",
		"Super admins can perform all platform, workspace, and workbench actions",
		RoleScopeSystem,
		merge(
			RoleAuthenticated.Permissions,
			RolePlatformSettingsManager.Permissions,
			RolePlatformUserManager.Permissions,
			RolePlatformOrganizationManager.Permissions,
			RolePlatformAuditor.Permissions,
			RolePlatformWorkspaceManager.Permissions,
			RoleAppStoreAdmin.Permissions,
			RolePlatformDataManager.Permissions,
			RoleWorkspaceAdmin.Permissions,
			RoleWorkspaceDataManager.Permissions,
			RoleWorkbenchAdmin.Permissions,
			RoleHealthchecker.Permissions,
			grant(PermInitializeTenant),
		),
		ContextUser,
		ContextWorkspace,
		ContextWorkbench,
	)
)
