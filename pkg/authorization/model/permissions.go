package model

import "fmt"

// permissionDefinitions is the single source of every permission's definition,
// populated by the perm0/perm1/perm2 declarations in permissions.go.
var permissionDefinitions []PermissionDefinition

func registerPermissionDefinition(name PermissionName, description string, dims ...ContextDimension) PermissionDefinition {
	def := PermissionDefinition{
		Name:                      name,
		Description:               description,
		RequiredContextDimensions: dims,
	}
	permissionDefinitions = append(permissionDefinitions, def)
	return def
}

// GetPermissionDefinitions returns every declared permission definition.
func GetPermissionDefinitions() []PermissionDefinition { return permissionDefinitions }

// -------------------------------------------------------------------
// Permission factories
// -------------------------------------------------------------------

// A permission factory embeds its definition and adds a compile-time-typed
// constructor: the id type parameter is the required dimension, so .For only
// accepts the matching id. The types are unexported — declare with
// perm0/perm1/perm2 (permissions.go) and reach them through .For.

// permissionFactory is any permission factory, regardless of arity — grant reads the
// name off it (promoted from the embedded PermissionDefinition).
type permissionFactory interface{ name() PermissionName }

func (d PermissionDefinition) name() PermissionName { return d.Name }

type permissionFactoryNoContext struct{ PermissionDefinition }

func (p permissionFactoryNoContext) For() Permission { return Permission{Name: p.Name} }

type permissionFactoryOneContext[A contextID] struct{ PermissionDefinition }

func (p permissionFactoryOneContext[A]) For(a A) Permission {
	return Permission{Name: p.Name, Context: Context{a.dimension(): fmt.Sprint(a)}}
}

type permissionFactoryTwoContexts[A, B contextID] struct{ PermissionDefinition }

func (p permissionFactoryTwoContexts[A, B]) For(a A, b B) Permission {
	return Permission{Name: p.Name, Context: Context{
		a.dimension(): fmt.Sprint(a),
		b.dimension(): fmt.Sprint(b),
	}}
}

func newPermissionFactoryNoContext(name PermissionName, description string) permissionFactoryNoContext {
	return permissionFactoryNoContext{registerPermissionDefinition(name, description)}
}

func newPermissionFactoryOneContext[A contextID](name PermissionName, description string) permissionFactoryOneContext[A] {
	var a A
	return permissionFactoryOneContext[A]{registerPermissionDefinition(name, description, a.dimension())}
}

func newPermissionFactoryTwoContexts[A, B contextID](name PermissionName, description string) permissionFactoryTwoContexts[A, B] {
	var a A
	var b B
	return permissionFactoryTwoContexts[A, B]{registerPermissionDefinition(name, description, a.dimension(), b.dimension())}
}

