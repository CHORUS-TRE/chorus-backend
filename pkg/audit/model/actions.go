package model

type AuditAction string

const (
	// Authentication
	AuditActionUserLogin         AuditAction = "UserLogin"
	AuditActionUserLoginFailed   AuditAction = "UserLoginFailed"
	AuditActionUserLogout        AuditAction = "UserLogout"
	AuditActionUserTokenRefresh  AuditAction = "UserTokenRefresh"
	AuditActionUserOAuthInitiate AuditAction = "UserOAuthInitiate"
	AuditActionUserOAuthCallback AuditAction = "UserOAuthCallback"

	// User CRUD
	AuditActionUserCreate         AuditAction = "UserCreate"
	AuditActionUserRead           AuditAction = "UserRead"
	AuditActionUserUpdate         AuditAction = "UserUpdate"
	AuditActionUserDelete         AuditAction = "UserDelete"
	AuditActionUserList           AuditAction = "UserList"
	AuditActionUserRoleAssign     AuditAction = "UserRoleAssign"
	AuditActionUserRoleRevoke     AuditAction = "UserRoleRevoke"
	AuditActionUserPasswordChange AuditAction = "UserPasswordChange"
	AuditActionUserPasswordReset  AuditAction = "UserPasswordReset"
	AuditActionUserTotpEnable     AuditAction = "UserTotpEnable"
	AuditActionUserTotpReset      AuditAction = "UserTotpReset"

	// Workspace
	AuditActionWorkspaceCreate       AuditAction = "WorkspaceCreate"
	AuditActionWorkspaceRead         AuditAction = "WorkspaceRead"
	AuditActionWorkspaceUpdate       AuditAction = "WorkspaceUpdate"
	AuditActionWorkspaceDelete       AuditAction = "WorkspaceDelete"
	AuditActionWorkspaceList         AuditAction = "WorkspaceList"
	AuditActionWorkspaceMemberAdd    AuditAction = "WorkspaceMemberAdd"
	AuditActionWorkspaceMemberUpdate AuditAction = "WorkspaceMemberUpdate"
	AuditActionWorkspaceMemberRemove AuditAction = "WorkspaceMemberRemove"

	// Workbench
	AuditActionWorkbenchCreate       AuditAction = "WorkbenchCreate"
	AuditActionWorkbenchRead         AuditAction = "WorkbenchRead"
	AuditActionWorkbenchUpdate       AuditAction = "WorkbenchUpdate"
	AuditActionWorkbenchDelete       AuditAction = "WorkbenchDelete"
	AuditActionWorkbenchList         AuditAction = "WorkbenchList"
	AuditActionWorkbenchMemberAdd    AuditAction = "WorkbenchMemberAdd"
	AuditActionWorkbenchMemberRemove AuditAction = "WorkbenchMemberRemove"

	// App
	AuditActionAppCreate     AuditAction = "AppCreate"
	AuditActionAppRead       AuditAction = "AppRead"
	AuditActionAppUpdate     AuditAction = "AppUpdate"
	AuditActionAppDelete     AuditAction = "AppDelete"
	AuditActionAppList       AuditAction = "AppList"
	AuditActionAppBulkCreate AuditAction = "AppBulkCreate"

	// App Instance
	AuditActionAppInstanceCreate AuditAction = "AppInstanceCreate"
	AuditActionAppInstanceRead   AuditAction = "AppInstanceRead"
	AuditActionAppInstanceUpdate AuditAction = "AppInstanceUpdate"
	AuditActionAppInstanceDelete AuditAction = "AppInstanceDelete"
	AuditActionAppInstanceList   AuditAction = "AppInstanceList"

	// File
	AuditActionFileCreate         AuditAction = "FileCreate"
	AuditActionFileRead           AuditAction = "FileRead"
	AuditActionFileUpdate         AuditAction = "FileUpdate"
	AuditActionFileDelete         AuditAction = "FileDelete"
	AuditActionFileList           AuditAction = "FileList"
	AuditActionFileUploadInitiate AuditAction = "FileUploadInitiate"
	AuditActionFileUploadComplete AuditAction = "FileUploadComplete"
	AuditActionFileUploadAbort    AuditAction = "FileUploadAbort"

	// Admin
	AuditActionTenantInitialize AuditAction = "TenantInitialize"
)
