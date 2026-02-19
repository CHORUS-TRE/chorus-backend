package model

type AuditAction string

const (
	// Authentication
	AuditActionUserLogin       AuditAction = "LoginUser"
	AuditActionUserLoginFailed AuditAction = "LoginUserFailed"
	AuditActionUserLogout      AuditAction = "LogoutUser"

	// User CRUD
	AuditActionUserCreate         AuditAction = "CreateUser"
	AuditActionUserRead           AuditAction = "ReadUser"
	AuditActionUserUpdate         AuditAction = "UpdateUser"
	AuditActionUserDelete         AuditAction = "DeleteUser"
	AuditActionUserList           AuditAction = "ListUser"
	AuditActionUserRoleAssign     AuditAction = "AssignUserRole"
	AuditActionUserRoleRevoke     AuditAction = "RevokeUserRole"
	AuditActionUserPasswordChange AuditAction = "ChangeUserPassword"
	AuditActionUserPasswordReset  AuditAction = "ResetUserPassword"
	AuditActionUserTotpEnable     AuditAction = "EnableUserTotp"
	AuditActionUserTotpReset      AuditAction = "ResetUserTotp"

	// Workspace
	AuditActionWorkspaceCreate       AuditAction = "CreateWorkspace"
	AuditActionWorkspaceRead         AuditAction = "ReadWorkspace"
	AuditActionWorkspaceUpdate       AuditAction = "UpdateWorkspace"
	AuditActionWorkspaceDelete       AuditAction = "DeleteWorkspace"
	AuditActionWorkspaceList         AuditAction = "ListWorkspace"
	AuditActionWorkspaceMemberAdd    AuditAction = "AddWorkspaceMember"
	AuditActionWorkspaceMemberUpdate AuditAction = "UpdateWorkspaceMember"
	AuditActionWorkspaceMemberRemove AuditAction = "RemoveWorkspaceMember"

	// Workbench
	AuditActionWorkbenchCreate       AuditAction = "CreateWorkbench"
	AuditActionWorkbenchRead         AuditAction = "ReadWorkbench"
	AuditActionWorkbenchUpdate       AuditAction = "UpdateWorkbench"
	AuditActionWorkbenchDelete       AuditAction = "DeleteWorkbench"
	AuditActionWorkbenchList         AuditAction = "ListWorkbench"
	AuditActionWorkbenchMemberAdd    AuditAction = "AddWorkbenchMember"
	AuditActionWorkbenchMemberRemove AuditAction = "RemoveWorkbenchMember"

	// App
	AuditActionAppCreate     AuditAction = "CreateApp"
	AuditActionAppRead       AuditAction = "ReadApp"
	AuditActionAppUpdate     AuditAction = "UpdateApp"
	AuditActionAppDelete     AuditAction = "DeleteApp"
	AuditActionAppList       AuditAction = "ListApp"
	AuditActionAppBulkCreate AuditAction = "BulkCreateApp"

	// App Instance
	AuditActionAppInstanceCreate AuditAction = "CreateAppInstance"
	AuditActionAppInstanceRead   AuditAction = "ReadAppInstance"
	AuditActionAppInstanceUpdate AuditAction = "UpdateAppInstance"
	AuditActionAppInstanceDelete AuditAction = "DeleteAppInstance"
	AuditActionAppInstanceList   AuditAction = "ListAppInstance"

	// File
	AuditActionFileCreate         AuditAction = "CreateFile"
	AuditActionFileRead           AuditAction = "ReadFile"
	AuditActionFileUpdate         AuditAction = "UpdateFile"
	AuditActionFileDelete         AuditAction = "DeleteFile"
	AuditActionFileList           AuditAction = "ListFile"
	AuditActionFileUploadInitiate AuditAction = "InitiateFileUpload"
	AuditActionFileUploadComplete AuditAction = "CompleteFileUpload"
	AuditActionFileUploadAbort    AuditAction = "AbortFileUpload"

	// Approval Request
	AuditActionApprovalRequestCreate       AuditAction = "CreateApprovalRequest"
	AuditActionApprovalRequestRead         AuditAction = "ReadApprovalRequest"
	AuditActionApprovalRequestList         AuditAction = "ListApprovalRequest"
	AuditActionApprovalRequestApprove      AuditAction = "ApproveApprovalRequest"
	AuditActionApprovalRequestDelete       AuditAction = "DeleteApprovalRequest"
	AuditActionDataExtractionRequestCreate AuditAction = "CreateDataExtractionRequest"
	AuditActionDataTransferRequestCreate   AuditAction = "CreateDataTransferRequest"
	AuditActionApprovalRequestFileDownload AuditAction = "DownloadApprovalRequestFile"

	// Admin
	AuditActionTenantInitialize AuditAction = "InitializeTenant"

	// Audit
	AuditActionPlatformAuditList  AuditAction = "ListPlatformAudit"
	AuditActionWorkspaceAuditList AuditAction = "ListWorkspaceAudit"
	AuditActionWorkbenchAuditList AuditAction = "ListWorkbenchAudit"
	AuditActionUserAuditList      AuditAction = "ListUserAudit"
)
