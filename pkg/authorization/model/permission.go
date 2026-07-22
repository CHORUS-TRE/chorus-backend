package model

import (
	"fmt"
	"slices"
)

// Permission is a permission check target: a PermissionName plus the concrete
// context it applies to. See PermissionDefinition for the schema counterpart.
type Permission struct {
	Name    PermissionName
	Context Context
}

func NewPermission(name PermissionName, opts ...NewContextOption) Permission {
	context := NewContext(opts...)
	return Permission{
		Name:    name,
		Context: context,
	}
}

type PermissionName string

const (
	PermissionListAppInstances  PermissionName = "listAppInstances"
	PermissionCreateAppInstance PermissionName = "createAppInstance"
	PermissionUpdateAppInstance PermissionName = "updateAppInstance"
	PermissionGetAppInstance    PermissionName = "getAppInstance"
	PermissionDeleteAppInstance PermissionName = "deleteAppInstance"

	PermissionListWorkbenchs         PermissionName = "listWorkbenchs"
	PermissionCreateWorkbench        PermissionName = "createWorkbench"
	PermissionUpdateWorkbench        PermissionName = "updateWorkbench"
	PermissionGetWorkbench           PermissionName = "getWorkbench"
	PermissionStreamWorkbench        PermissionName = "streamWorkbench"
	PermissionDeleteWorkbench        PermissionName = "deleteWorkbench"
	PermissionManageUsersInWorkbench PermissionName = "manageUsersInWorkbench"
	PermissionAuditWorkbench         PermissionName = "auditWorkbench"

	PermissionListWorkspaces                 PermissionName = "listWorkspaces"
	PermissionListPublicWorkspaces           PermissionName = "listPublicWorkspaces"
	PermissionCreateWorkspace                PermissionName = "createWorkspace"
	PermissionUpdateWorkspace                PermissionName = "updateWorkspace"
	PermissionGetWorkspace                   PermissionName = "getWorkspace"
	PermissionDeleteWorkspace                PermissionName = "deleteWorkspace"
	PermissionManageUsersInWorkspace         PermissionName = "manageUsersInWorkspace"
	PermissionManageUsersDataRoleInWorkspace PermissionName = "manageUsersDataRoleInWorkspace"
	PermissionListFilesInWorkspace           PermissionName = "listFilesInWorkspace"
	PermissionUploadFilesToWorkspace         PermissionName = "uploadFilesToWorkspace"
	PermissionDownloadFilesFromWorkspace     PermissionName = "downloadFilesFromWorkspace"
	PermissionModifyFilesInWorkspace         PermissionName = "modifyFilesInWorkspace"
	PermissionAuditWorkspace                 PermissionName = "auditWorkspace"

	PermissionListWorkspaceServiceInstances  PermissionName = "listWorkspaceServiceInstances"
	PermissionCreateWorkspaceServiceInstance PermissionName = "createWorkspaceServiceInstance"
	PermissionUpdateWorkspaceServiceInstance PermissionName = "updateWorkspaceServiceInstance"
	PermissionGetWorkspaceServiceInstance    PermissionName = "getWorkspaceServiceInstance"
	PermissionDeleteWorkspaceServiceInstance PermissionName = "deleteWorkspaceServiceInstance"

	PermissionGetWorkspaceServiceInstanceSecret PermissionName = "getWorkspaceServiceInstanceSecret"

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

	PermissionListUsers       PermissionName = "listUsers"
	PermissionSearchUsers     PermissionName = "searchUsers"
	PermissionCreateUser      PermissionName = "createUser"
	PermissionUpdateUser      PermissionName = "updateUser"
	PermissionGetMyOwnUser    PermissionName = "getMyOwnUser"
	PermissionUpdatePassword  PermissionName = "updatePassword"
	PermissionEnableTotp      PermissionName = "enableTotp"
	PermissionResetTotp       PermissionName = "resetTotp"
	PermissionGetUser         PermissionName = "getUser"
	PermissionDeleteUser      PermissionName = "deleteUser"
	PermissionResetPassword   PermissionName = "resetPassword"
	PermissionManageUserRoles PermissionName = "manageUserRoles"
	PermissionAuditUser       PermissionName = "auditUser"

	PermissionGetPlatformSettings PermissionName = "getPlatformSettings"
	PermissionSetPlatformSettings PermissionName = "setPlatformSettings"
	PermissionAuditPlatform       PermissionName = "auditPlatform"
	PermissionManageDynamicRoles  PermissionName = "manageDynamicRoles"

	PermissionCreateTermsOfUseVersion     PermissionName = "createTermsOfUseVersion"
	PermissionUpdateTermsOfUseVersion     PermissionName = "updateTermsOfUseVersion"
	PermissionPublishTermsOfUseVersion    PermissionName = "publishTermsOfUseVersion"
	PermissionGetTermsOfUseVersion        PermissionName = "getTermsOfUseVersion"
	PermissionListTermsOfUseVersions      PermissionName = "listTermsOfUseVersions"
	PermissionGetCurrentTermsOfUseVersion PermissionName = "getCurrentTermsOfUseVersion"
	PermissionListTermsOfUseAcceptances   PermissionName = "listTermsOfUseAcceptances"
	PermissionGetMyTermsOfUseStatus       PermissionName = "getMyTermsOfUseStatus"
	PermissionAcceptTermsOfUse            PermissionName = "acceptTermsOfUse"

	PermissionListRequests   PermissionName = "listRequests"
	PermissionListMyRequests PermissionName = "listMyRequests"
	PermissionGetRequest     PermissionName = "getRequest"
	PermissionCreateRequest  PermissionName = "createRequest"
	PermissionApproveRequest PermissionName = "approveRequest"
	PermissionDeleteRequest  PermissionName = "deleteRequest"

	PermissionListOrganizations  PermissionName = "listOrganizations"
	PermissionGetOrganization    PermissionName = "getOrganization"
	PermissionCreateOrganization PermissionName = "createOrganization"
	PermissionUpdateOrganization PermissionName = "updateOrganization"
	PermissionDeleteOrganization PermissionName = "deleteOrganization"
)

