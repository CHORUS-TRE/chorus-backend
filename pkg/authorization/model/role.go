package model

import (
	"fmt"
	"slices"
	"strings"

	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
)

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

func GetWorkspaceRoles() []RoleName {
	return []RoleName{
		RoleWorkspaceGuest,
		RoleWorkspaceMember,
		RoleWorkspaceMaintainer,
		RoleWorkspaceDataManager,
		RoleWorkspaceAdmin,
	}
}

func GetWorkbenchRoles() []RoleName {
	return []RoleName{
		RoleWorkbenchViewer,
		RoleWorkbenchMember,
		RoleWorkbenchAdmin,
	}
}

func RoleIn(role RoleName, roles []RoleName) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// UniqueRoleNames returns a sorted, deduplicated list of role names from a slice of JWT roles.
func UniqueRoleNames(roles []jwt_model.Role) []string {
	seen := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		seen[r.Name] = struct{}{}
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}
