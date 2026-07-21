package model

func GetDefaultSchema() AuthorizationSchema {
	publicPermissions := permissionList(
		[]PermissionName{
			PermissionAuthenticate,
			PermissionGetListOfPossibleWayToAuthenticate,
			PermissionAuthenticateUsingAuth2_0,
			PermissionAuthenticateRedirectUsingAuth2_0,
			PermissionGetPlatformSettings,
		},
	)

	authenticatedPermissions := permissionList(
		publicPermissions,
		[]PermissionName{
			PermissionListNotifications,
			PermissionCountUnreadNotifications,
			PermissionMarkNotificationAsRead,
			PermissionLogout,
			PermissionRefreshToken,
			PermissionUpdateUser,
			PermissionGetMyOwnUser,
			PermissionUpdatePassword,
			PermissionEnableTotp,
			PermissionResetTotp,
			PermissionResetPassword,
			PermissionListWorkspaces,
			PermissionListPublicWorkspaces,
			PermissionListWorkbenchs,
			PermissionListApps,
			PermissionListAppInstances,
			PermissionListMyRequests,
			PermissionAuditUser,
			PermissionGetCurrentTermsOfUseVersion,
			PermissionGetMyTermsOfUseStatus,
			PermissionAcceptTermsOfUse,
			PermissionListOrganizations,
			PermissionGetOrganization,
		},
	)

	workspaceGuestPermissions := permissionList(
		authenticatedPermissions,
		[]PermissionName{
			PermissionListWorkspaces,
			PermissionGetWorkspace,
			PermissionListUsers,
			PermissionCreateRequest,
			PermissionListWorkspaceServiceInstances,
		},
	)

	workspaceMemberPermissions := permissionList(
		workspaceGuestPermissions,
		[]PermissionName{
			PermissionCreateWorkbench,
			PermissionListFilesInWorkspace,
			PermissionCreateRequest,
			PermissionGetWorkspaceServiceInstance,
			PermissionGetWorkspaceServiceInstanceSecret,
		},
	)

	workspaceMaintainerPermissions := permissionList(
		workspaceMemberPermissions,
		[]PermissionName{
			PermissionUpdateWorkspace,
			PermissionUploadFilesToWorkspace,
			PermissionModifyFilesInWorkspace,
			PermissionSearchUsers,
			PermissionCreateRequest,
		},
	)

	workspaceDataManagerPermissions := permissionList(
		workspaceMemberPermissions,
		[]PermissionName{
			PermissionUploadFilesToWorkspace,
			PermissionModifyFilesInWorkspace,
			PermissionDownloadFilesFromWorkspace,
			PermissionManageUsersDataRoleInWorkspace,
			PermissionCreateRequest,
			PermissionListRequests,
		},
	)

	workspaceAdminPermissions := permissionList(
		workspaceMaintainerPermissions,
		[]PermissionName{
			PermissionListAppInstances,
			PermissionListWorkbenchs,
			PermissionUpdateWorkbench,
			PermissionGetWorkbench,
			PermissionStreamWorkbench,
			PermissionDeleteWorkbench,
			PermissionAuditWorkbench,
			PermissionManageUsersInWorkbench,
			PermissionDeleteWorkspace,
			PermissionAuditWorkspace,
			PermissionManageUsersInWorkspace,
			PermissionListRequests,
			PermissionGetRequest,
			PermissionApproveRequest,
			PermissionDeleteRequest,
			PermissionCreateWorkspaceServiceInstance,
			PermissionUpdateWorkspaceServiceInstance,
			PermissionDeleteWorkspaceServiceInstance,
		},
	)

	workbenchViewerPermissions := permissionList(
		authenticatedPermissions,
		[]PermissionName{
			PermissionListAppInstances,
			PermissionListWorkbenchs,
			PermissionGetWorkbench,
			PermissionStreamWorkbench,
			PermissionListUsers,
		},
	)

	workbenchMemberPermissions := permissionList(
		workbenchViewerPermissions,
		[]PermissionName{
			PermissionCreateAppInstance,
			PermissionUpdateAppInstance,
			PermissionGetAppInstance,
			PermissionDeleteAppInstance,
			PermissionUpdateWorkbench,
		},
	)

	workbenchAdminPermissions := permissionList(
		workbenchMemberPermissions,
		[]PermissionName{
			PermissionDeleteWorkbench,
			PermissionManageUsersInWorkbench,
			PermissionSearchUsers,
			PermissionAuditWorkbench,
		},
	)

	platformSettingsManagerPermissions := permissionList(
		authenticatedPermissions,
		[]PermissionName{
			PermissionSetPlatformSettings,
			PermissionListTermsOfUseVersions,
			PermissionGetTermsOfUseVersion,
			PermissionCreateTermsOfUseVersion,
			PermissionUpdateTermsOfUseVersion,
			PermissionPublishTermsOfUseVersion,
		},
	)

	platformUserManagerPermissions := permissionList(
		authenticatedPermissions,
		[]PermissionName{
			PermissionListUsers,
			PermissionCreateUser,
			PermissionUpdateUser,
			PermissionManageUserRoles,
			PermissionManageDynamicRoles,
			PermissionGetUser,
			PermissionDeleteUser,
			PermissionResetPassword,
			PermissionListTermsOfUseAcceptances,
		},
	)

	platformOrganizationManagerPermissions := permissionList(
		authenticatedPermissions,
		[]PermissionName{
			PermissionCreateOrganization,
			PermissionUpdateOrganization,
			PermissionDeleteOrganization,
		},
	)

	platformAuditorPermissions := permissionList(
		authenticatedPermissions,
		[]PermissionName{
			PermissionAuditPlatform,
		},
	)

	platformWorkspaceManagerPermissions := permissionList(
		authenticatedPermissions,
		[]PermissionName{
			PermissionCreateWorkspace,
			PermissionGetWorkspace,
			PermissionUpdateWorkspace,
			PermissionDeleteWorkspace,
		},
	)

	appStoreAdminPermissions := permissionList(
		authenticatedPermissions,
		[]PermissionName{
			PermissionListApps,
			PermissionCreateApp,
			PermissionUpdateApp,
			PermissionGetApp,
			PermissionDeleteApp,
		},
	)

	dataManagerPermissions := permissionList(
		authenticatedPermissions,
		[]PermissionName{
			PermissionListWorkspaces,
			PermissionGetWorkspace,
			PermissionListFilesInWorkspace,
			PermissionUploadFilesToWorkspace,
			PermissionModifyFilesInWorkspace,
			PermissionDownloadFilesFromWorkspace,
		},
	)

	healthcheckerPermissions := permissionList(
		[]PermissionName{
			PermissionGetHealthCheck,
		},
	)

	superAdminPermissions := permissionList(
		authenticatedPermissions,
		platformSettingsManagerPermissions,
		platformUserManagerPermissions,
		platformOrganizationManagerPermissions,
		platformAuditorPermissions,
		platformWorkspaceManagerPermissions,
		appStoreAdminPermissions,
		dataManagerPermissions,
		workspaceAdminPermissions,
		workspaceDataManagerPermissions,
		workbenchAdminPermissions,
		healthcheckerPermissions,
		[]PermissionName{
			PermissionInitializeTenant,
		},
	)

	return AuthorizationSchema{
		Permissions: []PermissionDefinition{
			permissionDefinition(PermissionAuthenticate, "Allow the user to authenticate", nil),
			permissionDefinition(PermissionLogout, "Allow the user to logout", nil),
			permissionDefinition(PermissionRefreshToken, "Allow the user to refresh the jwt token", nil),
			permissionDefinition(PermissionGetListOfPossibleWayToAuthenticate, "Allow the user to get a list of possible ways to authenticate", nil),
			permissionDefinition(PermissionAuthenticateUsingAuth2_0, "Allow the user to authenticate using oauth2", nil),
			permissionDefinition(PermissionAuthenticateRedirectUsingAuth2_0, "Allow the user to be redirected after authenticating using oauth2", nil),

			permissionDefinition(PermissionGetHealthCheck, "Allow the user to get the healthcheck status", nil),

			permissionDefinition(PermissionInitializeTenant, "Allow the user to initialize the tenant", nil),

			permissionDefinition(PermissionListNotifications, "Allow the user to list notifications", contexts(one(RoleContextUser))),
			permissionDefinition(PermissionCountUnreadNotifications, "Allow the user to count unread notifications", contexts(one(RoleContextUser))),
			permissionDefinition(PermissionMarkNotificationAsRead, "Allow the user to mark a notification as read", contexts(one(RoleContextUser))),

			permissionDefinition(PermissionListUsers, "Allow the user to list users", contexts(one(RoleContextUser))),
			permissionDefinition(PermissionSearchUsers, "Allow the user to search users", nil),
			permissionDefinition(PermissionCreateUser, "Allow the user to create a user", nil),
			permissionDefinition(PermissionUpdateUser, "Allow the user to update a user", contexts(one(RoleContextUser))),
			permissionDefinition(PermissionGetMyOwnUser, "Allow the user to get his own user", contexts(one(RoleContextUser))),
			permissionDefinition(PermissionUpdatePassword, "Allow the user to update his password", contexts(one(RoleContextUser))),
			permissionDefinition(PermissionEnableTotp, "Allow the user to enable TOTP", contexts(one(RoleContextUser))),
			permissionDefinition(PermissionResetTotp, "Allow the user to reset TOTP", contexts(one(RoleContextUser))),
			permissionDefinition(PermissionGetUser, "Allow the user to get a user", contexts(one(RoleContextUser))),
			permissionDefinition(PermissionDeleteUser, "Allow the user to delete a user", contexts(one(RoleContextUser))),
			permissionDefinition(PermissionResetPassword, "Allow the user to reset a user's password", contexts(one(RoleContextUser))),
			permissionDefinition(PermissionManageUserRoles, "Allow the user to manage user roles", contexts(one(RoleContextUser))),
			permissionDefinition(PermissionAuditUser, "Allow the user to audit users", contexts(one(RoleContextUser))),

			permissionDefinition(PermissionGetPlatformSettings, "Allow the user to get platform settings", nil),
			permissionDefinition(PermissionSetPlatformSettings, "Allow the user to set platform settings", nil),
			permissionDefinition(PermissionAuditPlatform, "Allow the user to audit the platform", nil),
			permissionDefinition(PermissionManageDynamicRoles, "Allow the user to create dynamic roles", nil),

			permissionDefinition(PermissionListAppInstances, "Allow the user to list app instances", contexts(one(RoleContextWorkbench))),
			permissionDefinition(PermissionCreateAppInstance, "Allow the user to create an app instance", contexts(one(RoleContextWorkbench))),
			permissionDefinition(PermissionUpdateAppInstance, "Allow the user to update an app instance", contexts(one(RoleContextWorkbench))),
			permissionDefinition(PermissionGetAppInstance, "Allow the user to get an app instance", contexts(one(RoleContextWorkbench))),
			permissionDefinition(PermissionDeleteAppInstance, "Allow the user to delete an app instance", contexts(one(RoleContextWorkbench))),

			permissionDefinition(PermissionListWorkbenchs, "Allow the user to list workbenchs", contexts(one(RoleContextWorkbench))),
			permissionDefinition(PermissionCreateWorkbench, "Allow the user to create a workbench", contexts(one(RoleContextWorkspace))),
			permissionDefinition(PermissionUpdateWorkbench, "Allow the user to update a workbench", contexts(one(RoleContextWorkbench))),
			permissionDefinition(PermissionGetWorkbench, "Allow the user to get a workbench", contexts(one(RoleContextWorkbench))),
			permissionDefinition(PermissionStreamWorkbench, "Allow the user to stream a workbench", contexts(one(RoleContextWorkbench))),
			permissionDefinition(PermissionDeleteWorkbench, "Allow the user to delete a workbench", contexts(one(RoleContextWorkbench))),
			permissionDefinition(PermissionAuditWorkbench, "Allow the user to audit a workbench", contexts(one(RoleContextWorkbench))),
			permissionDefinition(PermissionManageUsersInWorkbench, "Allow the user to manage users in a workbench", contexts(one(RoleContextWorkbench))),

			permissionDefinition(PermissionListWorkspaces, "Allow the user to list workspaces", contexts(one(RoleContextWorkspace))),
			permissionDefinition(PermissionListPublicWorkspaces, "Allow the user to list public workspaces", nil),
			permissionDefinition(PermissionCreateWorkspace, "Allow the user to create a workspace", nil),
			permissionDefinition(PermissionUpdateWorkspace, "Allow the user to update a workspace", contexts(one(RoleContextWorkspace))),
			permissionDefinition(PermissionGetWorkspace, "Allow the user to get a workspace", contexts(one(RoleContextWorkspace))),
			permissionDefinition(PermissionDeleteWorkspace, "Allow the user to delete a workspace", contexts(one(RoleContextWorkspace))),
			permissionDefinition(PermissionManageUsersInWorkspace, "Allow the user to manage users in a workspace", contexts(one(RoleContextWorkspace))),
			permissionDefinition(PermissionManageUsersDataRoleInWorkspace, "Allow the user to manage users' data role in a workspace", contexts(one(RoleContextWorkspace))),
			permissionDefinition(PermissionListFilesInWorkspace, "Allow the user to list files in a workspace", contexts(one(RoleContextWorkspace))),
			permissionDefinition(PermissionUploadFilesToWorkspace, "Allow the user to upload files to a workspace", contexts(one(RoleContextWorkspace))),
			permissionDefinition(PermissionDownloadFilesFromWorkspace, "Allow the user to download files from a workspace", contexts(one(RoleContextWorkspace))),
			permissionDefinition(PermissionModifyFilesInWorkspace, "Allow the user to modify files in a workspace", contexts(one(RoleContextWorkspace))),
			permissionDefinition(PermissionAuditWorkspace, "Allow the user to audit a workspace", contexts(one(RoleContextWorkspace))),

			permissionDefinition(PermissionListWorkspaceServiceInstances, "Allow the user to list workspace service instances", contexts(one(RoleContextWorkspace))),
			permissionDefinition(PermissionCreateWorkspaceServiceInstance, "Allow the user to create a workspace service instance", contexts(one(RoleContextWorkspace))),
			permissionDefinition(PermissionUpdateWorkspaceServiceInstance, "Allow the user to update a workspace service instance", contexts(one(RoleContextWorkspace))),
			permissionDefinition(PermissionGetWorkspaceServiceInstance, "Allow the user to get a workspace service instance", contexts(one(RoleContextWorkspace))),
			permissionDefinition(PermissionGetWorkspaceServiceInstanceSecret, "Allow the user to get the secrets of a workspace service instance", contexts(one(RoleContextWorkspace))),
			permissionDefinition(PermissionDeleteWorkspaceServiceInstance, "Allow the user to delete a workspace service instance", contexts(one(RoleContextWorkspace))),

			permissionDefinition(PermissionListApps, "Allow the user to list apps", nil),
			permissionDefinition(PermissionCreateApp, "Allow the user to create an app", nil),
			permissionDefinition(PermissionUpdateApp, "Allow the user to update an app", nil),
			permissionDefinition(PermissionGetApp, "Allow the user to get an app", nil),
			permissionDefinition(PermissionDeleteApp, "Allow the user to delete an app", nil),

			permissionDefinition(PermissionListRequests, "Allow the user to list requests", contexts(one(RoleContextWorkspace))),
			permissionDefinition(PermissionListMyRequests, "Allow the user to list his requests", nil),
			permissionDefinition(PermissionGetRequest, "Allow the user to get a request", contexts(one(RoleContextWorkspace))),
			permissionDefinition(PermissionCreateRequest, "Allow the user to create a request", contexts(one(RoleContextWorkspace))),
			permissionDefinition(PermissionApproveRequest, "Allow the user to approve a request", contexts(one(RoleContextWorkspace))),
			permissionDefinition(PermissionDeleteRequest, "Allow the user to delete a request", contexts(one(RoleContextWorkspace))),

			permissionDefinition(PermissionCreateTermsOfUseVersion, "Allow the user to create a terms of use version", nil),
			permissionDefinition(PermissionUpdateTermsOfUseVersion, "Allow the user to update a terms of use version", nil),
			permissionDefinition(PermissionPublishTermsOfUseVersion, "Allow the user to publish a terms of use version", nil),
			permissionDefinition(PermissionGetTermsOfUseVersion, "Allow the user to get a terms of use version", nil),
			permissionDefinition(PermissionListTermsOfUseVersions, "Allow the user to list terms of use versions", nil),
			permissionDefinition(PermissionGetCurrentTermsOfUseVersion, "Allow the user to get the current terms of use version", nil),
			permissionDefinition(PermissionListTermsOfUseAcceptances, "Allow the user to list terms of use acceptances", contexts(one(RoleContextUser))),
			permissionDefinition(PermissionGetMyTermsOfUseStatus, "Allow the user to get his terms of use acceptance status", contexts(one(RoleContextUser))),
			permissionDefinition(PermissionAcceptTermsOfUse, "Allow the user to accept the terms of use", contexts(one(RoleContextUser))),

			permissionDefinition(PermissionListOrganizations, "Allow the user to list organizations", nil),
			permissionDefinition(PermissionGetOrganization, "Allow the user to get an organization", nil),
			permissionDefinition(PermissionCreateOrganization, "Allow the user to create an organization", nil),
			permissionDefinition(PermissionUpdateOrganization, "Allow the user to update an organization", nil),
			permissionDefinition(PermissionDeleteOrganization, "Allow the user to delete an organization", nil),
		},
		Roles: []*RoleDefinition{
			roleDefinition(
				RolePublic,
				"Public users can authenticate and read public platform settings",
				RoleScopePlatform,
				nil,
				publicPermissions,
			),
			roleDefinition(
				RoleAuthenticated,
				"Authenticated users can manage their own session, profile, notifications, and base resources",
				RoleScopePlatform,
				contexts(one(RoleContextUser)),
				authenticatedPermissions,
			),
			roleDefinition(
				RoleWorkspaceGuest,
				"Workspace guests can view workspace metadata and create requests",
				RoleScopeWorkspace,
				contexts(one(RoleContextWorkspace)),
				workspaceGuestPermissions,
			),
			roleDefinition(
				RoleWorkspaceMember,
				"Workspace members can create workbenches and list workspace files",
				RoleScopeWorkspace,
				contexts(one(RoleContextWorkspace)),
				workspaceMemberPermissions,
			),
			roleDefinition(
				RoleWorkspaceMaintainer,
				"Workspace maintainers can update workspace metadata and manage workspace files",
				RoleScopeWorkspace,
				contexts(one(RoleContextWorkspace)),
				workspaceMaintainerPermissions,
			),
			roleDefinition(
				RoleWorkspaceDataManager,
				"Workspace data managers can manage workspace files and data-manager assignments",
				RoleScopeWorkspace,
				contexts(one(RoleContextWorkspace)),
				workspaceDataManagerPermissions,
			),
			roleDefinition(
				RoleWorkspaceAdmin,
				"Workspace admins can administer workspace users, requests, workbenches, files, and services",
				RoleScopeWorkspace,
				contexts(one(RoleContextWorkspace)),
				workspaceAdminPermissions,
			),
			roleDefinition(
				RoleWorkbenchViewer,
				"Workbench viewers can view and stream workbenches",
				RoleScopeWorkbench,
				contexts(one(RoleContextWorkbench)),
				workbenchViewerPermissions,
			),
			roleDefinition(
				RoleWorkbenchMember,
				"Workbench members can update workbenches and manage app instances",
				RoleScopeWorkbench,
				contexts(one(RoleContextWorkbench)),
				workbenchMemberPermissions,
			),
			roleDefinition(
				RoleWorkbenchAdmin,
				"Workbench admins can administer workbenches and their users",
				RoleScopeWorkbench,
				contexts(one(RoleContextWorkbench)),
				workbenchAdminPermissions,
			),
			roleDefinition(
				RoleHealthchecker,
				"Healthcheckers can read healthcheck status",
				RoleScopePlatform,
				contexts(wildcard(RoleContextUser)),
				healthcheckerPermissions,
			),
			roleDefinition(
				RolePlatformSettingsManager,
				"Platform settings managers can manage platform settings",
				RoleScopePlatform,
				contexts(wildcard(RoleContextUser)),
				platformSettingsManagerPermissions,
			),
			roleDefinition(
				RolePlatformUserManager,
				"Platform user managers can administer platform users and their roles",
				RoleScopePlatform,
				contexts(wildcard(RoleContextUser)),
				platformUserManagerPermissions,
			),
			roleDefinition(
				RolePlatformOrganizationManager,
				"Platform organization managers can manage organizations",
				RoleScopePlatform,
				contexts(wildcard(RoleContextUser)),
				platformOrganizationManagerPermissions,
			),
			roleDefinition(
				RolePlatformAuditor,
				"Platform auditors can audit the platform",
				RoleScopePlatform,
				contexts(wildcard(RoleContextUser)),
				platformAuditorPermissions,
			),
			roleDefinition(
				RolePlatformWorkspaceManager,
				"Platform workspace managers can manage any workspace",
				anyContext(RoleContextWorkspace),
				platformWorkspaceManagerPermissions,
			),
			roleDefinition(
				RoleAppStoreAdmin,
				"App store admins can administer apps",
				RoleScopePlatform,
				contexts(wildcard(RoleContextUser)),
				appStoreAdminPermissions,
			),
			roleDefinition(
				RolePlatformDataManager,
				"Data managers can manage workspace data across workspaces",
				RoleScopePlatform,
				contexts(wildcard(RoleContextWorkspace)),
				dataManagerPermissions,
			),
			roleDefinition(
				RoleSuperAdmin,
				"Super admins can perform all platform, workspace, and workbench actions",
				RoleScopeSystem,
				contexts(wildcard(RoleContextUser), wildcard(RoleContextWorkspace), wildcard(RoleContextWorkbench)),
				superAdminPermissions,
			),
		},
	}
}

