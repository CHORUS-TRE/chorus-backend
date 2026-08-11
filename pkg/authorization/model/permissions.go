package model

import "fmt"

// permissionDefinitions is the single source of every permission's definition,
// populated by the perm0/perm1/perm2 declarations in permissions.go.
var permissionDefinitions []PermissionDefinition

func registerPermission(name PermissionName, description string, dims ...ContextDimension) PermissionDefinition {
	def := PermissionDefinition{
		Name:                      name,
		Description:               description,
		RequiredContextDimensions: dims,
	}
	permissionDefinitions = append(permissionDefinitions, def)
	return def
}

// PermissionDefinitions returns every declared permission definition.
func PermissionDefinitions() []PermissionDefinition { return permissionDefinitions }

// PermissionDefinitionsMap returns every declared permission definition as a map keyed by permission name.
func PermissionDefinitionsMap() map[PermissionName]PermissionDefinition {
	m := make(map[PermissionName]PermissionDefinition, len(permissionDefinitions))
	for _, def := range permissionDefinitions {
		m[def.Name] = def
	}
	return m
}

// -------------------------------------------------------------------
// Permission factories
// -------------------------------------------------------------------

// A permission factory embeds its definition and adds a compile-time-typed
// constructor: the id type parameter is the required dimension, so .For only
// accepts the matching id. The types are unexported — declare with
// perm0/perm1/perm2 (permissions.go) and reach them through .For.

// permFactory is any permission factory, regardless of arity — grant reads the
// name off it (promoted from the embedded PermissionDefinition).
type permFactory interface{ name() PermissionName }

func (d PermissionDefinition) name() PermissionName { return d.Name }

type permFactoryNoContext struct{ PermissionDefinition }

func (p permFactoryNoContext) For() Permission { return Permission{Name: p.Name} }

type permFactoryOneContext[A contextID] struct{ PermissionDefinition }

func (p permFactoryOneContext[A]) For(a A) Permission {
	return Permission{Name: p.Name, Context: Context{a.dimension(): fmt.Sprint(a)}}
}

type permFactoryTwoContexts[A, B contextID] struct{ PermissionDefinition }

func (p permFactoryTwoContexts[A, B]) For(a A, b B) Permission {
	return Permission{Name: p.Name, Context: Context{
		a.dimension(): fmt.Sprint(a),
		b.dimension(): fmt.Sprint(b),
	}}
}

func permWithNoContext(name PermissionName, description string) permFactoryNoContext {
	return permFactoryNoContext{registerPermission(name, description)}
}

func permWithOneContext[A contextID](name PermissionName, description string) permFactoryOneContext[A] {
	var a A
	return permFactoryOneContext[A]{registerPermission(name, description, a.dimension())}
}

func permWithTwoContexts[A, B contextID](name PermissionName, description string) permFactoryTwoContexts[A, B] {
	var a A
	var b B
	return permFactoryTwoContexts[A, B]{registerPermission(name, description, a.dimension(), b.dimension())}
}

