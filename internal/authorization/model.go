package authorization

import "fmt"

type Permission string

const (
	PermissionListAppInstances  Permission = "listAppInstances"
	PermissionCreateAppInstance Permission = "createAppInstance"
	PermissionUpdateAppInstance Permission = "updateAppInstance"
	PermissionGetAppInstance    Permission = "getAppInstance"
	PermissionDeleteAppInstance Permission = "deleteAppInstance"

	PermissionListWorkbenchs    Permission = "listWorkbenchs"
	PermissionCreateWorkbench   Permission = "createWorkbench"
	PermissionUpdateWorkbench   Permission = "updateWorkbench"
	PermissionGetWorkbench      Permission = "getWorkbench"
	PermissionStreamWorkbench   Permission = "streamWorkbench"
	PermissionDeleteWorkbench   Permission = "deleteWorkbench"
	PermissionInviteInWorkbench Permission = "inviteInWorkbench"

	PermissionListWorkspaces    Permission = "listWorkspaces"
	PermissionCreateWorkspace   Permission = "createWorkspace"
	PermissionUpdateWorkspace   Permission = "updateWorkspace"
	PermissionGetWorkspace      Permission = "getWorkspace"
	PermissionDeleteWorkspace   Permission = "deleteWorkspace"
	PermissionInviteInWorkspace Permission = "inviteInWorkspace"

	PermissionListApps  Permission = "listApps"
	PermissionCreateApp Permission = "createApp"
	PermissionUpdateApp Permission = "updateApp"
	PermissionGetApp    Permission = "getApp"
	PermissionDeleteApp Permission = "deleteApp"

	PermissionAuthenticate                       Permission = "authenticate"
	PermissionLogout                             Permission = "logout"
	PermissionGetListOfPossibleWayToAuthenticate Permission = "getListOfPossibleWayToAuthenticate"
	PermissionAuthenticateUsingAuth2_0           Permission = "authenticateUsingAuth2.0"
	PermissionAuthenticateRedirectUsingAuth2_0   Permission = "authenticateRedirectUsingAuth2.0"
	PermissionRefreshToken                       Permission = "refreshToken"

	PermissionGetHealthCheck Permission = "getHealthCheck"

	PermissionListNotifications        Permission = "listNotifications"
	PermissionCountUnreadNotifications Permission = "countUnreadNotifications"
	PermissionMarkNotificationAsRead   Permission = "markNotificationAsRead"

	PermissionInitializeTenant Permission = "initializeTenant"

	PermissionListUsers      Permission = "listUsers"
	PermissionCreateUser     Permission = "createUser"
	PermissionUpdateUser     Permission = "updateUser"
	PermissionGetMyOwnUser   Permission = "getMyOwnUser"
	PermissionUpdatePassword Permission = "updatePassword"
	PermissionEnableTotp     Permission = "enableTotp"
	PermissionResetTotp      Permission = "resetTotp"
	PermissionGetUser        Permission = "getUser"
	PermissionDeleteUser     Permission = "deleteUser"
	PermissionResetPassword  Permission = "resetPassword"
)

func (p Permission) String() string {
	return string(p)
}

