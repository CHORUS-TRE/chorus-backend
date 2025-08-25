package authorization

import (
	"context"
	"fmt"

	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
)

func NewPermission(name PermissionName, opts ...NewContextOption) Permission {
	context := NewContext(opts...)
	return Permission{
		Name:    name,
		Context: context,
	}
}

type Permission struct {
	Name    PermissionName
	Context Context
}

type PermissionName string

const (
	PermissionListAppInstances  PermissionName = "listAppInstances"
	PermissionCreateAppInstance PermissionName = "createAppInstance"
	PermissionUpdateAppInstance PermissionName = "updateAppInstance"
	PermissionGetAppInstance    PermissionName = "getAppInstance"
	PermissionDeleteAppInstance PermissionName = "deleteAppInstance"

	PermissionListWorkbenchs    PermissionName = "listWorkbenchs"
	PermissionCreateWorkbench   PermissionName = "createWorkbench"
	PermissionUpdateWorkbench   PermissionName = "updateWorkbench"
	PermissionGetWorkbench      PermissionName = "getWorkbench"
	PermissionStreamWorkbench   PermissionName = "streamWorkbench"
	PermissionDeleteWorkbench   PermissionName = "deleteWorkbench"
	PermissionInviteInWorkbench PermissionName = "inviteInWorkbench"

	PermissionListWorkspaces    PermissionName = "listWorkspaces"
	PermissionCreateWorkspace   PermissionName = "createWorkspace"
	PermissionUpdateWorkspace   PermissionName = "updateWorkspace"
	PermissionGetWorkspace      PermissionName = "getWorkspace"
	PermissionDeleteWorkspace   PermissionName = "deleteWorkspace"
	PermissionInviteInWorkspace PermissionName = "inviteInWorkspace"

	PermissionListApps  PermissionName = "listApps"
	PermissionCreateApp PermissionName = "createApp"
	PermissionUpdateApp PermissionName = "updateApp"
	PermissionGetApp    PermissionName = "getApp"
	PermissionDeleteApp PermissionName = "deleteApp"

	PermissionAuthenticate                       PermissionName = "authenticate"
	PermissionLogout                             PermissionName = "logout"
	PermissionGetListOfPossibleWayToAuthenticate PermissionName = "getListOfPossibleWayToAuthenticate"
	PermissionAuthenticateUsingAuth2_0           PermissionName = "authenticateUsingAuth2.0"
	PermissionAuthenticateRedirectUsingAuth2_0   PermissionName = "authenticateRedirectUsingAuth2.0"
	PermissionRefreshToken                       PermissionName = "refreshToken"

	PermissionGetHealthCheck PermissionName = "getHealthCheck"

	PermissionListNotifications        PermissionName = "listNotifications"
	PermissionCountUnreadNotifications PermissionName = "countUnreadNotifications"
	PermissionMarkNotificationAsRead   PermissionName = "markNotificationAsRead"

	PermissionInitializeTenant PermissionName = "initializeTenant"

	PermissionListUsers      PermissionName = "listUsers"
	PermissionCreateUser     PermissionName = "createUser"
	PermissionUpdateUser     PermissionName = "updateUser"
	PermissionGetMyOwnUser   PermissionName = "getMyOwnUser"
	PermissionUpdatePassword PermissionName = "updatePassword"
	PermissionEnableTotp     PermissionName = "enableTotp"
	PermissionResetTotp      PermissionName = "resetTotp"
	PermissionGetUser        PermissionName = "getUser"
	PermissionDeleteUser     PermissionName = "deleteUser"
	PermissionResetPassword  PermissionName = "resetPassword"
)

func (p PermissionName) String() string {
	return string(p)
}

func ToPermissionName(p string) (PermissionName, error) {
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

type NewContextOption func(*Context)

func WithWorkspace(workspace any) NewContextOption {
	return func(c *Context) {
		(*c)[RoleContextWorkspace] = fmt.Sprintf("%v", workspace)
	}
}

func WithWorkbench(workbench any) NewContextOption {
	return func(c *Context) {
		(*c)[RoleContextWorkbench] = fmt.Sprintf("%v", workbench)
	}
}

func WithUser(user any) NewContextOption {
	return func(c *Context) {
		(*c)[RoleContextUser] = fmt.Sprintf("%v", user)
	}
}

func WithUserFromCtx(ctx context.Context) NewContextOption {
	uID := ""
	f := func(c *Context) {
		(*c)[RoleContextUser] = uID
	}

	claims, ok := ctx.Value(jwt_model.JWTClaimsContextKey).(*jwt_model.JWTClaims)
	if !ok {
		return f
	}

	uID = fmt.Sprintf("%v", claims.ID)

	return f
}

func NewContext(opts ...NewContextOption) Context {
	c := make(Context)
	for _, v := range opts {
		v(&c)
	}
	return c
}

type Context map[ContextDimension]string

type Role struct {
	Name    RoleName
	Context Context
}

type RoleName string

const (
	RolePublic               RoleName = "Public"
	RoleAuthenticated        RoleName = "Authenticated"
	RoleWorkspaceGuest       RoleName = "WorkspaceGuest"
	RoleWorkspaceMember      RoleName = "WorkspaceMember"
	RoleWorkspaceMaintainer  RoleName = "WorkspaceMaintainer"
	RoleWorkspaceAdmin       RoleName = "WorkspaceAdmin"
	RoleWorkbenchViewer      RoleName = "WorkbenchViewer"
	RoleWorkbenchMember      RoleName = "WorkbenchMember"
	RoleWorkbenchAdmin       RoleName = "WorkbenchAdmin"
	RoleHealthchecker        RoleName = "Healthchecker"
	RolePlateformUserManager RoleName = "PlateformUserManager"
	RoleAppStoreAdmin        RoleName = "AppStoreAdmin"
	RoleSuperAdmin           RoleName = "SuperAdmin"
)

func (r RoleName) String() string {
	return string(r)
}

func ToRoleName(r string) (RoleName, error) {
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

type ContextDimension string

const (
	RoleContextWorkspace ContextDimension = "workspace"
	RoleContextWorkbench ContextDimension = "workbench"
	RoleContextUser      ContextDimension = "user"
)

func (r ContextDimension) String() string {
	return string(r)
}

func ToRoleContext(r string) (ContextDimension, error) {
	switch r {
	case string(RoleContextWorkspace):
		return RoleContextWorkspace, nil
	case string(RoleContextWorkbench):
		return RoleContextWorkbench, nil
	case string(RoleContextUser):
		return RoleContextUser, nil
	}

	return "", fmt.Errorf("unknown role context type: %s", r)
}
