package model

import (
	"fmt"
	"slices"
	"strings"
	"sync"
)

// Typed resource ids. Each knows its context dimension, so a permission or role
// derives its dimension from the id type — no separate ContextDimension to pass
// or keep in sync.
type (
	WorkspaceID uint64
	WorkbenchID uint64
	UserID      uint64
)

func (WorkspaceID) dimension() ContextDimension { return ContextWorkspace }
func (WorkbenchID) dimension() ContextDimension { return ContextWorkbench }
func (UserID) dimension() ContextDimension      { return ContextUser }

// contextID is a typed resource id that knows its dimension.
type contextID interface {
	~uint64
	dimension() ContextDimension
}

// PermissionDefinition declares a permission and the context dimensions a check
// against it requires. See Permission for the contextual check counterpart.
type PermissionDefinition struct {
	Name                      PermissionName
	Description               string
	RequiredContextDimensions []ContextDimension
}

// RoleDefinition declares what a role grants: the permissions it confers and the
// context dimensions an assignment must bind. See Role for the assignment
// counterpart carried by users.
type RoleDefinition struct {
	Name                      RoleName
	Description               string
	Scope                     RoleScope
	Dynamic                   bool
	RequiredContextDimensions map[ContextDimension]ContextQuantifier
	Permissions               []PermissionName
}

// FindUsersWithPermissionFilter narrows a user search to those granted a
// permission, optionally in a specific context or via specific roles.
type FindUsersWithPermissionFilter struct {
	PermissionName          PermissionName
	Context                 Context
	ViaRoles                []RoleName
	PreferExactContextMatch bool
}

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

// permissionNameIndex resolves a raw string to its declared PermissionName; a
// permission absent from the registry is not resolvable. It is built lazily on
// first use — the permission registry is populated by package-level factory
// vars, so it is only complete once package initialization has finished.
var permissionNameIndex = sync.OnceValue(func() map[string]PermissionName {
	index := make(map[string]PermissionName, len(PermissionDefinitions()))
	for _, def := range PermissionDefinitions() {
		index[string(def.Name)] = def.Name
	}
	return index
})

