package middleware

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/audit"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	audit_service "github.com/CHORUS-TRE/chorus-backend/pkg/audit/service"
)

var _ chorus.TermsOfUseServiceServer = (*termsOfUseControllerAudit)(nil)

type termsOfUseControllerAudit struct {
	next        chorus.TermsOfUseServiceServer
	auditWriter audit_service.AuditWriter
}

func NewTermsOfUseAuditMiddleware(auditWriter audit_service.AuditWriter) func(chorus.TermsOfUseServiceServer) chorus.TermsOfUseServiceServer {
	return func(next chorus.TermsOfUseServiceServer) chorus.TermsOfUseServiceServer {
		return &termsOfUseControllerAudit{next: next, auditWriter: auditWriter}
	}
}

func (c termsOfUseControllerAudit) CreateTermsOfUseVersion(ctx context.Context, req *chorus.CreateTermsOfUseVersionRequest) (*chorus.CreateTermsOfUseVersionReply, error) {
	res, err := c.next.CreateTermsOfUseVersion(ctx, req)

	opts := []audit.Option{}
	if err != nil {
		opts = append(opts, audit.WithDescription("Failed to create terms of use version."), audit.WithError(err))
	} else {
		opts = append(opts, audit.WithDescription(fmt.Sprintf("Created terms of use version %d.", res.Result.TermsOfUseVersion.Id)))
	}
	audit.Record(ctx, c.auditWriter, model.AuditActionTermsOfUseVersionCreate, opts...)

	return res, err
}

func (c termsOfUseControllerAudit) UpdateTermsOfUseVersion(ctx context.Context, req *chorus.UpdateTermsOfUseVersionRequest) (*chorus.UpdateTermsOfUseVersionReply, error) {
	res, err := c.next.UpdateTermsOfUseVersion(ctx, req)

	opts := []audit.Option{audit.WithDetail("version_id", req.Id)}
	if err != nil {
		opts = append(opts, audit.WithDescription(fmt.Sprintf("Failed to update terms of use version %d.", req.Id)), audit.WithError(err))
	} else {
		opts = append(opts, audit.WithDescription(fmt.Sprintf("Updated terms of use version %d.", req.Id)))
	}
	audit.Record(ctx, c.auditWriter, model.AuditActionTermsOfUseVersionUpdate, opts...)

	return res, err
}

func (c termsOfUseControllerAudit) PublishTermsOfUseVersion(ctx context.Context, req *chorus.PublishTermsOfUseVersionRequest) (*chorus.PublishTermsOfUseVersionReply, error) {
	res, err := c.next.PublishTermsOfUseVersion(ctx, req)

	opts := []audit.Option{audit.WithDetail("version_id", req.Id)}
	if err != nil {
		opts = append(opts, audit.WithDescription(fmt.Sprintf("Failed to publish terms of use version %d.", req.Id)), audit.WithError(err))
	} else {
		opts = append(opts, audit.WithDescription(fmt.Sprintf("Published terms of use version %d.", req.Id)))
	}
	audit.Record(ctx, c.auditWriter, model.AuditActionTermsOfUseVersionPublish, opts...)

	return res, err
}

func (c termsOfUseControllerAudit) GetTermsOfUseVersion(ctx context.Context, req *chorus.GetTermsOfUseVersionRequest) (*chorus.GetTermsOfUseVersionReply, error) {
	return c.next.GetTermsOfUseVersion(ctx, req)
}

func (c termsOfUseControllerAudit) ListTermsOfUseVersions(ctx context.Context, req *chorus.ListTermsOfUseVersionsRequest) (*chorus.ListTermsOfUseVersionsReply, error) {
	return c.next.ListTermsOfUseVersions(ctx, req)
}

func (c termsOfUseControllerAudit) GetCurrentTermsOfUseVersion(ctx context.Context, req *chorus.GetCurrentTermsOfUseVersionRequest) (*chorus.GetCurrentTermsOfUseVersionReply, error) {
	return c.next.GetCurrentTermsOfUseVersion(ctx, req)
}

func (c termsOfUseControllerAudit) ListTermsOfUseAcceptances(ctx context.Context, req *chorus.ListTermsOfUseAcceptancesRequest) (*chorus.ListTermsOfUseAcceptancesReply, error) {
	return c.next.ListTermsOfUseAcceptances(ctx, req)
}

func (c termsOfUseControllerAudit) GetMyTermsOfUseStatus(ctx context.Context, req *chorus.GetMyTermsOfUseStatusRequest) (*chorus.GetMyTermsOfUseStatusReply, error) {
	return c.next.GetMyTermsOfUseStatus(ctx, req)
}

func (c termsOfUseControllerAudit) AcceptTermsOfUse(ctx context.Context, req *chorus.AcceptTermsOfUseRequest) (*chorus.AcceptTermsOfUseReply, error) {
	res, err := c.next.AcceptTermsOfUse(ctx, req)

	opts := []audit.Option{}
	if err != nil {
		opts = append(opts, audit.WithDescription("Failed to accept terms of use."), audit.WithError(err))
	} else {
		opts = append(opts, audit.WithDescription(fmt.Sprintf("Accepted terms of use version %d.", res.Result.TermsOfUseAcceptance.TermsOfUseVersionId)))
	}
	audit.Record(ctx, c.auditWriter, model.AuditActionTermsOfUseAccept, opts...)

	return res, err
}
