# Audit Log

## Overview

CHORUS records an audit trail of user actions across the platform. Every audit entry captures:

- **Who** performed the action (actor)
- **What** action was performed
- **Where** it happened (workspace, workbench, target user)
- **When** it occurred
- **Whether it succeeded or failed**, including error details on failure

## Audit Entry Structure

| Field | Description |
|-------|-------------|
| **Actor** | The authenticated user who performed the action (ID + username) |
| **Action** | The type of operation (e.g. `CreateWorkspace`, `DeleteWorkbench`) |
| **Workspace ID** | The workspace where the action occurred (when applicable) |
| **Workbench ID** | The workbench where the action occurred (when applicable) |
| **User ID** | The user being acted upon (when applicable, e.g. role assignment) |
| **Description** | Human-readable summary of what happened |
| **Details** | Additional context such as entity names, error messages, gRPC status codes |
| **Correlation ID** | Links related actions within the same request |
| **Created At** | When the action occurred (UTC) |

Context fields (workspace, workbench, user) are only present when relevant to the action. For example, a workspace creation includes the workspace ID, but a login does not.

---

## What Is Audited

### Authentication

| Operation | Action | Tracks Success | Tracks Failure | Context |
|-----------|--------|:-:|:-:|---------|
| Login | `LoginUser` | Yes | No | - |
| OAuth login | `LoginUser` | No | Yes | - |
| OAuth redirect (successful login) | `LoginUser` | Yes | No | - |
| Logout | `LogoutUser` | Yes | Yes | - |

### Users

| Operation | Action | Tracks Success | Tracks Failure | Context |
|-----------|--------|:-:|:-:|---------|
| Create user | `CreateUser` | Yes | Yes | target user |
| Update user | `UpdateUser` | Yes | Yes | target user |
| Delete user | `DeleteUser` | Yes | Yes | target user |
| Assign role to user | `AssignUserRole` | Yes | Yes | target user |
| Revoke role from user | `RevokeUserRole` | Yes | Yes | target user |
| Change password | `ChangeUserPassword` | Yes | Yes | - |
| Reset password (admin) | `ResetUserPassword` | Yes | Yes | target user |
| Enable TOTP | `EnableUserTotp` | Yes | Yes | - |
| Reset TOTP | `ResetUserTotp` | Yes | Yes | - |

### Workspaces

| Operation | Action | Tracks Success | Tracks Failure | Context |
|-----------|--------|:-:|:-:|---------|
| Create workspace | `CreateWorkspace` | Yes | Yes | workspace |
| Update workspace | `UpdateWorkspace` | Yes | Yes | workspace |
| Delete workspace | `DeleteWorkspace` | Yes | Yes | workspace |
| Add member to workspace | `AddWorkspaceMember` | Yes | Yes | workspace, target user |
| Remove member from workspace | `RemoveWorkspaceMember` | Yes | Yes | workspace, target user |

### Workbenches

| Operation | Action | Tracks Success | Tracks Failure | Context |
|-----------|--------|:-:|:-:|---------|
| Create workbench | `CreateWorkbench` | Yes | Yes | workspace, workbench |
| Update workbench | `UpdateWorkbench` | Yes | Yes | workspace, workbench |
| Delete workbench | `DeleteWorkbench` | Yes | Yes | workspace, workbench |
| Add member to workbench | `AddWorkbenchMember` | Yes | Yes | workbench, target user (workspace on success) |
| Remove member from workbench | `RemoveWorkbenchMember` | Yes | Yes | workbench, target user (workspace on success) |

### Apps

| Operation | Action | Tracks Success | Tracks Failure | Context |
|-----------|--------|:-:|:-:|---------|
| Create app | `CreateApp` | Yes | Yes | - |
| Update app | `UpdateApp` | Yes | Yes | - |
| Delete app | `DeleteApp` | Yes | Yes | - |
| Bulk create apps | `BulkCreateApp` | Yes | Yes | - |

