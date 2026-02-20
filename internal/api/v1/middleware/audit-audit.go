package middleware

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/audit"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/service"
)

var _ chorus.AuditServiceServer = (*auditControllerAudit)(nil)

type auditControllerAudit struct {
	next        chorus.AuditServiceServer
	auditWriter service.AuditWriter
}

func NewAuditAuditMiddleware(auditWriter service.AuditWriter) func(chorus.AuditServiceServer) chorus.AuditServiceServer {
	return func(next chorus.AuditServiceServer) chorus.AuditServiceServer {
		return &auditControllerAudit{
			next:        next,
			auditWriter: auditWriter,
		}
	}
}

func (c auditControllerAudit) ListAuditEntries(ctx context.Context, req *chorus.ListAuditEntriesRequest) (*chorus.ListAuditEntriesReply, error) {
	res, err := c.next.ListAuditEntries(ctx, req)

	opts := []audit.Option{}

	if req.Filter != nil {
		opts = append(opts,
			audit.WithDetail("filter", req.Filter),
		)
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription("Failed to list audit entries."),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription("Listed audit entries."),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionPlatformAuditList, opts...)

	return res, err
}

func (c auditControllerAudit) ListWorkspaceAudit(ctx context.Context, req *chorus.ListEntityAuditRequest) (*chorus.ListAuditEntriesReply, error) {
	res, err := c.next.ListWorkspaceAudit(ctx, req)

	if err != nil {
		audit.Record(ctx, c.auditWriter, model.AuditActionWorkspaceAuditList,
			audit.WithWorkspaceID(req.Id),
			audit.WithDetail("workspace_id", req.Id),
			audit.WithDescription(fmt.Sprintf("Failed to list audit entries for workspace %d.", req.Id)),
			audit.WithError(err),
		)
	}

	return res, err
}

func (c auditControllerAudit) ListWorkbenchAudit(ctx context.Context, req *chorus.ListEntityAuditRequest) (*chorus.ListAuditEntriesReply, error) {
	res, err := c.next.ListWorkbenchAudit(ctx, req)

	if err != nil {
		audit.Record(ctx, c.auditWriter, model.AuditActionWorkbenchAuditList,
			audit.WithWorkbenchID(req.Id),
			audit.WithDetail("workbench_id", req.Id),
			audit.WithDescription(fmt.Sprintf("Failed to list audit entries for workbench %d.", req.Id)),
			audit.WithError(err),
		)
	}

	return res, err
}

func (c auditControllerAudit) ListUserAudit(ctx context.Context, req *chorus.ListEntityAuditRequest) (*chorus.ListAuditEntriesReply, error) {
	res, err := c.next.ListUserAudit(ctx, req)

	if err != nil {
		audit.Record(ctx, c.auditWriter, model.AuditActionUserAuditList,
			audit.WithDetail("user_id", req.Id),
			audit.WithDescription(fmt.Sprintf("Failed to list audit entries for user %d.", req.Id)),
			audit.WithError(err),
		)
	}

	return res, err
}