func ToPermissionName(p string) (PermissionName, error) {
	if name, ok := permissionNameIndex()[p]; ok {
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

// Role is a role assignment: a RoleName plus the concrete context it applies
// to. It is what users carry (e.g. in the JWT). See RoleDefinition for the
// schema counterpart declaring what a role grants.
type Role struct {
	Name    RoleName `json:"name"`
	Context Context  `json:"context"`
}

func NewRole(name RoleName, opts ...NewContextOption) Role {
	context := NewContext(opts...)
	return Role{
		Name:    name,
		Context: context,
	}
}

func ToRole(name string, context map[string]string) (Role, error) {
	roleName, err := ToRoleName(name)
	if err != nil {
		return Role{}, err
	}

	ctx := make(Context)
	for k, v := range context {
		ctx[ContextDimension(k)] = v
	}

	return Role{
		Name:    roleName,
		Context: ctx,
	}, nil
}

func (r Role) String() string {
	if len(r.Context) == 0 {
		return r.Name.String()
	}

	return fmt.Sprintf("%s@%s", r.Name, r.Context.String())
}

type RoleName string

const (
	RolePublic                      RoleName = "Public"
	RoleAuthenticated               RoleName = "Authenticated"
	RoleWorkspaceGuest              RoleName = "WorkspaceGuest"
	RoleWorkspaceMember             RoleName = "WorkspaceMember"
	RoleWorkspaceMaintainer         RoleName = "WorkspaceMaintainer"
	RoleWorkspaceDataManager        RoleName = "WorkspaceDataManager"
	RoleWorkspaceAdmin              RoleName = "WorkspaceAdmin"
	RoleWorkbenchViewer             RoleName = "WorkbenchViewer"
	RoleWorkbenchMember             RoleName = "WorkbenchMember"
	RoleWorkbenchAdmin              RoleName = "WorkbenchAdmin"
	RoleHealthchecker               RoleName = "Healthchecker"
	RolePlatformSettingsManager     RoleName = "PlatformSettingsManager"
	RolePlatformUserManager         RoleName = "PlatformUserManager"
	RolePlatformOrganizationManager RoleName = "PlatformOrganizationManager"
	RolePlatformAuditor             RoleName = "PlatformAuditor"
	RolePlatformWorkspaceManager    RoleName = "PlatformWorkspaceManager"
	RoleAppStoreAdmin               RoleName = "AppStoreAdmin"
	RolePlatformDataManager         RoleName = "PlatformDataManager"
	RoleSuperAdmin                  RoleName = "SuperAdmin"
)

func (r RoleName) String() string {
	return string(r)
}

// systemRoleIndex resolves the system roles by name; unknown names are
// accepted by ToRoleName as dynamic roles.
var systemRoleIndex = func() map[string]RoleName {
	index := make(map[string]RoleName)
	for _, role := range GetAllRoles() {
		index[string(role)] = role
	}
	return index
}()

func ToRoleName(r string) (RoleName, error) {
	if name, ok := systemRoleIndex[r]; ok {
		return name, nil
	}

	if strings.TrimSpace(r) == "" {
		return "", fmt.Errorf("empty role type")
	}

	return RoleName(r), nil
}

func IsSystemRole(role RoleName) bool {
	_, ok := systemRoleIndex[string(role)]
	return ok
}

func GetAllRoles() []RoleName {
	return []RoleName{
		RolePublic,
		RoleAuthenticated,
		RoleWorkspaceGuest,
		RoleWorkspaceMember,
		RoleWorkspaceMaintainer,
		RoleWorkspaceDataManager,
		RoleWorkspaceAdmin,
		RoleWorkbenchViewer,
		RoleWorkbenchMember,
		RoleWorkbenchAdmin,
		RoleHealthchecker,
		RolePlatformSettingsManager,
		RolePlatformUserManager,
		RolePlatformOrganizationManager,
		RolePlatformAuditor,
		RolePlatformWorkspaceManager,
		RoleAppStoreAdmin,
		RolePlatformDataManager,
		RoleSuperAdmin,
	}
}

// AuthorizationSchema is the full authorization model in effect: every
// permission (always code-defined) and every role (the code defaults plus any
// dynamic roles loaded from the store).
type AuthorizationSchema struct {
	Roles       []*RoleDefinition
	Permissions []PermissionDefinition
}

// GetDefaultSchema returns the code-defined schema — the permission and role
// definitions declared by the factories in permissions.go / roles.go. It hands
// back fresh copies so a caller may adjust a role (e.g. a config-driven grant)
// without mutating the registries.
func GetDefaultSchema() AuthorizationSchema {
	roles := make([]*RoleDefinition, len(RoleDefinitions()))
	for i, def := range RoleDefinitions() {
		clone := *def
		clone.Permissions = append([]PermissionName(nil), def.Permissions...)
		roles[i] = &clone
	}
	return AuthorizationSchema{
		Permissions: append([]PermissionDefinition(nil), PermissionDefinitions()...),
		Roles:       roles,
	}
}

type RoleScope string

const (
	RoleScopeSystem    RoleScope = "system"
	RoleScopePlatform  RoleScope = "platform"
	RoleScopeWorkspace RoleScope = "workspace"
	RoleScopeWorkbench RoleScope = "workbench"
)

func (s RoleScope) String() string {
	return string(s)
}

func ToRoleScope(scope string) (RoleScope, error) {
	switch scope {
	case string(RoleScopeSystem):
		return RoleScopeSystem, nil
	case string(RoleScopePlatform):
		return RoleScopePlatform, nil
	case string(RoleScopeWorkspace):
		return RoleScopeWorkspace, nil
	case string(RoleScopeWorkbench):
		return RoleScopeWorkbench, nil
	}
	return "", fmt.Errorf("unknown role scope: %s", scope)
}
