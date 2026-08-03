package model

import "fmt"

// roleDefinitions is the single source of every role's definition, populated by
// the role0/role1 declarations in roles.go.
var roleDefinitions []*RoleDefinition

func registerRole(name RoleName, description string, scope RoleScope, required map[ContextDimension]ContextQuantifier, permissions []PermissionName) {
	roleDefinitions = append(roleDefinitions, &RoleDefinition{
		Name:                      name,
		Description:               description,
		Scope:                     scope,
		RequiredContextDimensions: required,
		Permissions:               append([]PermissionName(nil), permissions...),
	})
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

// roleFactory0 is the factory for a role that binds no exact context value. It
// may still bake in wildcard dimensions (granted for any value).
type roleFactory0 struct {
	Name      RoleName
	wildcards []ContextDimension
}

func (r roleFactory0) For() Role {
	ctx := make(Context, len(r.wildcards))
	for _, dim := range r.wildcards {
		ctx[dim] = Wildcard
	}
	return Role{Name: r.Name, Context: ctx}
}

// roleFactory1 is the factory for a role that binds one exact context value.
type roleFactory1[A contextID] struct {
	Name      RoleName
	wildcards []ContextDimension
}

func (r roleFactory1[A]) For(a A) Role {
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

func role0(name RoleName, description string, scope RoleScope, permissions []PermissionName, wildcards ...ContextDimension) roleFactory0 {
	registerRole(name, description, scope, requiredContext(nil, wildcards), permissions)
	return roleFactory0{Name: name, wildcards: wildcards}
}

func role1[A contextID](name RoleName, description string, scope RoleScope, permissions []PermissionName, wildcards ...ContextDimension) roleFactory1[A] {
	var a A
	registerRole(name, description, scope, requiredContext([]ContextDimension{a.dimension()}, wildcards), permissions)
	return roleFactory1[A]{Name: name, wildcards: wildcards}
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
	Public = role0(
		RolePublic,
		"Public users can authenticate and read public platform settings",
		RoleScopePlatform,
		publicPermissions,
	)
	Authenticated = role1[UserID](
		RoleAuthenticated,
		"Authenticated users can manage their own session, profile, notifications, and base resources",
		RoleScopePlatform,
		authenticatedPermissions,
	)
	WorkspaceGuest = role1[WorkspaceID](
		RoleWorkspaceGuest,
		"Workspace guests can view workspace metadata and create requests",
		RoleScopeWorkspace,
		workspaceGuestPermissions,
	)
	WorkspaceMember = role1[WorkspaceID](
		RoleWorkspaceMember,
		"Workspace members can create workbenches and list workspace files",
		RoleScopeWorkspace,
		workspaceMemberPermissions,
	)
	WorkspaceMaintainer = role1[WorkspaceID](
		RoleWorkspaceMaintainer,
		"Workspace maintainers can update workspace metadata and manage workspace files",
		RoleScopeWorkspace,
		workspaceMaintainerPermissions,
	)
	WorkspaceDataManager = role1[WorkspaceID](
		RoleWorkspaceDataManager,
		"Workspace data managers can manage workspace files and data-manager assignments",
		RoleScopeWorkspace,
		workspaceDataManagerPermissions,
	)
	WorkspaceAdmin = role1[WorkspaceID](
		RoleWorkspaceAdmin,
		"Workspace admins can administer workspace users, requests, workbenches, files, and services",
		RoleScopeWorkspace,
		workspaceAdminPermissions,
	)
	WorkbenchViewer = role1[WorkbenchID](
		RoleWorkbenchViewer,
		"Workbench viewers can view and stream workbenches",
		RoleScopeWorkbench,
		workbenchViewerPermissions,
	)
	WorkbenchMember = role1[WorkbenchID](
		RoleWorkbenchMember,
		"Workbench members can update workbenches and manage app instances",
		RoleScopeWorkbench,
		workbenchMemberPermissions,
	)
	WorkbenchAdmin = role1[WorkbenchID](
		RoleWorkbenchAdmin,
		"Workbench admins can administer workbenches and their users",
		RoleScopeWorkbench,
		workbenchAdminPermissions,
	)
	Healthchecker = role0(
		RoleHealthchecker,
		"Healthcheckers can read healthcheck status",
		RoleScopePlatform,
		healthcheckerPermissions,
		ContextUser,
	)
	PlatformSettingsManager = role0(
		RolePlatformSettingsManager,
		"Platform settings managers can manage platform settings",
		RoleScopePlatform,
		platformSettingsManagerPermissions,
		ContextUser,
	)
	PlatformUserManager = role0(
		RolePlatformUserManager,
		"Platform user managers can administer platform users and their roles",
		RoleScopePlatform,
		platformUserManagerPermissions,
		ContextUser,
	)
	PlatformOrganizationManager = role0(
		RolePlatformOrganizationManager,
		"Platform organization managers can manage organizations",
		RoleScopePlatform,
		platformOrganizationManagerPermissions,
		ContextUser,
	)
	PlatformAuditor = role0(
		RolePlatformAuditor,
		"Platform auditors can audit the platform",
		RoleScopePlatform,
		platformAuditorPermissions,
		ContextUser,
	)
	PlatformWorkspaceManager = role0(
		RolePlatformWorkspaceManager,
		"Platform workspace managers can create, update, and delete any workspace",
		RoleScopePlatform,
		platformWorkspaceManagerPermissions,
		ContextWorkspace,
	)
	AppStoreAdmin = role0(
		RoleAppStoreAdmin,
		"App store admins can administer apps",
		RoleScopePlatform,
		appStoreAdminPermissions,
		ContextUser,
	)
	PlatformDataManager = role0(
		RolePlatformDataManager,
		"Data managers can manage workspace data across workspaces",
		RoleScopePlatform,
		dataManagerPermissions,
		ContextWorkspace,
	)
	SuperAdmin = role0(
		RoleSuperAdmin,
		"Super admins can perform all platform, workspace, and workbench actions",
		RoleScopeSystem,
		superAdminPermissions,
		ContextUser,
		ContextWorkspace,
		ContextWorkbench,
	)
)
