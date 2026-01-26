package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
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
	// Audit logic can be added here
	return c.next.ListWorkspaces(ctx, req)
}

func (c workspaceControllerAudit) CreateWorkspace(ctx context.Context, req *chorus.Workspace) (*chorus.CreateWorkspaceReply, error) {
	// Audit logic can be added here
	return c.next.CreateWorkspace(ctx, req)
}

func (c workspaceControllerAudit) GetWorkspace(ctx context.Context, req *chorus.GetWorkspaceRequest) (*chorus.GetWorkspaceReply, error) {
	// Audit logic can be added here
	return c.next.GetWorkspace(ctx, req)
}

func (c workspaceControllerAudit) UpdateWorkspace(ctx context.Context, req *chorus.Workspace) (*chorus.UpdateWorkspaceReply, error) {
	// Audit logic can be added here
	return c.next.UpdateWorkspace(ctx, req)
}

func (c workspaceControllerAudit) DeleteWorkspace(ctx context.Context, req *chorus.DeleteWorkspaceRequest) (*chorus.DeleteWorkspaceReply, error) {
	// Audit logic can be added here
	return c.next.DeleteWorkspace(ctx, req)
}

func (c workspaceControllerAudit) ManageUserRoleInWorkspace(ctx context.Context, req *chorus.ManageUserRoleInWorkspaceRequest) (*chorus.ManageUserRoleInWorkspaceReply, error) {
	// Audit logic can be added here
	return c.next.ManageUserRoleInWorkspace(ctx, req)
}

func (c workspaceControllerAudit) RemoveUserFromWorkspace(ctx context.Context, req *chorus.RemoveUserFromWorkspaceRequest) (*chorus.RemoveUserFromWorkspaceReply, error) {
	// Audit logic can be added here
	return c.next.RemoveUserFromWorkspace(ctx, req)
}
