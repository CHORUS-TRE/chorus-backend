package model

import "fmt"

// roleDefinitions is the single source of every role's definition, populated by
// the role0/role1 declarations in roles.go.
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

// A role factory bakes in wildcard dimensions and exposes a typed .For for the
// exact-bound value. Declare with role0/role1 (roles.go).

// roleFactoryNoContext is the factory for a role that binds no exact context value. It
// may still bake in wildcard dimensions (granted for any value).
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

// Permission grants, composed by inheritance. Each list is the single source
// for the role(s) that use it; merge de-duplicates across groups.
var (
	publicPermissions = grant(
		Authenticate,
		GetListOfPossibleWayToAuthenticate,
		AuthenticateUsingAuth2_0,
		AuthenticateRedirectUsingAuth2_0,
		GetPlatformSettings,
	)

	authenticatedPermissions = merge(publicPermissions, grant(
		ListNotifications,
		CountUnreadNotifications,
		MarkNotificationAsRead,
		Logout,
		RefreshToken,
		UpdateUser,
		GetMyOwnUser,
		UpdatePassword,
		EnableTotp,
		ResetTotp,
		ResetPassword,
		ListWorkspaces,
		ListPublicWorkspaces,
		ListWorkbenchs,
		ListApps,
		ListAppInstances,
		ListMyRequests,
		AuditUser,
		GetCurrentTermsOfUseVersion,
		GetMyTermsOfUseStatus,
		AcceptTermsOfUse,
		ListOrganizations,
		GetOrganization,
	))

	workspaceGuestPermissions = merge(authenticatedPermissions, grant(
		ListWorkspaces,
		GetWorkspace,
		ListUsers,
		CreateRequest,
		ListWorkspaceServiceInstances,
	))

	workspaceMemberPermissions = merge(workspaceGuestPermissions, grant(
		CreateWorkbench,
		ListFilesInWorkspace,
		CreateRequest,
		GetWorkspaceServiceInstance,
		GetWorkspaceServiceInstanceSecret,
	))

	workspaceMaintainerPermissions = merge(workspaceMemberPermissions, grant(
		UpdateWorkspace,
		UploadFilesToWorkspace,
		ModifyFilesInWorkspace,
		SearchUsers,
		CreateRequest,
	))

	workspaceDataManagerPermissions = merge(workspaceMemberPermissions, grant(
		UploadFilesToWorkspace,
		ModifyFilesInWorkspace,
		DownloadFilesFromWorkspace,
		ManageUsersDataRoleInWorkspace,
		CreateRequest,
		ListRequests,
	))

	workspaceAdminPermissions = merge(workspaceMaintainerPermissions, grant(
		ListAppInstances,
		ListWorkbenchs,
		UpdateWorkbench,
		GetWorkbench,
		StreamWorkbench,
		DeleteWorkbench,
		AuditWorkbench,
		ManageUsersInWorkbench,
		DeleteWorkspace,
		AuditWorkspace,
		ManageUsersInWorkspace,
		ListRequests,
		GetRequest,
		ApproveRequest,
		DeleteRequest,
		CreateWorkspaceServiceInstance,
		UpdateWorkspaceServiceInstance,
		DeleteWorkspaceServiceInstance,
	))

	workbenchViewerPermissions = merge(authenticatedPermissions, grant(
		ListAppInstances,
		ListWorkbenchs,
		GetWorkbench,
		StreamWorkbench,
		ListUsers,
	))

	workbenchMemberPermissions = merge(workbenchViewerPermissions, grant(
		CreateAppInstance,
		UpdateAppInstance,
		GetAppInstance,
		DeleteAppInstance,
		UpdateWorkbench,
	))

	workbenchAdminPermissions = merge(workbenchMemberPermissions, grant(
		DeleteWorkbench,
		ManageUsersInWorkbench,
		SearchUsers,
		AuditWorkbench,
	))

	platformSettingsManagerPermissions = merge(authenticatedPermissions, grant(
		SetPlatformSettings,
		ListTermsOfUseVersions,
		GetTermsOfUseVersion,
		CreateTermsOfUseVersion,
		UpdateTermsOfUseVersion,
		PublishTermsOfUseVersion,
	))

	platformUserManagerPermissions = merge(authenticatedPermissions, grant(
		ListUsers,
		CreateUser,
		UpdateUser,
		ManageUserRoles,
		ManageDynamicRoles,
		GetUser,
		DeleteUser,
		ResetPassword,
		ListTermsOfUseAcceptances,
	))

	platformOrganizationManagerPermissions = merge(authenticatedPermissions, grant(
		CreateOrganization,
		UpdateOrganization,
		DeleteOrganization,
	))

	platformAuditorPermissions = merge(authenticatedPermissions, grant(
		AuditPlatform,
	))

	platformWorkspaceManagerPermissions = merge(authenticatedPermissions, grant(
		CreateWorkspace,
		GetWorkspace,
		UpdateWorkspace,
		DeleteWorkspace,
	))

	appStoreAdminPermissions = merge(authenticatedPermissions, grant(
		ListApps,
		CreateApp,
		UpdateApp,
		GetApp,
		DeleteApp,
	))

	dataManagerPermissions = merge(authenticatedPermissions, grant(
		ListWorkspaces,
		GetWorkspace,
		ListFilesInWorkspace,
		UploadFilesToWorkspace,
		ModifyFilesInWorkspace,
		DownloadFilesFromWorkspace,
	))

	healthcheckerPermissions = grant(
		GetHealthCheck,
	)

	superAdminPermissions = merge(
		authenticatedPermissions,
		platformSettingsManagerPermissions,
		platformUserManagerPermissions,
		platformOrganizationManagerPermissions,
		platformAuditorPermissions,
		platformWorkspaceManagerPermissions,
		appStoreAdminPermissions,
		dataManagerPermissions,
		workspaceAdminPermissions,
		workspaceDataManagerPermissions,
		workbenchAdminPermissions,
		healthcheckerPermissions,
		grant(InitializeTenant),
	)
)