func permissionDefinition(name PermissionName, description string, context map[ContextDimension]ContextQuantifier) PermissionDefinition {
	return PermissionDefinition{
		Name:                      name,
		Description:               description,
		RequiredContextDimensions: contextDimensions(context),
	}
}

func roleDefinition(name RoleName, description string, scope RoleScope, context map[ContextDimension]ContextQuantifier, permissions []PermissionName) *RoleDefinition {
	return &RoleDefinition{
		Name:                      name,
		Description:               description,
		Scope:                     scope,
		RequiredContextDimensions: context,
		Permissions:               append([]PermissionName(nil), permissions...),
	}
}

func permissionList(permissionGroups ...[]PermissionName) []PermissionName {
	result := make([]PermissionName, 0)
	seen := make(map[PermissionName]bool)
	for _, group := range permissionGroups {
		for _, permission := range group {
			if seen[permission] {
				continue
			}
			seen[permission] = true
			result = append(result, permission)
		}
	}
	return result
}

func contextDimensions(context map[ContextDimension]ContextQuantifier) []ContextDimension {
	if len(context) == 0 {
		return nil
	}

	result := make([]ContextDimension, 0, len(context))
	for dimension := range context {
		result = append(result, dimension)
	}
	return result
}

// one declares a dimension the role binds to a concrete value
func one(dim ContextDimension) map[ContextDimension]ContextQuantifier {
	return map[ContextDimension]ContextQuantifier{dim: ContextQuantifierOne}
}

// wildcard declares a dimension the role binds to "*" (any value)
func wildcard(dim ContextDimension) map[ContextDimension]ContextQuantifier {
	return map[ContextDimension]ContextQuantifier{dim: ContextQuantifierAny}
}

// contexts merges one()/wildcard() declarations into a single required-context map
func contexts(dims ...map[ContextDimension]ContextQuantifier) map[ContextDimension]ContextQuantifier {
	if len(dims) == 0 {
		return nil
	}
	merged := make(map[ContextDimension]ContextQuantifier, len(dims))
	for _, dim := range dims {
		for k, v := range dim {
			merged[k] = v
		}
	}
	return merged
}
