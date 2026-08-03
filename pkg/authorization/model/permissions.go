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

type permFactory0 struct{ PermissionDefinition }

func (p permFactory0) For() Permission { return Permission{Name: p.Name} }

type permFactory1[A contextID] struct{ PermissionDefinition }

func (p permFactory1[A]) For(a A) Permission {
	return Permission{Name: p.Name, Context: Context{a.dimension(): fmt.Sprint(a)}}
}

type permFactory2[A, B contextID] struct{ PermissionDefinition }

func (p permFactory2[A, B]) For(a A, b B) Permission {
	return Permission{Name: p.Name, Context: Context{
		a.dimension(): fmt.Sprint(a),
		b.dimension(): fmt.Sprint(b),
	}}
}

func perm0(name PermissionName, description string) permFactory0 {
	return permFactory0{registerPermission(name, description)}
}

func perm1[A contextID](name PermissionName, description string) permFactory1[A] {
	var a A
	return permFactory1[A]{registerPermission(name, description, a.dimension())}
}

func perm2[A, B contextID](name PermissionName, description string) permFactory2[A, B] {
	var a A
	var b B
	return permFactory2[A, B]{registerPermission(name, description, a.dimension(), b.dimension())}
}

// One factory per permission — name, description, and required context (via the
// id type).
var (
	// Authentication
	Authenticate                       = perm0("authenticate", "Allow the user to authenticate")
	Logout                             = perm0("logout", "Allow the user to logout")
	RefreshToken                       = perm0("refreshToken", "Allow the user to refresh the jwt token")
	GetListOfPossibleWayToAuthenticate = perm0("getListOfPossibleWayToAuthenticate", "Allow the user to get a list of possible ways to authenticate")
	AuthenticateUsingAuth2_0           = perm0("authenticateUsingAuth2.0", "Allow the user to authenticate using oauth2")
	AuthenticateRedirectUsingAuth2_0   = perm0("authenticateRedirectUsingAuth2.0", "Allow the user to be redirected after authenticating using oauth2")

	// Health
	GetHealthCheck = perm0("getHealthCheck", "Allow the user to get the healthcheck status")

	// Tenant
	InitializeTenant = perm0("initializeTenant", "Allow the user to initialize the tenant")

	// Notifications
	ListNotifications        = perm1[UserID]("listNotifications", "Allow the user to list notifications")
	CountUnreadNotifications = perm1[UserID]("countUnreadNotifications", "Allow the user to count unread notifications")
	MarkNotificationAsRead   = perm1[UserID]("markNotificationAsRead", "Allow the user to mark a notification as read")

	// Users
	ListUsers       = perm0("listUsers", "Allow the user to list users")
	SearchUsers     = perm0("searchUsers", "Allow the user to search users")
	CreateUser      = perm0("createUser", "Allow the user to create a user")
	UpdateUser      = perm1[UserID]("updateUser", "Allow the user to update a user")
	GetMyOwnUser    = perm1[UserID]("getMyOwnUser", "Allow the user to get his own user")
	UpdatePassword  = perm1[UserID]("updatePassword", "Allow the user to update his password")
	EnableTotp      = perm1[UserID]("enableTotp", "Allow the user to enable TOTP")
	ResetTotp       = perm1[UserID]("resetTotp", "Allow the user to reset TOTP")
	GetUser         = perm1[UserID]("getUser", "Allow the user to get a user")
	DeleteUser      = perm1[UserID]("deleteUser", "Allow the user to delete a user")
	ResetPassword   = perm1[UserID]("resetPassword", "Allow the user to reset a user's password")
	ManageUserRoles = perm1[UserID]("manageUserRoles", "Allow the user to manage user roles")
	AuditUser       = perm1[UserID]("auditUser", "Allow the user to audit users")

	// Platform
	GetPlatformSettings = perm0("getPlatformSettings", "Allow the user to get platform settings")
	SetPlatformSettings = perm0("setPlatformSettings", "Allow the user to set platform settings")
	AuditPlatform       = perm0("auditPlatform", "Allow the user to audit the platform")
	ManageDynamicRoles  = perm0("manageDynamicRoles", "Allow the user to create dynamic roles")

	// App instances
	ListAppInstances  = perm0("listAppInstances", "Allow the user to list app instances")
	CreateAppInstance = perm1[WorkbenchID]("createAppInstance", "Allow the user to create an app instance")
	UpdateAppInstance = perm1[WorkbenchID]("updateAppInstance", "Allow the user to update an app instance")
	GetAppInstance    = perm1[WorkbenchID]("getAppInstance", "Allow the user to get an app instance")
	DeleteAppInstance = perm1[WorkbenchID]("deleteAppInstance", "Allow the user to delete an app instance")

	// Workbenches
	ListWorkbenchs         = perm1[WorkbenchID]("listWorkbenchs", "Allow the user to list workbenchs")
	CreateWorkbench        = perm1[WorkspaceID]("createWorkbench", "Allow the user to create a workbench")
	UpdateWorkbench        = perm1[WorkbenchID]("updateWorkbench", "Allow the user to update a workbench")
	GetWorkbench           = perm1[WorkbenchID]("getWorkbench", "Allow the user to get a workbench")
	StreamWorkbench        = perm1[WorkbenchID]("streamWorkbench", "Allow the user to stream a workbench")
	DeleteWorkbench        = perm1[WorkbenchID]("deleteWorkbench", "Allow the user to delete a workbench")
	AuditWorkbench         = perm1[WorkbenchID]("auditWorkbench", "Allow the user to audit a workbench")
	ManageUsersInWorkbench = perm1[WorkbenchID]("manageUsersInWorkbench", "Allow the user to manage users in a workbench")

	// Workspaces
	ListWorkspaces                 = perm1[WorkspaceID]("listWorkspaces", "Allow the user to list workspaces")
	ListPublicWorkspaces           = perm0("listPublicWorkspaces", "Allow the user to list public workspaces")
	CreateWorkspace                = perm0("createWorkspace", "Allow the user to create a workspace")
	UpdateWorkspace                = perm1[WorkspaceID]("updateWorkspace", "Allow the user to update a workspace")
	GetWorkspace                   = perm1[WorkspaceID]("getWorkspace", "Allow the user to get a workspace")
	DeleteWorkspace                = perm1[WorkspaceID]("deleteWorkspace", "Allow the user to delete a workspace")
	ManageUsersInWorkspace         = perm1[WorkspaceID]("manageUsersInWorkspace", "Allow the user to manage users in a workspace")
	ManageUsersDataRoleInWorkspace = perm1[WorkspaceID]("manageUsersDataRoleInWorkspace", "Allow the user to manage users' data role in a workspace")
	ListFilesInWorkspace           = perm1[WorkspaceID]("listFilesInWorkspace", "Allow the user to list files in a workspace")
	UploadFilesToWorkspace         = perm1[WorkspaceID]("uploadFilesToWorkspace", "Allow the user to upload files to a workspace")
	DownloadFilesFromWorkspace     = perm1[WorkspaceID]("downloadFilesFromWorkspace", "Allow the user to download files from a workspace")
	ModifyFilesInWorkspace         = perm1[WorkspaceID]("modifyFilesInWorkspace", "Allow the user to modify files in a workspace")
	AuditWorkspace                 = perm1[WorkspaceID]("auditWorkspace", "Allow the user to audit a workspace")

	// Workspace service instances
	ListWorkspaceServiceInstances     = perm1[WorkspaceID]("listWorkspaceServiceInstances", "Allow the user to list workspace service instances")
	CreateWorkspaceServiceInstance    = perm1[WorkspaceID]("createWorkspaceServiceInstance", "Allow the user to create a workspace service instance")
	UpdateWorkspaceServiceInstance    = perm1[WorkspaceID]("updateWorkspaceServiceInstance", "Allow the user to update a workspace service instance")
	GetWorkspaceServiceInstance       = perm1[WorkspaceID]("getWorkspaceServiceInstance", "Allow the user to get a workspace service instance")
	GetWorkspaceServiceInstanceSecret = perm1[WorkspaceID]("getWorkspaceServiceInstanceSecret", "Allow the user to get the secrets of a workspace service instance")
	DeleteWorkspaceServiceInstance    = perm1[WorkspaceID]("deleteWorkspaceServiceInstance", "Allow the user to delete a workspace service instance")

	// Apps
	ListApps  = perm0("listApps", "Allow the user to list apps")
	CreateApp = perm0("createApp", "Allow the user to create an app")
	UpdateApp = perm0("updateApp", "Allow the user to update an app")
	GetApp    = perm0("getApp", "Allow the user to get an app")
	DeleteApp = perm0("deleteApp", "Allow the user to delete an app")

	// Requests
	ListRequests   = perm1[WorkspaceID]("listRequests", "Allow the user to list requests")
	ListMyRequests = perm0("listMyRequests", "Allow the user to list his requests")
	GetRequest     = perm1[WorkspaceID]("getRequest", "Allow the user to get a request")
	CreateRequest  = perm1[WorkspaceID]("createRequest", "Allow the user to create a request")
	ApproveRequest = perm1[WorkspaceID]("approveRequest", "Allow the user to approve a request")
	DeleteRequest  = perm1[WorkspaceID]("deleteRequest", "Allow the user to delete a request")

	// Terms of use
	CreateTermsOfUseVersion     = perm0("createTermsOfUseVersion", "Allow the user to create a terms of use version")
	UpdateTermsOfUseVersion     = perm0("updateTermsOfUseVersion", "Allow the user to update a terms of use version")
	PublishTermsOfUseVersion    = perm0("publishTermsOfUseVersion", "Allow the user to publish a terms of use version")
	GetTermsOfUseVersion        = perm0("getTermsOfUseVersion", "Allow the user to get a terms of use version")
	ListTermsOfUseVersions      = perm0("listTermsOfUseVersions", "Allow the user to list terms of use versions")
	GetCurrentTermsOfUseVersion = perm0("getCurrentTermsOfUseVersion", "Allow the user to get the current terms of use version")
	ListTermsOfUseAcceptances   = perm1[UserID]("listTermsOfUseAcceptances", "Allow the user to list terms of use acceptances")
	GetMyTermsOfUseStatus       = perm1[UserID]("getMyTermsOfUseStatus", "Allow the user to get his terms of use acceptance status")
	AcceptTermsOfUse            = perm1[UserID]("acceptTermsOfUse", "Allow the user to accept the terms of use")

	// Organizations
	ListOrganizations  = perm0("listOrganizations", "Allow the user to list organizations")
	GetOrganization    = perm0("getOrganization", "Allow the user to get an organization")
	CreateOrganization = perm0("createOrganization", "Allow the user to create an organization")
	UpdateOrganization = perm0("updateOrganization", "Allow the user to update an organization")
	DeleteOrganization = perm0("deleteOrganization", "Allow the user to delete an organization")
)
