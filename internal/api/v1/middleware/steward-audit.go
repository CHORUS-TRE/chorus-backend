package middleware

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/audit"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/service"
	empty "google.golang.org/protobuf/types/known/emptypb"
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

func (c stewardControllerAudit) InitializeTenant(ctx context.Context, req *chorus.InitializeTenantRequest) (*empty.Empty, error) {
	res, err := c.next.InitializeTenant(ctx, req)
	if err != nil {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionTenantInitialize,
			audit.WithDescription(fmt.Sprintf("Failed to initialize tenant %d.", req.TenantId)),
			audit.WithError(err),
			audit.WithDetail("tenant_id", req.TenantId),
		)
	} else {
		audit.Record(ctx, c.auditWriter,
			model.AuditActionTenantInitialize,
			audit.WithDescription(fmt.Sprintf("Initialized tenant %d.", req.TenantId)),
			audit.WithDetail("tenant_id", req.TenantId),
		)
	}

	return res, err
}