// One factory per permission — name, description, and required context (via the
// id type).
var (
	// Authentication
	PermAuthenticate                       = permWithNoContext("authenticate", "Allow the user to authenticate")
	PermLogout                             = permWithNoContext("logout", "Allow the user to logout")
	PermRefreshToken                       = permWithNoContext("refreshToken", "Allow the user to refresh the jwt token")
	PermGetListOfPossibleWayToAuthenticate = permWithNoContext("getListOfPossibleWayToAuthenticate", "Allow the user to get a list of possible ways to authenticate")
	PermAuthenticateUsingAuth2_0           = permWithNoContext("authenticateUsingAuth2.0", "Allow the user to authenticate using oauth2")
	PermAuthenticateRedirectUsingAuth2_0   = permWithNoContext("authenticateRedirectUsingAuth2.0", "Allow the user to be redirected after authenticating using oauth2")

	// Health
	PermGetHealthCheck = permWithNoContext("getHealthCheck", "Allow the user to get the healthcheck status")

	// Tenant
	PermInitializeTenant = permWithNoContext("initializeTenant", "Allow the user to initialize the tenant")

	// Notifications
	PermListNotifications        = permWithOneContext[UserID]("listNotifications", "Allow the user to list notifications")
	PermCountUnreadNotifications = permWithOneContext[UserID]("countUnreadNotifications", "Allow the user to count unread notifications")
	PermMarkNotificationAsRead   = permWithOneContext[UserID]("markNotificationAsRead", "Allow the user to mark a notification as read")

	// Users
	PermListUsers       = permWithNoContext("listUsers", "Allow the user to list users")
	PermSearchUsers     = permWithNoContext("searchUsers", "Allow the user to search users")
	PermCreateUser      = permWithNoContext("createUser", "Allow the user to create a user")
	PermUpdateUser      = permWithOneContext[UserID]("updateUser", "Allow the user to update a user")
	PermGetMyOwnUser    = permWithOneContext[UserID]("getMyOwnUser", "Allow the user to get his own user")
	PermUpdatePassword  = permWithOneContext[UserID]("updatePassword", "Allow the user to update his password")
	PermEnableTotp      = permWithOneContext[UserID]("enableTotp", "Allow the user to enable TOTP")
	PermResetTotp       = permWithOneContext[UserID]("resetTotp", "Allow the user to reset TOTP")
	PermGetUser         = permWithOneContext[UserID]("getUser", "Allow the user to get a user")
	PermDeleteUser      = permWithOneContext[UserID]("deleteUser", "Allow the user to delete a user")
	PermResetPassword   = permWithOneContext[UserID]("resetPassword", "Allow the user to reset a user's password")
	PermManageUserRoles = permWithOneContext[UserID]("manageUserRoles", "Allow the user to manage user roles")
	PermAuditUser       = permWithOneContext[UserID]("auditUser", "Allow the user to audit users")

	// Platform
	PermGetPlatformSettings = permWithNoContext("getPlatformSettings", "Allow the user to get platform settings")
	PermSetPlatformSettings = permWithNoContext("setPlatformSettings", "Allow the user to set platform settings")
	PermAuditPlatform       = permWithNoContext("auditPlatform", "Allow the user to audit the platform")
	PermManageDynamicRoles  = permWithNoContext("manageDynamicRoles", "Allow the user to create dynamic roles")

	// App instances
	PermListAppInstances  = permWithNoContext("listAppInstances", "Allow the user to list app instances")
	PermCreateAppInstance = permWithOneContext[WorkbenchID]("createAppInstance", "Allow the user to create an app instance")
	PermUpdateAppInstance = permWithOneContext[WorkbenchID]("updateAppInstance", "Allow the user to update an app instance")
	PermGetAppInstance    = permWithOneContext[WorkbenchID]("getAppInstance", "Allow the user to get an app instance")
	PermDeleteAppInstance = permWithOneContext[WorkbenchID]("deleteAppInstance", "Allow the user to delete an app instance")

	// Workbenches
	PermListWorkbenchs         = permWithOneContext[WorkbenchID]("listWorkbenchs", "Allow the user to list workbenchs")
	PermCreateWorkbench        = permWithOneContext[WorkspaceID]("createWorkbench", "Allow the user to create a workbench")
	PermUpdateWorkbench        = permWithOneContext[WorkbenchID]("updateWorkbench", "Allow the user to update a workbench")
	PermGetWorkbench           = permWithOneContext[WorkbenchID]("getWorkbench", "Allow the user to get a workbench")
	PermStreamWorkbench        = permWithOneContext[WorkbenchID]("streamWorkbench", "Allow the user to stream a workbench")
	PermDeleteWorkbench        = permWithOneContext[WorkbenchID]("deleteWorkbench", "Allow the user to delete a workbench")
	PermAuditWorkbench         = permWithOneContext[WorkbenchID]("auditWorkbench", "Allow the user to audit a workbench")
	PermManageUsersInWorkbench = permWithOneContext[WorkbenchID]("manageUsersInWorkbench", "Allow the user to manage users in a workbench")

	// Workspaces
	PermListWorkspaces                 = permWithOneContext[WorkspaceID]("listWorkspaces", "Allow the user to list workspaces")
	PermListPublicWorkspaces           = permWithNoContext("listPublicWorkspaces", "Allow the user to list public workspaces")
	PermCreateWorkspace                = permWithNoContext("createWorkspace", "Allow the user to create a workspace")
	PermUpdateWorkspace                = permWithOneContext[WorkspaceID]("updateWorkspace", "Allow the user to update a workspace")
	PermGetWorkspace                   = permWithOneContext[WorkspaceID]("getWorkspace", "Allow the user to get a workspace")
	PermDeleteWorkspace                = permWithOneContext[WorkspaceID]("deleteWorkspace", "Allow the user to delete a workspace")
	PermManageUsersInWorkspace         = permWithOneContext[WorkspaceID]("manageUsersInWorkspace", "Allow the user to manage users in a workspace")
	PermManageUsersDataRoleInWorkspace = permWithOneContext[WorkspaceID]("manageUsersDataRoleInWorkspace", "Allow the user to manage users' data role in a workspace")
	PermListFilesInWorkspace           = permWithOneContext[WorkspaceID]("listFilesInWorkspace", "Allow the user to list files in a workspace")
	PermUploadFilesToWorkspace         = permWithOneContext[WorkspaceID]("uploadFilesToWorkspace", "Allow the user to upload files to a workspace")
	PermDownloadFilesFromWorkspace     = permWithOneContext[WorkspaceID]("downloadFilesFromWorkspace", "Allow the user to download files from a workspace")
	PermModifyFilesInWorkspace         = permWithOneContext[WorkspaceID]("modifyFilesInWorkspace", "Allow the user to modify files in a workspace")
	PermAuditWorkspace                 = permWithOneContext[WorkspaceID]("auditWorkspace", "Allow the user to audit a workspace")

	// Workspace service instances
	PermListWorkspaceServiceInstances     = permWithOneContext[WorkspaceID]("listWorkspaceServiceInstances", "Allow the user to list workspace service instances")
	PermCreateWorkspaceServiceInstance    = permWithOneContext[WorkspaceID]("createWorkspaceServiceInstance", "Allow the user to create a workspace service instance")
	PermUpdateWorkspaceServiceInstance    = permWithOneContext[WorkspaceID]("updateWorkspaceServiceInstance", "Allow the user to update a workspace service instance")
	PermGetWorkspaceServiceInstance       = permWithOneContext[WorkspaceID]("getWorkspaceServiceInstance", "Allow the user to get a workspace service instance")
	PermGetWorkspaceServiceInstanceSecret = permWithOneContext[WorkspaceID]("getWorkspaceServiceInstanceSecret", "Allow the user to get the secrets of a workspace service instance")
	PermDeleteWorkspaceServiceInstance    = permWithOneContext[WorkspaceID]("deleteWorkspaceServiceInstance", "Allow the user to delete a workspace service instance")

	// Apps
	PermListApps  = permWithNoContext("listApps", "Allow the user to list apps")
	PermCreateApp = permWithNoContext("createApp", "Allow the user to create an app")
	PermUpdateApp = permWithNoContext("updateApp", "Allow the user to update an app")
	PermGetApp    = permWithNoContext("getApp", "Allow the user to get an app")
	PermDeleteApp = permWithNoContext("deleteApp", "Allow the user to delete an app")

	// Requests
	PermListRequests   = permWithOneContext[WorkspaceID]("listRequests", "Allow the user to list requests")
	PermListMyRequests = permWithNoContext("listMyRequests", "Allow the user to list his requests")
	PermGetRequest     = permWithOneContext[WorkspaceID]("getRequest", "Allow the user to get a request")
	PermCreateRequest  = permWithOneContext[WorkspaceID]("createRequest", "Allow the user to create a request")
	PermApproveRequest = permWithOneContext[WorkspaceID]("approveRequest", "Allow the user to approve a request")
	PermDeleteRequest  = permWithOneContext[WorkspaceID]("deleteRequest", "Allow the user to delete a request")

	// Terms of use
	PermCreateTermsOfUseVersion     = permWithNoContext("createTermsOfUseVersion", "Allow the user to create a terms of use version")
	PermUpdateTermsOfUseVersion     = permWithNoContext("updateTermsOfUseVersion", "Allow the user to update a terms of use version")
	PermPublishTermsOfUseVersion    = permWithNoContext("publishTermsOfUseVersion", "Allow the user to publish a terms of use version")
	PermGetTermsOfUseVersion        = permWithNoContext("getTermsOfUseVersion", "Allow the user to get a terms of use version")
	PermListTermsOfUseVersions      = permWithNoContext("listTermsOfUseVersions", "Allow the user to list terms of use versions")
	PermGetCurrentTermsOfUseVersion = permWithNoContext("getCurrentTermsOfUseVersion", "Allow the user to get the current terms of use version")
	PermListTermsOfUseAcceptances   = permWithNoContext("listTermsOfUseAcceptances", "Allow the user to list terms of use acceptances")
	PermGetMyTermsOfUseStatus       = permWithOneContext[UserID]("getMyTermsOfUseStatus", "Allow the user to get his terms of use acceptance status")
	PermAcceptTermsOfUse            = permWithOneContext[UserID]("acceptTermsOfUse", "Allow the user to accept the terms of use")

	// Organizations
	PermListOrganizations  = permWithNoContext("listOrganizations", "Allow the user to list organizations")
	PermGetOrganization    = permWithNoContext("getOrganization", "Allow the user to get an organization")
	PermCreateOrganization = permWithNoContext("createOrganization", "Allow the user to create an organization")
	PermUpdateOrganization = permWithNoContext("updateOrganization", "Allow the user to update an organization")
	PermDeleteOrganization = permWithNoContext("deleteOrganization", "Allow the user to delete an organization")
)
