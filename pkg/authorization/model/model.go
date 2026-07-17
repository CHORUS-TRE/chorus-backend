package model

// FindUsersWithPermissionFilter narrows a user search to those granted a
// permission, optionally in a specific context or via specific roles.
type FindUsersWithPermissionFilter struct {
	PermissionName          PermissionName
	Context                 Context
	ViaRoles                []RoleName
	PreferExactContextMatch bool
}
