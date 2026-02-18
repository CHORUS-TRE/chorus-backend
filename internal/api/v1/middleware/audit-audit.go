package middleware

import (
	"context"

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
			audit.WithDetail("filter_workspace_id", req.Filter.WorkspaceId),
			audit.WithDetail("filter_workbench_id", req.Filter.WorkbenchId),
			audit.WithDetail("filter_user_id", req.Filter.UserId),
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

	audit.Record(ctx, c.auditWriter, model.AuditActionAuditList, opts...)

	return res, err
}
