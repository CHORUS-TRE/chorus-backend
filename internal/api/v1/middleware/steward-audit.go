package middleware

import (
	"context"
	"fmt"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/audit"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/service"
)

var _ chorus.StewardServiceServer = (*stewardControllerAudit)(nil)

type stewardControllerAudit struct {
	next        chorus.StewardServiceServer
	auditWriter service.AuditWriter
}

func NewStewardAuditMiddleware(auditWriter service.AuditWriter) func(chorus.StewardServiceServer) chorus.StewardServiceServer {
	return func(next chorus.StewardServiceServer) chorus.StewardServiceServer {
		return &stewardControllerAudit{
			next:        next,
			auditWriter: auditWriter,
		}
	}
}

func (c stewardControllerAudit) InitializeTenant(ctx context.Context, req *chorus.InitializeTenantRequest) (*chorus.InitializeTenantReply, error) {
	res, err := c.next.InitializeTenant(ctx, req)

	opts := []audit.Option{
		audit.WithDetail("name", req.Name),
	}

	if err != nil {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Failed to initialize tenant %q.", req.Name)),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDetail("id", res.Result.Id),
			audit.WithDescription(fmt.Sprintf("Initialized tenant %q with id %d.", req.Name, res.Result.Id)),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionTenantInitialize, opts...)

	return res, err
}