func (p PermissionName) String() string {
	return string(p)
}

// permissionNameIndex is derived from the default schema, which declares every
// permission exactly once. A permission absent from the schema is not resolvable.
var permissionNameIndex = func() map[string]PermissionName {
	schema := GetDefaultSchema()
	index := make(map[string]PermissionName, len(schema.Permissions))
	for _, permission := range schema.Permissions {
		index[string(permission.Name)] = permission.Name
	}
	return index
}()

func ToPermissionName(p string) (PermissionName, error) {
	if name, ok := permissionNameIndex[p]; ok {
		return name, nil
	}
	return "", fmt.Errorf("unknown permission type: %s", p)
}

func ToPermission(p string, c map[string]string) (Permission, error) {
	permissionName, err := ToPermissionName(p)
	if err != nil {
		return Permission{}, err
	}

	ctx := make(Context)
	for k, v := range c {
		cd, err := ToContextDimension(k)
		if err != nil {
			return Permission{}, fmt.Errorf("invalid context dimension in permission: %s", err)
		}
		ctx[cd] = v
	}

	return Permission{
		Name:    permissionName,
		Context: ctx,
	}, nil
}

func (p Permission) String() string {
	if len(p.Context) == 0 {
		return p.Name.String()
	}

	return fmt.Sprintf("%s@%s", p.Name, p.Context.String())
}

func (p Permission) Copy() Permission {
	newContext := make(Context, len(p.Context))
	for k, v := range p.Context {
		newContext[k] = v
	}
	return Permission{
		Name:    p.Name,
		Context: newContext,
	}
}

// UniquePermissionNames returns a sorted, deduplicated list of permission names from a slice of permissions.
func UniquePermissionNames(permissions []Permission) []string {
	seen := make(map[string]struct{}, len(permissions))
	for _, p := range permissions {
		seen[string(p.Name)] = struct{}{}
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}