### App Instances

| Operation | Action | Tracks Success | Tracks Failure | Context |
|-----------|--------|:-:|:-:|---------|
| Create app instance | `CreateAppInstance` | Yes | Yes | workspace, workbench |
| Update app instance | `UpdateAppInstance` | Yes | Yes | workspace, workbench |
| Delete app instance | `DeleteAppInstance` | Yes | Yes | workspace, workbench |

### Files

| Operation | Action | Tracks Success | Tracks Failure | Context |
|-----------|--------|:-:|:-:|---------|
| Get file | `ReadFile` | Yes | Yes | workspace |
| List files | `ListFile` | Yes | Yes | workspace |
| Create file | `CreateFile` | Yes | Yes | workspace |
| Update file | `UpdateFile` | Yes | Yes | workspace |
| Delete file | `DeleteFile` | Yes | Yes | workspace |
| Initiate multipart upload | `InitiateFileUpload` | Yes | Yes | workspace |
| Complete multipart upload | `CompleteFileUpload` | Yes | Yes | workspace |
| Abort multipart upload | `AbortFileUpload` | Yes | Yes | workspace |

### Approval Requests

| Operation | Action | Tracks Success | Tracks Failure | Context |
|-----------|--------|:-:|:-:|---------|
| Create data extraction request | `CreateDataExtractionRequest` | Yes | Yes | workspace |
| Create data transfer request | `CreateDataTransferRequest` | Yes | Yes | source + destination workspace |
| Approve request | `ApproveApprovalRequest` | Yes | Yes | - |
| Delete request | `DeleteApprovalRequest` | Yes | Yes | - |
| Download request file | `DownloadApprovalRequestFile` | Yes | Yes | - |

### Platform Administration

| Operation | Action | Tracks Success | Tracks Failure | Context |
|-----------|--------|:-:|:-:|---------|
| Initialize tenant | `InitializeTenant` | Yes | Yes | - |

### Audit Log Access

| Operation | Action | Tracks Success | Tracks Failure | Context |
|-----------|--------|:-:|:-:|---------|
| List platform audit | `ListPlatformAudit` | No | Yes | - |
| List workspace audit | `ListWorkspaceAudit` | No | Yes | workspace |
| List workbench audit | `ListWorkbenchAudit` | No | Yes | workbench |
| List user audit | `ListUserAudit` | No | Yes | target user |
| List actor audit | `ListActorAudit` | No | Yes | - |

---

## What Is Not Audited

The following operations are intentionally excluded from the audit log:

| Operation | Reason |
|-----------|--------|
| **Successful reads and listings** (get user, list workspaces, etc.) | These are high-frequency, non-destructive operations. Recording every read would significantly increase log volume without proportional security value. Failed reads are still audited as they may indicate unauthorized access attempts. |
| **File upload parts** | Individual chunk uploads during multipart file uploads are not audited. Only the initiate, complete, and abort operations are recorded. A single file upload can involve hundreds of parts. |
| **Token refresh** | Automatic token renewal is a background operation, not a user-initiated action. |
| **Authentication mode discovery** | Querying available login methods is informational and carries no security significance. |
| **Successful OAuth initiation** | Only the redirect callback (actual login) and failures are recorded. |

---

## Querying Audit Logs

Audit logs can be queried through five endpoints, each scoped to a different view:

| Endpoint | Description | Filters |
|----------|-------------|---------|
| `GET /api/rest/v1/audit` | Platform-wide audit log | actor, action, time range |
| `GET /api/rest/v1/audit/workspaces/{id}` | All actions within a workspace | action, time range |
| `GET /api/rest/v1/audit/workbenches/{id}` | All actions within a workbench | action, time range |
| `GET /api/rest/v1/audit/users/{id}` | All actions involving a user (as actor or target) | action, time range |
| `GET /api/rest/v1/audit/actors/{id}` | All actions performed by a specific user | action, time range |

All endpoints support pagination and can filter by action type and time range.
