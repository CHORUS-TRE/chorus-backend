package middleware

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/audit"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/grpc"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/service"
)

var _ chorus.WorkspaceServiceServer = (*workspaceControllerAudit)(nil)

type workspaceControllerAudit struct {
	next        chorus.WorkspaceServiceServer
	auditWriter service.AuditWriter
}

func NewWorkspaceAuditMiddleware(auditWriter service.AuditWriter) func(chorus.WorkspaceServiceServer) chorus.WorkspaceServiceServer {
	return func(next chorus.WorkspaceServiceServer) chorus.WorkspaceServiceServer {
		return &workspaceControllerAudit{
			next:        next,
			auditWriter: auditWriter,
		}
	}
}

func (c workspaceControllerAudit) ListWorkspaces(ctx context.Context, req *chorus.ListWorkspacesRequest) (*chorus.ListWorkspacesReply, error) {
	res, err := c.next.ListWorkspaces(ctx, req)
	if err != nil {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionWorkspaceList,
			audit.WithDescription("Failed to list workspaces."),
			audit.WithErrorMessage(err.Error()),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("filter", req.Filter),
		)
	}
	//  else {
	// 	audit.Record(ctx, c.auditWriter,
	// 		model.AuditActionWorkspaceList,
	// 		audit.WithDescription("Listed workspaces."),
	// 		audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
	// 		audit.WithDetail("filter", req.Filter),
	// 		audit.WithDetail("result_count", len(res.Result.Workspaces)),
	// 	)
	// }

	return res, err
}

func (c workspaceControllerAudit) CreateWorkspace(ctx context.Context, req *chorus.Workspace) (*chorus.CreateWorkspaceReply, error) {
	res, err := c.next.CreateWorkspace(ctx, req)
	if err != nil {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionWorkspaceCreate,
			audit.WithDescription("Failed to create workspace."),
			audit.WithErrorMessage(err.Error()),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("workspace_name", req.Name),
		)
	} else {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionWorkspaceCreate,
			audit.WithDescription(fmt.Sprintf("Created workspace with ID %d.", res.Result.Workspace.Id)),
			audit.WithWorkspaceID(res.Result.Workspace.Id),
			audit.WithDetail("workspace_id", res.Result.Workspace.Id),
			audit.WithDetail("workspace_name", req.Name),
		)
	}

	return res, err
}

func (c workspaceControllerAudit) GetWorkspace(ctx context.Context, req *chorus.GetWorkspaceRequest) (*chorus.GetWorkspaceReply, error) {
	res, err := c.next.GetWorkspace(ctx, req)
	if err != nil {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionWorkspaceRead,
			audit.WithDescription("Failed to get workspace."),
			audit.WithWorkspaceID(req.Id),
			audit.WithErrorMessage(err.Error()),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("workspace_id", req.Id),
		)
	}
	//  else {
	// 	audit.Record(ctx, c.auditWriter,
	// 		model.AuditActionWorkspaceRead,
	// 		audit.WithDescription(fmt.Sprintf("Retrieved workspace with ID %d.", req.Id)),
	// 		audit.WithWorkspaceID(req.Id),
	// 		audit.WithDetail("workspace_id", req.Id),
	// 		audit.WithDetail("workspace_name", res.Result.Workspace.Name),
	// 	)
	// }

	return res, err
}

func (c workspaceControllerAudit) UpdateWorkspace(ctx context.Context, req *chorus.Workspace) (*chorus.UpdateWorkspaceReply, error) {
	res, err := c.next.UpdateWorkspace(ctx, req)
	if err != nil {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionWorkspaceUpdate,
			audit.WithDescription("Failed to update workspace."),
			audit.WithWorkspaceID(req.Id),
			audit.WithErrorMessage(err.Error()),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("workspace_id", req.Id),
		)
	} else {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionWorkspaceUpdate,
			audit.WithDescription(fmt.Sprintf("Updated workspace with ID %d.", req.Id)),
			audit.WithWorkspaceID(req.Id),
			audit.WithDetail("workspace_id", req.Id),
			audit.WithDetail("workspace_name", req.Name),
		)
	}

	return res, err
}

func (c workspaceControllerAudit) DeleteWorkspace(ctx context.Context, req *chorus.DeleteWorkspaceRequest) (*chorus.DeleteWorkspaceReply, error) {
	res, err := c.next.DeleteWorkspace(ctx, req)
	if err != nil {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionWorkspaceDelete,
			audit.WithDescription("Failed to delete workspace."),
			audit.WithWorkspaceID(req.Id),
			audit.WithErrorMessage(err.Error()),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("workspace_id", req.Id),
		)
	} else {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionWorkspaceDelete,
			audit.WithDescription(fmt.Sprintf("Deleted workspace with ID %d.", req.Id)),
			audit.WithWorkspaceID(req.Id),
			audit.WithDetail("workspace_id", req.Id),
		)
	}

	return res, err
}

func (c workspaceControllerAudit) ManageUserRoleInWorkspace(ctx context.Context, req *chorus.ManageUserRoleInWorkspaceRequest) (*chorus.ManageUserRoleInWorkspaceReply, error) {
	res, err := c.next.ManageUserRoleInWorkspace(ctx, req)
	if err != nil {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionWorkspaceMemberAdd,
			audit.WithDescription(fmt.Sprintf("Failed to add user %d to workspace %d with role %s.", req.UserId, req.Id, req.Role)),
			audit.WithWorkspaceID(req.Id),
			audit.WithErrorMessage(err.Error()),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("workspace_id", req.Id),
			audit.WithDetail("user_id", req.UserId),
			audit.WithDetail("role", req.Role),
		)
	} else {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionWorkspaceMemberAdd,
			audit.WithDescription(fmt.Sprintf("Added user %d to workspace %d with role %s.", req.UserId, req.Id, req.Role)),
			audit.WithWorkspaceID(req.Id),
			audit.WithDetail("workspace_id", req.Id),
			audit.WithDetail("user_id", req.UserId),
			audit.WithDetail("role", req.Role),
		)
	}

	return res, err
}

func (c workspaceControllerAudit) RemoveUserFromWorkspace(ctx context.Context, req *chorus.RemoveUserFromWorkspaceRequest) (*chorus.RemoveUserFromWorkspaceReply, error) {
	res, err := c.next.RemoveUserFromWorkspace(ctx, req)
	if err != nil {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionWorkspaceMemberRemove,
			audit.WithDescription(fmt.Sprintf("Failed to remove user %d from workspace %d.", req.UserId, req.Id)),
			audit.WithWorkspaceID(req.Id),
			audit.WithErrorMessage(err.Error()),
			audit.WithGRPCStatusCode(grpc.ErrorCode(err)),
			audit.WithDetail("workspace_id", req.Id),
			audit.WithDetail("user_id", req.UserId),
		)
	} else {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionWorkspaceMemberRemove,
			audit.WithDescription(fmt.Sprintf("Removed user %d from workspace %d.", req.UserId, req.Id)),
			audit.WithWorkspaceID(req.Id),
			audit.WithDetail("workspace_id", req.Id),
			audit.WithDetail("user_id", req.UserId),
		)
	}

	return res, err
}
