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
	Authenticate                       = permWithNoContext("authenticate", "Allow the user to authenticate")
	Logout                             = permWithNoContext("logout", "Allow the user to logout")
	RefreshToken                       = permWithNoContext("refreshToken", "Allow the user to refresh the jwt token")
	GetListOfPossibleWayToAuthenticate = permWithNoContext("getListOfPossibleWayToAuthenticate", "Allow the user to get a list of possible ways to authenticate")
	AuthenticateUsingAuth2_0           = permWithNoContext("authenticateUsingAuth2.0", "Allow the user to authenticate using oauth2")
	AuthenticateRedirectUsingAuth2_0   = permWithNoContext("authenticateRedirectUsingAuth2.0", "Allow the user to be redirected after authenticating using oauth2")

	// Health
	GetHealthCheck = permWithNoContext("getHealthCheck", "Allow the user to get the healthcheck status")

	// Tenant
	InitializeTenant = permWithNoContext("initializeTenant", "Allow the user to initialize the tenant")

	// Notifications
	ListNotifications        = permWithOneContext[UserID]("listNotifications", "Allow the user to list notifications")
	CountUnreadNotifications = permWithOneContext[UserID]("countUnreadNotifications", "Allow the user to count unread notifications")
	MarkNotificationAsRead   = permWithOneContext[UserID]("markNotificationAsRead", "Allow the user to mark a notification as read")

	// Users
	ListUsers       = permWithNoContext("listUsers", "Allow the user to list users")
	SearchUsers     = permWithNoContext("searchUsers", "Allow the user to search users")
	CreateUser      = permWithNoContext("createUser", "Allow the user to create a user")
	UpdateUser      = permWithOneContext[UserID]("updateUser", "Allow the user to update a user")
	GetMyOwnUser    = permWithOneContext[UserID]("getMyOwnUser", "Allow the user to get his own user")
	UpdatePassword  = permWithOneContext[UserID]("updatePassword", "Allow the user to update his password")
	EnableTotp      = permWithOneContext[UserID]("enableTotp", "Allow the user to enable TOTP")
	ResetTotp       = permWithOneContext[UserID]("resetTotp", "Allow the user to reset TOTP")
	GetUser         = permWithOneContext[UserID]("getUser", "Allow the user to get a user")
	DeleteUser      = permWithOneContext[UserID]("deleteUser", "Allow the user to delete a user")
	ResetPassword   = permWithOneContext[UserID]("resetPassword", "Allow the user to reset a user's password")
	ManageUserRoles = permWithOneContext[UserID]("manageUserRoles", "Allow the user to manage user roles")
	AuditUser       = permWithOneContext[UserID]("auditUser", "Allow the user to audit users")

	// Platform
	GetPlatformSettings = permWithNoContext("getPlatformSettings", "Allow the user to get platform settings")
	SetPlatformSettings = permWithNoContext("setPlatformSettings", "Allow the user to set platform settings")
	AuditPlatform       = permWithNoContext("auditPlatform", "Allow the user to audit the platform")
	ManageDynamicRoles  = permWithNoContext("manageDynamicRoles", "Allow the user to create dynamic roles")

	// App instances
	ListAppInstances  = permWithNoContext("listAppInstances", "Allow the user to list app instances")
	CreateAppInstance = permWithOneContext[WorkbenchID]("createAppInstance", "Allow the user to create an app instance")
	UpdateAppInstance = permWithOneContext[WorkbenchID]("updateAppInstance", "Allow the user to update an app instance")
	GetAppInstance    = permWithOneContext[WorkbenchID]("getAppInstance", "Allow the user to get an app instance")
	DeleteAppInstance = permWithOneContext[WorkbenchID]("deleteAppInstance", "Allow the user to delete an app instance")

	// Workbenches
	ListWorkbenchs         = permWithOneContext[WorkbenchID]("listWorkbenchs", "Allow the user to list workbenchs")
	CreateWorkbench        = permWithOneContext[WorkspaceID]("createWorkbench", "Allow the user to create a workbench")
	UpdateWorkbench        = permWithOneContext[WorkbenchID]("updateWorkbench", "Allow the user to update a workbench")
	GetWorkbench           = permWithOneContext[WorkbenchID]("getWorkbench", "Allow the user to get a workbench")
	StreamWorkbench        = permWithOneContext[WorkbenchID]("streamWorkbench", "Allow the user to stream a workbench")
	DeleteWorkbench        = permWithOneContext[WorkbenchID]("deleteWorkbench", "Allow the user to delete a workbench")
	AuditWorkbench         = permWithOneContext[WorkbenchID]("auditWorkbench", "Allow the user to audit a workbench")
	ManageUsersInWorkbench = permWithOneContext[WorkbenchID]("manageUsersInWorkbench", "Allow the user to manage users in a workbench")

	// Workspaces
	ListWorkspaces                 = permWithOneContext[WorkspaceID]("listWorkspaces", "Allow the user to list workspaces")
	ListPublicWorkspaces           = permWithNoContext("listPublicWorkspaces", "Allow the user to list public workspaces")
	CreateWorkspace                = permWithNoContext("createWorkspace", "Allow the user to create a workspace")
	UpdateWorkspace                = permWithOneContext[WorkspaceID]("updateWorkspace", "Allow the user to update a workspace")
	GetWorkspace                   = permWithOneContext[WorkspaceID]("getWorkspace", "Allow the user to get a workspace")
	DeleteWorkspace                = permWithOneContext[WorkspaceID]("deleteWorkspace", "Allow the user to delete a workspace")
	ManageUsersInWorkspace         = permWithOneContext[WorkspaceID]("manageUsersInWorkspace", "Allow the user to manage users in a workspace")
	ManageUsersDataRoleInWorkspace = permWithOneContext[WorkspaceID]("manageUsersDataRoleInWorkspace", "Allow the user to manage users' data role in a workspace")
	ListFilesInWorkspace           = permWithOneContext[WorkspaceID]("listFilesInWorkspace", "Allow the user to list files in a workspace")
	UploadFilesToWorkspace         = permWithOneContext[WorkspaceID]("uploadFilesToWorkspace", "Allow the user to upload files to a workspace")
	DownloadFilesFromWorkspace     = permWithOneContext[WorkspaceID]("downloadFilesFromWorkspace", "Allow the user to download files from a workspace")
	ModifyFilesInWorkspace         = permWithOneContext[WorkspaceID]("modifyFilesInWorkspace", "Allow the user to modify files in a workspace")
	AuditWorkspace                 = permWithOneContext[WorkspaceID]("auditWorkspace", "Allow the user to audit a workspace")

	// Workspace service instances
	ListWorkspaceServiceInstances     = permWithOneContext[WorkspaceID]("listWorkspaceServiceInstances", "Allow the user to list workspace service instances")
	CreateWorkspaceServiceInstance    = permWithOneContext[WorkspaceID]("createWorkspaceServiceInstance", "Allow the user to create a workspace service instance")
	UpdateWorkspaceServiceInstance    = permWithOneContext[WorkspaceID]("updateWorkspaceServiceInstance", "Allow the user to update a workspace service instance")
	GetWorkspaceServiceInstance       = permWithOneContext[WorkspaceID]("getWorkspaceServiceInstance", "Allow the user to get a workspace service instance")
	GetWorkspaceServiceInstanceSecret = permWithOneContext[WorkspaceID]("getWorkspaceServiceInstanceSecret", "Allow the user to get the secrets of a workspace service instance")
	DeleteWorkspaceServiceInstance    = permWithOneContext[WorkspaceID]("deleteWorkspaceServiceInstance", "Allow the user to delete a workspace service instance")

	// Apps
	ListApps  = permWithNoContext("listApps", "Allow the user to list apps")
	CreateApp = permWithNoContext("createApp", "Allow the user to create an app")
	UpdateApp = permWithNoContext("updateApp", "Allow the user to update an app")
	GetApp    = permWithNoContext("getApp", "Allow the user to get an app")
	DeleteApp = permWithNoContext("deleteApp", "Allow the user to delete an app")

	// Requests
	ListRequests   = permWithOneContext[WorkspaceID]("listRequests", "Allow the user to list requests")
	ListMyRequests = permWithNoContext("listMyRequests", "Allow the user to list his requests")
	GetRequest     = permWithOneContext[WorkspaceID]("getRequest", "Allow the user to get a request")
	CreateRequest  = permWithOneContext[WorkspaceID]("createRequest", "Allow the user to create a request")
	ApproveRequest = permWithOneContext[WorkspaceID]("approveRequest", "Allow the user to approve a request")
	DeleteRequest  = permWithOneContext[WorkspaceID]("deleteRequest", "Allow the user to delete a request")

	// Terms of use
	CreateTermsOfUseVersion     = permWithNoContext("createTermsOfUseVersion", "Allow the user to create a terms of use version")
	UpdateTermsOfUseVersion     = permWithNoContext("updateTermsOfUseVersion", "Allow the user to update a terms of use version")
	PublishTermsOfUseVersion    = permWithNoContext("publishTermsOfUseVersion", "Allow the user to publish a terms of use version")
	GetTermsOfUseVersion        = permWithNoContext("getTermsOfUseVersion", "Allow the user to get a terms of use version")
	ListTermsOfUseVersions      = permWithNoContext("listTermsOfUseVersions", "Allow the user to list terms of use versions")
	GetCurrentTermsOfUseVersion = permWithNoContext("getCurrentTermsOfUseVersion", "Allow the user to get the current terms of use version")
	ListTermsOfUseAcceptances   = permWithOneContext[UserID]("listTermsOfUseAcceptances", "Allow the user to list terms of use acceptances")
	GetMyTermsOfUseStatus       = permWithOneContext[UserID]("getMyTermsOfUseStatus", "Allow the user to get his terms of use acceptance status")
	AcceptTermsOfUse            = permWithOneContext[UserID]("acceptTermsOfUse", "Allow the user to accept the terms of use")

	// Organizations
	ListOrganizations  = permWithNoContext("listOrganizations", "Allow the user to list organizations")
	GetOrganization    = permWithNoContext("getOrganization", "Allow the user to get an organization")
	CreateOrganization = permWithNoContext("createOrganization", "Allow the user to create an organization")
	UpdateOrganization = permWithNoContext("updateOrganization", "Allow the user to update an organization")
	DeleteOrganization = permWithNoContext("deleteOrganization", "Allow the user to delete an organization")
)
