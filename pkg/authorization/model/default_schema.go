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
			PermissionCreateWorkspace,
			PermissionListWorkspaces,
			PermissionListWorkbenchs,
			PermissionListApps,
			PermissionListAppInstances,
			PermissionListMyRequests,
			PermissionAuditUser,
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
			PermissionGetWorkspaceServiceInstance,
		},
	)

	workspaceMemberPermissions := permissionList(
		workspaceGuestPermissions,
		[]PermissionName{
			PermissionCreateWorkbench,
			PermissionListFilesInWorkspace,
			PermissionCreateRequest,
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
		},
	)

	platformUserManagerPermissions := permissionList(
		authenticatedPermissions,
		[]PermissionName{
			PermissionListUsers,
			PermissionCreateUser,
			PermissionUpdateUser,
			PermissionManageUserRoles,
			PermissionGetUser,
			PermissionDeleteUser,
			PermissionResetPassword,
		},
	)

	platformAuditorPermissions := permissionList(
		authenticatedPermissions,
		[]PermissionName{
			PermissionAuditPlatform,
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
		platformAuditorPermissions,
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

			permissionDefinition(PermissionListNotifications, "Allow the user to list notifications", oneContext(RoleContextUser)),
			permissionDefinition(PermissionCountUnreadNotifications, "Allow the user to count unread notifications", oneContext(RoleContextUser)),
			permissionDefinition(PermissionMarkNotificationAsRead, "Allow the user to mark a notification as read", oneContext(RoleContextUser)),

			permissionDefinition(PermissionListUsers, "Allow the user to list users", oneContext(RoleContextUser)),
			permissionDefinition(PermissionSearchUsers, "Allow the user to search users", nil),
			permissionDefinition(PermissionCreateUser, "Allow the user to create a user", nil),
			permissionDefinition(PermissionUpdateUser, "Allow the user to update a user", oneContext(RoleContextUser)),
			permissionDefinition(PermissionGetMyOwnUser, "Allow the user to get his own user", oneContext(RoleContextUser)),
			permissionDefinition(PermissionUpdatePassword, "Allow the user to update his password", oneContext(RoleContextUser)),
			permissionDefinition(PermissionEnableTotp, "Allow the user to enable TOTP", oneContext(RoleContextUser)),
			permissionDefinition(PermissionResetTotp, "Allow the user to reset TOTP", oneContext(RoleContextUser)),
			permissionDefinition(PermissionGetUser, "Allow the user to get a user", oneContext(RoleContextUser)),
			permissionDefinition(PermissionDeleteUser, "Allow the user to delete a user", oneContext(RoleContextUser)),
			permissionDefinition(PermissionResetPassword, "Allow the user to reset a user's password", oneContext(RoleContextUser)),
			permissionDefinition(PermissionManageUserRoles, "Allow the user to manage user roles", oneContext(RoleContextUser)),
			permissionDefinition(PermissionAuditUser, "Allow the user to audit users", oneContext(RoleContextUser)),

			permissionDefinition(PermissionGetPlatformSettings, "Allow the user to get platform settings", nil),
			permissionDefinition(PermissionSetPlatformSettings, "Allow the user to set platform settings", nil),
			permissionDefinition(PermissionAuditPlatform, "Allow the user to audit the platform", nil),

			permissionDefinition(PermissionListAppInstances, "Allow the user to list app instances", oneContext(RoleContextWorkbench)),
			permissionDefinition(PermissionCreateAppInstance, "Allow the user to create an app instance", oneContext(RoleContextWorkbench)),
			permissionDefinition(PermissionUpdateAppInstance, "Allow the user to update an app instance", oneContext(RoleContextWorkbench)),
			permissionDefinition(PermissionGetAppInstance, "Allow the user to get an app instance", oneContext(RoleContextWorkbench)),
			permissionDefinition(PermissionDeleteAppInstance, "Allow the user to delete an app instance", oneContext(RoleContextWorkbench)),

			permissionDefinition(PermissionListWorkbenchs, "Allow the user to list workbenchs", oneContext(RoleContextWorkbench)),
			permissionDefinition(PermissionCreateWorkbench, "Allow the user to create a workbench", oneContext(RoleContextWorkspace)),
			permissionDefinition(PermissionUpdateWorkbench, "Allow the user to update a workbench", oneContext(RoleContextWorkbench)),
			permissionDefinition(PermissionGetWorkbench, "Allow the user to get a workbench", oneContext(RoleContextWorkbench)),
			permissionDefinition(PermissionStreamWorkbench, "Allow the user to stream a workbench", oneContext(RoleContextWorkbench)),
			permissionDefinition(PermissionDeleteWorkbench, "Allow the user to delete a workbench", oneContext(RoleContextWorkbench)),
			permissionDefinition(PermissionAuditWorkbench, "Allow the user to audit a workbench", oneContext(RoleContextWorkbench)),
			permissionDefinition(PermissionManageUsersInWorkbench, "Allow the user to manage users in a workbench", oneContext(RoleContextWorkbench)),

			permissionDefinition(PermissionListWorkspaces, "Allow the user to list workspaces", oneContext(RoleContextWorkspace)),
			permissionDefinition(PermissionCreateWorkspace, "Allow the user to create a workspace", nil),
			permissionDefinition(PermissionUpdateWorkspace, "Allow the user to update a workspace", oneContext(RoleContextWorkspace)),
			permissionDefinition(PermissionGetWorkspace, "Allow the user to get a workspace", oneContext(RoleContextWorkspace)),
			permissionDefinition(PermissionDeleteWorkspace, "Allow the user to delete a workspace", oneContext(RoleContextWorkspace)),
			permissionDefinition(PermissionManageUsersInWorkspace, "Allow the user to manage users in a workspace", oneContext(RoleContextWorkspace)),
			permissionDefinition(PermissionManageUsersDataRoleInWorkspace, "Allow the user to manage users' data role in a workspace", oneContext(RoleContextWorkspace)),
			permissionDefinition(PermissionListFilesInWorkspace, "Allow the user to list files in a workspace", oneContext(RoleContextWorkspace)),
			permissionDefinition(PermissionUploadFilesToWorkspace, "Allow the user to upload files to a workspace", oneContext(RoleContextWorkspace)),
			permissionDefinition(PermissionDownloadFilesFromWorkspace, "Allow the user to download files from a workspace", oneContext(RoleContextWorkspace)),
			permissionDefinition(PermissionModifyFilesInWorkspace, "Allow the user to modify files in a workspace", oneContext(RoleContextWorkspace)),
			permissionDefinition(PermissionAuditWorkspace, "Allow the user to audit a workspace", oneContext(RoleContextWorkspace)),

			permissionDefinition(PermissionListWorkspaceServiceInstances, "Allow the user to list workspace service instances", oneContext(RoleContextWorkspace)),
			permissionDefinition(PermissionCreateWorkspaceServiceInstance, "Allow the user to create a workspace service instance", oneContext(RoleContextWorkspace)),
			permissionDefinition(PermissionUpdateWorkspaceServiceInstance, "Allow the user to update a workspace service instance", oneContext(RoleContextWorkspace)),
			permissionDefinition(PermissionGetWorkspaceServiceInstance, "Allow the user to get a workspace service instance", oneContext(RoleContextWorkspace)),
			permissionDefinition(PermissionDeleteWorkspaceServiceInstance, "Allow the user to delete a workspace service instance", oneContext(RoleContextWorkspace)),

			permissionDefinition(PermissionListApps, "Allow the user to list apps", nil),
			permissionDefinition(PermissionCreateApp, "Allow the user to create an app", nil),
			permissionDefinition(PermissionUpdateApp, "Allow the user to update an app", nil),
			permissionDefinition(PermissionGetApp, "Allow the user to get an app", nil),
			permissionDefinition(PermissionDeleteApp, "Allow the user to delete an app", nil),

			permissionDefinition(PermissionListRequests, "Allow the user to list requests", oneContext(RoleContextWorkspace)),
			permissionDefinition(PermissionListMyRequests, "Allow the user to list his requests", nil),
			permissionDefinition(PermissionGetRequest, "Allow the user to get a request", oneContext(RoleContextWorkspace)),
			permissionDefinition(PermissionCreateRequest, "Allow the user to create a request", oneContext(RoleContextWorkspace)),
			permissionDefinition(PermissionApproveRequest, "Allow the user to approve a request", oneContext(RoleContextWorkspace)),
			permissionDefinition(PermissionDeleteRequest, "Allow the user to delete a request", oneContext(RoleContextWorkspace)),
		},
		Roles: []*RoleDefinition{
			roleDefinition(
				RolePublic,
				"Public users can authenticate and read public platform settings",
				nil,
				publicPermissions,
			),
			roleDefinition(
				RoleAuthenticated,
				"Authenticated users can manage their own session, profile, notifications, and base resources",
				oneContext(RoleContextUser),
				authenticatedPermissions,
			),
			roleDefinition(
				RoleWorkspaceGuest,
				"Workspace guests can view workspace metadata and create requests",
				oneContext(RoleContextWorkspace),
				workspaceGuestPermissions,
			),
			roleDefinition(
				RoleWorkspaceMember,
				"Workspace members can create workbenches and list workspace files",
				oneContext(RoleContextWorkspace),
				workspaceMemberPermissions,
			),
			roleDefinition(
				RoleWorkspaceMaintainer,
				"Workspace maintainers can update workspace metadata and manage workspace files",
				oneContext(RoleContextWorkspace),
				workspaceMaintainerPermissions,
			),
			roleDefinition(
				RoleWorkspaceDataManager,
				"Workspace data managers can manage workspace files and data-manager assignments",
				oneContext(RoleContextWorkspace),
				workspaceDataManagerPermissions,
			),
			roleDefinition(
				RoleWorkspaceAdmin,
				"Workspace admins can administer workspace users, requests, workbenches, files, and services",
				oneContext(RoleContextWorkspace),
				workspaceAdminPermissions,
			),
			roleDefinition(
				RoleWorkbenchViewer,
				"Workbench viewers can view and stream workbenches",
				oneContext(RoleContextWorkbench),
				workbenchViewerPermissions,
			),
			roleDefinition(
				RoleWorkbenchMember,
				"Workbench members can update workbenches and manage app instances",
				oneContext(RoleContextWorkbench),
				workbenchMemberPermissions,
			),
			roleDefinition(
				RoleWorkbenchAdmin,
				"Workbench admins can administer workbenches and their users",
				oneContext(RoleContextWorkbench),
				workbenchAdminPermissions,
			),
			roleDefinition(
				RoleHealthchecker,
				"Healthcheckers can read healthcheck status",
				anyContext(RoleContextUser),
				healthcheckerPermissions,
			),
			roleDefinition(
				RolePlatformSettingsManager,
				"Platform settings managers can manage platform settings",
				anyContext(RoleContextUser),
				platformSettingsManagerPermissions,
			),
			roleDefinition(
				RolePlateformUserManager,
				"Platform user managers can administer platform users and their roles",
				anyContext(RoleContextUser),
				platformUserManagerPermissions,
			),
			roleDefinition(
				RolePlatformAuditor,
				"Platform auditors can audit the platform",
				anyContext(RoleContextUser),
				platformAuditorPermissions,
			),
			roleDefinition(
				RoleAppStoreAdmin,
				"App store admins can administer apps",
				anyContext(RoleContextUser),
				appStoreAdminPermissions,
			),
			roleDefinition(
				RoleDataManager,
				"Data managers can manage workspace data across workspaces",
				anyContext(RoleContextWorkspace),
				dataManagerPermissions,
			),
			roleDefinition(
				RoleSuperAdmin,
				"Super admins can perform all platform, workspace, and workbench actions",
				map[ContextDimension]ContextQuantifier{
					RoleContextUser:      ContextQuantifierAny,
					RoleContextWorkspace: ContextQuantifierAny,
					RoleContextWorkbench: ContextQuantifierAny,
				},
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

func roleDefinition(name RoleName, description string, context map[ContextDimension]ContextQuantifier, permissions []PermissionName) *RoleDefinition {
	return &RoleDefinition{
		Name:                      name,
		Description:               description,
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

func oneContext(context ContextDimension) map[ContextDimension]ContextQuantifier {
	return map[ContextDimension]ContextQuantifier{context: ContextQuantifierOne}
}

func anyContext(context ContextDimension) map[ContextDimension]ContextQuantifier {
	return map[ContextDimension]ContextQuantifier{context: ContextQuantifierAny}
}