// One factory per role — name, description, scope, granted permissions, and the
// context binding: exact via the id type parameter, wildcard via trailing args.
var (
	Public = roleWithNoContext(
		RolePublic,
		"Public users can authenticate and read public platform settings",
		RoleScopePlatform,
		publicPermissions,
	)
	Authenticated = roleWithOneContext[UserID](
		RoleAuthenticated,
		"Authenticated users can manage their own session, profile, notifications, and base resources",
		RoleScopePlatform,
		authenticatedPermissions,
	)
	WorkspaceGuest = roleWithOneContext[WorkspaceID](
		RoleWorkspaceGuest,
		"Workspace guests can view workspace metadata and create requests",
		RoleScopeWorkspace,
		workspaceGuestPermissions,
	)
	WorkspaceMember = roleWithOneContext[WorkspaceID](
		RoleWorkspaceMember,
		"Workspace members can create workbenches and list workspace files",
		RoleScopeWorkspace,
		workspaceMemberPermissions,
	)
	WorkspaceMaintainer = roleWithOneContext[WorkspaceID](
		RoleWorkspaceMaintainer,
		"Workspace maintainers can update workspace metadata and manage workspace files",
		RoleScopeWorkspace,
		workspaceMaintainerPermissions,
	)
	WorkspaceDataManager = roleWithOneContext[WorkspaceID](
		RoleWorkspaceDataManager,
		"Workspace data managers can manage workspace files and data-manager assignments",
		RoleScopeWorkspace,
		workspaceDataManagerPermissions,
	)
	WorkspaceAdmin = roleWithOneContext[WorkspaceID](
		RoleWorkspaceAdmin,
		"Workspace admins can administer workspace users, requests, workbenches, files, and services",
		RoleScopeWorkspace,
		workspaceAdminPermissions,
	)
	WorkbenchViewer = roleWithOneContext[WorkbenchID](
		RoleWorkbenchViewer,
		"Workbench viewers can view and stream workbenches",
		RoleScopeWorkbench,
		workbenchViewerPermissions,
	)
	WorkbenchMember = roleWithOneContext[WorkbenchID](
		RoleWorkbenchMember,
		"Workbench members can update workbenches and manage app instances",
		RoleScopeWorkbench,
		workbenchMemberPermissions,
	)
	WorkbenchAdmin = roleWithOneContext[WorkbenchID](
		RoleWorkbenchAdmin,
		"Workbench admins can administer workbenches and their users",
		RoleScopeWorkbench,
		workbenchAdminPermissions,
	)
	Healthchecker = roleWithNoContext(
		RoleHealthchecker,
		"Healthcheckers can read healthcheck status",
		RoleScopePlatform,
		healthcheckerPermissions,
		ContextUser,
	)
	PlatformSettingsManager = roleWithNoContext(
		RolePlatformSettingsManager,
		"Platform settings managers can manage platform settings",
		RoleScopePlatform,
		platformSettingsManagerPermissions,
		ContextUser,
	)
	PlatformUserManager = roleWithNoContext(
		RolePlatformUserManager,
		"Platform user managers can administer platform users and their roles",
		RoleScopePlatform,
		platformUserManagerPermissions,
		ContextUser,
	)
	PlatformOrganizationManager = roleWithNoContext(
		RolePlatformOrganizationManager,
		"Platform organization managers can manage organizations",
		RoleScopePlatform,
		platformOrganizationManagerPermissions,
		ContextUser,
	)
	PlatformAuditor = roleWithNoContext(
		RolePlatformAuditor,
		"Platform auditors can audit the platform",
		RoleScopePlatform,
		platformAuditorPermissions,
		ContextUser,
	)
	PlatformWorkspaceManager = roleWithNoContext(
		RolePlatformWorkspaceManager,
		"Platform workspace managers can create, update, and delete any workspace",
		RoleScopePlatform,
		platformWorkspaceManagerPermissions,
		ContextWorkspace,
	)
	AppStoreAdmin = roleWithNoContext(
		RoleAppStoreAdmin,
		"App store admins can administer apps",
		RoleScopePlatform,
		appStoreAdminPermissions,
		ContextUser,
	)
	PlatformDataManager = roleWithNoContext(
		RolePlatformDataManager,
		"Data managers can manage workspace data across workspaces",
		RoleScopePlatform,
		dataManagerPermissions,
		ContextWorkspace,
	)
	SuperAdmin = roleWithNoContext(
		RoleSuperAdmin,
		"Super admins can perform all platform, workspace, and workbench actions",
		RoleScopeSystem,
		superAdminPermissions,
		ContextUser,
		ContextWorkspace,
		ContextWorkbench,
	)
)
