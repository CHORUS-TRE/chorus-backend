package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	authorization "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	authorization_service "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"
)

var _ chorus.AuditServiceServer = (*auditControllerAuthorization)(nil)

type auditControllerAuthorization struct {
	Authorization
	next chorus.AuditServiceServer
}

func AuditAuthorizing(logger *logger.ContextLogger, authorizer authorization_service.Authorizer) func(chorus.AuditServiceServer) chorus.AuditServiceServer {
	return func(next chorus.AuditServiceServer) chorus.AuditServiceServer {
		return &auditControllerAuthorization{
			Authorization: Authorization{
				logger:     logger,
				authorizer: authorizer,
			},
			next: next,
		}
	}
}

func (c auditControllerAuthorization) ListAuditEntries(ctx context.Context, req *chorus.ListAuditEntriesRequest) (*chorus.ListAuditEntriesReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionAuditPlatform)
	if err != nil {
		return nil, err
	}

	return c.next.ListAuditEntries(ctx, req)
}

func (c auditControllerAuthorization) ListWorkspaceAudit(ctx context.Context, req *chorus.ListEntityAuditRequest) (*chorus.ListAuditEntriesReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionAuditWorkspace, authorization.WithWorkspace(req.Id))
	if err != nil {
		return nil, err
	}

	return c.next.ListWorkspaceAudit(ctx, req)
}

func (c auditControllerAuthorization) ListWorkbenchAudit(ctx context.Context, req *chorus.ListEntityAuditRequest) (*chorus.ListAuditEntriesReply, error) {
	err := c.IsAuthorized(ctx, authorization.PermissionAuditWorkbench, authorization.WithWorkbench(req.Id))
	if err != nil {
		return nil, err
	}

	return c.next.ListWorkbenchAudit(ctx, req)
}

func (c auditControllerAuthorization) ListUserAudit(ctx context.Context, req *chorus.ListEntityAuditRequest) (*chorus.ListAuditEntriesReply, error) {
	if req.Filter != nil && req.Filter.WorkspaceId != 0 {
		// Workspace-scoped: caller must have audit permission for that specific workspace.
		err := c.IsAuthorized(ctx, authorization.PermissionAuditWorkspace, authorization.WithWorkspace(req.Filter.WorkspaceId))
		if err != nil {
			return nil, err
		}
	} else if req.Filter != nil && req.Filter.WorkbenchId != 0 {
		// Workbench-scoped: caller must have audit permission for that specific workbench.
		err := c.IsAuthorized(ctx, authorization.PermissionAuditWorkbench, authorization.WithWorkbench(req.Filter.WorkbenchId))
		if err != nil {
			return nil, err
		}
	} else {
		// No scope: self-audit (authenticated) or platform-level user audit permission required.
		err := c.IsAuthorized(ctx, authorization.PermissionAuditUser, authorization.WithUser(req.Id))
		if err != nil {
			return nil, err
		}
	}

	return c.next.ListUserAudit(ctx, req)
}
