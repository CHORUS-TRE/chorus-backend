package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/audit"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	audit_service "github.com/CHORUS-TRE/chorus-backend/pkg/audit/service"
)

var _ chorus.PlatformSettingsServiceServer = (*platformSettingsControllerAudit)(nil)

type platformSettingsControllerAudit struct {
	next        chorus.PlatformSettingsServiceServer
	auditWriter audit_service.AuditWriter
}

func NewPlatformSettingsAuditMiddleware(auditWriter audit_service.AuditWriter) func(chorus.PlatformSettingsServiceServer) chorus.PlatformSettingsServiceServer {
	return func(next chorus.PlatformSettingsServiceServer) chorus.PlatformSettingsServiceServer {
		return &platformSettingsControllerAudit{
			next:        next,
			auditWriter: auditWriter,
		}
	}
}

func (c platformSettingsControllerAudit) GetPlatformSettings(ctx context.Context, req *chorus.GetPlatformSettingsRequest) (*chorus.GetPlatformSettingsReply, error) {
	// No audit for getting platform settings - this is public information
	return c.next.GetPlatformSettings(ctx, req)
}

func (c platformSettingsControllerAudit) UpdatePlatformSettings(ctx context.Context, req *chorus.UpdatePlatformSettingsRequest) (*chorus.UpdatePlatformSettingsReply, error) {
	res, err := c.next.UpdatePlatformSettings(ctx, req)

	opts := []audit.Option{}

	if err != nil {
		opts = append(opts,
			audit.WithDescription("Failed to update platform settings."),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription("Updated platform settings."),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionPlatformSettingsUpdate, opts...)

	return res, err
}