func ToPermission(p string) (Permission, error) {
	switch p {
	case string(PermissionListAppInstances):
		return PermissionListAppInstances, nil
	case string(PermissionCreateAppInstance):
		return PermissionCreateAppInstance, nil
	case string(PermissionUpdateAppInstance):
		return PermissionUpdateAppInstance, nil
	case string(PermissionGetAppInstance):
		return PermissionGetAppInstance, nil
	case string(PermissionDeleteAppInstance):
		return PermissionDeleteAppInstance, nil

	case string(PermissionListWorkbenchs):
		return PermissionListWorkbenchs, nil
	case string(PermissionCreateWorkbench):
		return PermissionCreateWorkbench, nil
	case string(PermissionUpdateWorkbench):
		return PermissionUpdateWorkbench, nil
	case string(PermissionGetWorkbench):
		return PermissionGetWorkbench, nil
	case string(PermissionDeleteWorkbench):
		return PermissionDeleteWorkbench, nil
	case string(PermissionInviteInWorkbench):
		return PermissionInviteInWorkbench, nil

	case string(PermissionListWorkspaces):
		return PermissionListWorkspaces, nil
	case string(PermissionCreateWorkspace):
		return PermissionCreateWorkspace, nil
	case string(PermissionUpdateWorkspace):
		return PermissionUpdateWorkspace, nil
	case string(PermissionGetWorkspace):
		return PermissionGetWorkspace, nil
	case string(PermissionDeleteWorkspace):
		return PermissionDeleteWorkspace, nil
	case string(PermissionInviteInWorkspace):
		return PermissionInviteInWorkspace, nil

	case string(PermissionListApps):
		return PermissionListApps, nil
	case string(PermissionCreateApp):
		return PermissionCreateApp, nil
	case string(PermissionUpdateApp):
		return PermissionUpdateApp, nil
	case string(PermissionGetApp):
		return PermissionGetApp, nil
	case string(PermissionDeleteApp):
		return PermissionDeleteApp, nil

	case string(PermissionAuthenticate):
		return PermissionAuthenticate, nil
	case string(PermissionLogout):
		return PermissionLogout, nil
	case string(PermissionGetListOfPossibleWayToAuthenticate):
		return PermissionGetListOfPossibleWayToAuthenticate, nil
	case string(PermissionAuthenticateUsingAuth2_0):
		return PermissionAuthenticateUsingAuth2_0, nil
	case string(PermissionAuthenticateRedirectUsingAuth2_0):
		return PermissionAuthenticateRedirectUsingAuth2_0, nil
	case string(PermissionRefreshToken):
		return PermissionRefreshToken, nil

	case string(PermissionGetHealthCheck):
		return PermissionGetHealthCheck, nil

	case string(PermissionListNotifications):
		return PermissionListNotifications, nil
	case string(PermissionCountUnreadNotifications):
		return PermissionCountUnreadNotifications, nil
	case string(PermissionMarkNotificationAsRead):
		return PermissionMarkNotificationAsRead, nil

	case string(PermissionInitializeTenant):
		return PermissionInitializeTenant, nil

	case string(PermissionListUsers):
		return PermissionListUsers, nil
	case string(PermissionCreateUser):
		return PermissionCreateUser, nil
	case string(PermissionUpdateUser):
		return PermissionUpdateUser, nil
	case string(PermissionGetMyOwnUser):
		return PermissionGetMyOwnUser, nil
	case string(PermissionUpdatePassword):
		return PermissionUpdatePassword, nil
	case string(PermissionEnableTotp):
		return PermissionEnableTotp, nil
	case string(PermissionResetTotp):
		return PermissionResetTotp, nil
	case string(PermissionGetUser):
		return PermissionGetUser, nil
	case string(PermissionDeleteUser):
		return PermissionDeleteUser, nil
	case string(PermissionResetPassword):
		return PermissionResetPassword, nil
	}

	return "", fmt.Errorf("unknown permission type: %s", p)
}

type Role string

const (
	RolePublic               Role = "public"
	RoleAuthenticated        Role = "authenticated"
	RoleWorkspaceGuest       Role = "workspaceGuest"
	RoleWorkspaceMember      Role = "workspaceMember"
	RoleWorkspaceMaintainer  Role = "workspaceMaintainer"
	RoleWorkspaceAdmin       Role = "workspaceAdmin"
	RoleWorkbenchViewer      Role = "workbenchViewer"
	RoleWorkbenchMember      Role = "workbenchMember"
	RoleWorkbenchAdmin       Role = "workbenchAdmin"
	RoleHealthchecker        Role = "healthchecker"
	RolePlateformUserManager Role = "plateformUserManager"
	RoleAppStoreAdmin        Role = "appStoreAdmin"
	RoleSuperAdmin           Role = "superAdmin"
)

func (r Role) String() string {
	return string(r)
}

func ToRole(r string) (Role, error) {
	switch r {
	case string(RolePublic):
		return RolePublic, nil
	case string(RoleAuthenticated):
		return RoleAuthenticated, nil
	case string(RoleWorkspaceGuest):
		return RoleWorkspaceGuest, nil
	case string(RoleWorkspaceMember):
		return RoleWorkspaceMember, nil
	case string(RoleWorkspaceMaintainer):
		return RoleWorkspaceMaintainer, nil
	case string(RoleWorkspaceAdmin):
		return RoleWorkspaceAdmin, nil
	case string(RoleWorkbenchViewer):
		return RoleWorkbenchViewer, nil
	case string(RoleWorkbenchMember):
		return RoleWorkbenchMember, nil
	case string(RoleWorkbenchAdmin):
		return RoleWorkbenchAdmin, nil
	case string(RoleHealthchecker):
		return RoleHealthchecker, nil
	case string(RolePlateformUserManager):
		return RolePlateformUserManager, nil
	case string(RoleAppStoreAdmin):
		return RoleAppStoreAdmin, nil
	case string(RoleSuperAdmin):
		return RoleSuperAdmin, nil
	}

	return "", fmt.Errorf("unknown role type: %s", r)
}