// One factory per permission — name, description, and required context (via the
// id type).
var (
	// Authentication
	PermAuthenticate                       = newPermissionFactoryNoContext("authenticate", "Allow the user to authenticate")
	PermLogout                             = newPermissionFactoryNoContext("logout", "Allow the user to logout")
	PermRefreshToken                       = newPermissionFactoryNoContext("refreshToken", "Allow the user to refresh the jwt token")
	PermGetListOfPossibleWayToAuthenticate = newPermissionFactoryNoContext("getListOfPossibleWayToAuthenticate", "Allow the user to get a list of possible ways to authenticate")
	PermAuthenticateUsingAuth2_0           = newPermissionFactoryNoContext("authenticateUsingAuth2.0", "Allow the user to authenticate using oauth2")
	PermAuthenticateRedirectUsingAuth2_0   = newPermissionFactoryNoContext("authenticateRedirectUsingAuth2.0", "Allow the user to be redirected after authenticating using oauth2")

	// Health
	PermGetHealthCheck = newPermissionFactoryNoContext("getHealthCheck", "Allow the user to get the healthcheck status")

	// Tenant
	PermInitializeTenant = newPermissionFactoryNoContext("initializeTenant", "Allow the user to initialize the tenant")

	// Notifications
	PermListNotifications        = newPermissionFactoryOneContext[UserID]("listNotifications", "Allow the user to list notifications")
	PermCountUnreadNotifications = newPermissionFactoryOneContext[UserID]("countUnreadNotifications", "Allow the user to count unread notifications")
	PermMarkNotificationAsRead   = newPermissionFactoryOneContext[UserID]("markNotificationAsRead", "Allow the user to mark a notification as read")

	// Users
	PermListUsers       = newPermissionFactoryNoContext("listUsers", "Allow the user to list users")
	PermSearchUsers     = newPermissionFactoryNoContext("searchUsers", "Allow the user to search users")
	PermCreateUser      = newPermissionFactoryNoContext("createUser", "Allow the user to create a user")
	PermUpdateUser      = newPermissionFactoryOneContext[UserID]("updateUser", "Allow the user to update a user")
	PermGetMyOwnUser    = newPermissionFactoryOneContext[UserID]("getMyOwnUser", "Allow the user to get his own user")
	PermUpdatePassword  = newPermissionFactoryOneContext[UserID]("updatePassword", "Allow the user to update his password")
	PermEnableTotp      = newPermissionFactoryOneContext[UserID]("enableTotp", "Allow the user to enable TOTP")
	PermResetTotp       = newPermissionFactoryOneContext[UserID]("resetTotp", "Allow the user to reset TOTP")
	PermGetUser         = newPermissionFactoryOneContext[UserID]("getUser", "Allow the user to get a user")
	PermDeleteUser      = newPermissionFactoryOneContext[UserID]("deleteUser", "Allow the user to delete a user")
	PermResetPassword   = newPermissionFactoryOneContext[UserID]("resetPassword", "Allow the user to reset a user's password")
	PermManageUserRoles = newPermissionFactoryOneContext[UserID]("manageUserRoles", "Allow the user to manage user roles")
	PermAuditUser       = newPermissionFactoryOneContext[UserID]("auditUser", "Allow the user to audit users")

	// Platform
	PermGetPlatformSettings = newPermissionFactoryNoContext("getPlatformSettings", "Allow the user to get platform settings")
	PermSetPlatformSettings = newPermissionFactoryNoContext("setPlatformSettings", "Allow the user to set platform settings")
	PermAuditPlatform       = newPermissionFactoryNoContext("auditPlatform", "Allow the user to audit the platform")
	PermManageDynamicRoles  = newPermissionFactoryNoContext("manageDynamicRoles", "Allow the user to create dynamic roles")

	// App instances
	PermListAppInstances  = newPermissionFactoryNoContext("listAppInstances", "Allow the user to list app instances")
	PermCreateAppInstance = newPermissionFactoryOneContext[WorkbenchID]("createAppInstance", "Allow the user to create an app instance")
	PermUpdateAppInstance = newPermissionFactoryOneContext[WorkbenchID]("updateAppInstance", "Allow the user to update an app instance")
	PermGetAppInstance    = newPermissionFactoryOneContext[WorkbenchID]("getAppInstance", "Allow the user to get an app instance")
	PermDeleteAppInstance = newPermissionFactoryOneContext[WorkbenchID]("deleteAppInstance", "Allow the user to delete an app instance")

	// Workbenches
	PermListWorkbenches        = newPermissionFactoryOneContext[WorkbenchID]("listWorkbenches", "Allow the user to list workbenchs")
	PermCreateWorkbench        = newPermissionFactoryOneContext[WorkspaceID]("createWorkbench", "Allow the user to create a workbench")
	PermUpdateWorkbench        = newPermissionFactoryOneContext[WorkbenchID]("updateWorkbench", "Allow the user to update a workbench")
	PermGetWorkbench           = newPermissionFactoryOneContext[WorkbenchID]("getWorkbench", "Allow the user to get a workbench")
	PermStreamWorkbench        = newPermissionFactoryOneContext[WorkbenchID]("streamWorkbench", "Allow the user to stream a workbench")
	PermDeleteWorkbench        = newPermissionFactoryOneContext[WorkbenchID]("deleteWorkbench", "Allow the user to delete a workbench")
	PermAuditWorkbench         = newPermissionFactoryOneContext[WorkbenchID]("auditWorkbench", "Allow the user to audit a workbench")
	PermManageUsersInWorkbench = newPermissionFactoryOneContext[WorkbenchID]("manageUsersInWorkbench", "Allow the user to manage users in a workbench")

	// Workspaces
	PermListWorkspaces                 = newPermissionFactoryOneContext[WorkspaceID]("listWorkspaces", "Allow the user to list workspaces")
	PermListPublicWorkspaces           = newPermissionFactoryNoContext("listPublicWorkspaces", "Allow the user to list public workspaces")
	PermCreateWorkspace                = newPermissionFactoryNoContext("createWorkspace", "Allow the user to create a workspace")
	PermUpdateWorkspace                = newPermissionFactoryOneContext[WorkspaceID]("updateWorkspace", "Allow the user to update a workspace")
	PermGetWorkspace                   = newPermissionFactoryOneContext[WorkspaceID]("getWorkspace", "Allow the user to get a workspace")
	PermDeleteWorkspace                = newPermissionFactoryOneContext[WorkspaceID]("deleteWorkspace", "Allow the user to delete a workspace")
	PermManageUsersInWorkspace         = newPermissionFactoryOneContext[WorkspaceID]("manageUsersInWorkspace", "Allow the user to manage users in a workspace")
	PermManageUsersDataRoleInWorkspace = newPermissionFactoryOneContext[WorkspaceID]("manageUsersDataRoleInWorkspace", "Allow the user to manage users' data role in a workspace")
	PermListFilesInWorkspace           = newPermissionFactoryOneContext[WorkspaceID]("listFilesInWorkspace", "Allow the user to list files in a workspace")
	PermUploadFilesToWorkspace         = newPermissionFactoryOneContext[WorkspaceID]("uploadFilesToWorkspace", "Allow the user to upload files to a workspace")
	PermDownloadFilesFromWorkspace     = newPermissionFactoryOneContext[WorkspaceID]("downloadFilesFromWorkspace", "Allow the user to download files from a workspace")
	PermModifyFilesInWorkspace         = newPermissionFactoryOneContext[WorkspaceID]("modifyFilesInWorkspace", "Allow the user to modify files in a workspace")
	PermAuditWorkspace                 = newPermissionFactoryOneContext[WorkspaceID]("auditWorkspace", "Allow the user to audit a workspace")

	// Workspace service instances
	PermListWorkspaceServiceInstances     = newPermissionFactoryOneContext[WorkspaceID]("listWorkspaceServiceInstances", "Allow the user to list workspace service instances")
	PermCreateWorkspaceServiceInstance    = newPermissionFactoryOneContext[WorkspaceID]("createWorkspaceServiceInstance", "Allow the user to create a workspace service instance")
	PermUpdateWorkspaceServiceInstance    = newPermissionFactoryOneContext[WorkspaceID]("updateWorkspaceServiceInstance", "Allow the user to update a workspace service instance")
	PermGetWorkspaceServiceInstance       = newPermissionFactoryOneContext[WorkspaceID]("getWorkspaceServiceInstance", "Allow the user to get a workspace service instance")
	PermGetWorkspaceServiceInstanceSecret = newPermissionFactoryOneContext[WorkspaceID]("getWorkspaceServiceInstanceSecret", "Allow the user to get the secrets of a workspace service instance")
	PermDeleteWorkspaceServiceInstance    = newPermissionFactoryOneContext[WorkspaceID]("deleteWorkspaceServiceInstance", "Allow the user to delete a workspace service instance")

	// Apps
	PermListApps  = newPermissionFactoryNoContext("listApps", "Allow the user to list apps")
	PermCreateApp = newPermissionFactoryNoContext("createApp", "Allow the user to create an app")
	PermUpdateApp = newPermissionFactoryNoContext("updateApp", "Allow the user to update an app")
	PermGetApp    = newPermissionFactoryNoContext("getApp", "Allow the user to get an app")
	PermDeleteApp = newPermissionFactoryNoContext("deleteApp", "Allow the user to delete an app")

	// Requests
	PermListRequests   = newPermissionFactoryOneContext[WorkspaceID]("listRequests", "Allow the user to list requests")
	PermListMyRequests = newPermissionFactoryNoContext("listMyRequests", "Allow the user to list his requests")
	PermGetRequest     = newPermissionFactoryOneContext[WorkspaceID]("getRequest", "Allow the user to get a request")
	PermCreateRequest  = newPermissionFactoryOneContext[WorkspaceID]("createRequest", "Allow the user to create a request")
	PermApproveRequest = newPermissionFactoryOneContext[WorkspaceID]("approveRequest", "Allow the user to approve a request")
	PermDeleteRequest  = newPermissionFactoryOneContext[WorkspaceID]("deleteRequest", "Allow the user to delete a request")

	// Terms of use
	PermCreateTermsOfUseVersion     = newPermissionFactoryNoContext("createTermsOfUseVersion", "Allow the user to create a terms of use version")
	PermUpdateTermsOfUseVersion     = newPermissionFactoryNoContext("updateTermsOfUseVersion", "Allow the user to update a terms of use version")
	PermPublishTermsOfUseVersion    = newPermissionFactoryNoContext("publishTermsOfUseVersion", "Allow the user to publish a terms of use version")
	PermGetTermsOfUseVersion        = newPermissionFactoryNoContext("getTermsOfUseVersion", "Allow the user to get a terms of use version")
	PermListTermsOfUseVersions      = newPermissionFactoryNoContext("listTermsOfUseVersions", "Allow the user to list terms of use versions")
	PermGetCurrentTermsOfUseVersion = newPermissionFactoryNoContext("getCurrentTermsOfUseVersion", "Allow the user to get the current terms of use version")
	PermListTermsOfUseAcceptances   = newPermissionFactoryNoContext("listTermsOfUseAcceptances", "Allow the user to list terms of use acceptances")
	PermGetMyTermsOfUseStatus       = newPermissionFactoryOneContext[UserID]("getMyTermsOfUseStatus", "Allow the user to get his terms of use acceptance status")
	PermAcceptTermsOfUse            = newPermissionFactoryOneContext[UserID]("acceptTermsOfUse", "Allow the user to accept the terms of use")

	// Organizations
	PermListOrganizations  = newPermissionFactoryNoContext("listOrganizations", "Allow the user to list organizations")
	PermGetOrganization    = newPermissionFactoryNoContext("getOrganization", "Allow the user to get an organization")
	PermCreateOrganization = newPermissionFactoryNoContext("createOrganization", "Allow the user to create an organization")
	PermUpdateOrganization = newPermissionFactoryNoContext("updateOrganization", "Allow the user to update an organization")
	PermDeleteOrganization = newPermissionFactoryNoContext("deleteOrganization", "Allow the user to delete an organization")
)
