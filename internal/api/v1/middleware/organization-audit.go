package middleware

import (
	"context"
	"fmt"

	"google.golang.org/genproto/googleapis/api/httpbody"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/audit"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	audit_service "github.com/CHORUS-TRE/chorus-backend/pkg/audit/service"
)

var _ chorus.OrganizationServiceServer = (*organizationControllerAudit)(nil)

type organizationControllerAudit struct {
	next        chorus.OrganizationServiceServer
	auditWriter audit_service.AuditWriter
}

func NewOrganizationAuditMiddleware(auditWriter audit_service.AuditWriter) func(chorus.OrganizationServiceServer) chorus.OrganizationServiceServer {
	return func(next chorus.OrganizationServiceServer) chorus.OrganizationServiceServer {
		return &organizationControllerAudit{
			next:        next,
			auditWriter: auditWriter,
		}
	}
}

func (c organizationControllerAudit) ListOrganizations(ctx context.Context, req *chorus.ListOrganizationsRequest) (*chorus.ListOrganizationsReply, error) {
	// No audit for listing organizations - high-volume read, not a sensitive action.
	return c.next.ListOrganizations(ctx, req)
}

func (c organizationControllerAudit) GetOrganization(ctx context.Context, req *chorus.GetOrganizationRequest) (*chorus.GetOrganizationReply, error) {
	// No audit for getting an organization - high-volume read, not a sensitive action.
	return c.next.GetOrganization(ctx, req)
}

func (c organizationControllerAudit) GetOrganizationLogo(ctx context.Context, req *chorus.GetOrganizationLogoRequest) (*httpbody.HttpBody, error) {
	// No audit for getting an organization's logo - same rationale as GetOrganization.
	return c.next.GetOrganizationLogo(ctx, req)
}

func (c organizationControllerAudit) CreateOrganization(ctx context.Context, req *chorus.Organization) (*chorus.CreateOrganizationReply, error) {
	res, err := c.next.CreateOrganization(ctx, req)

	opts := []audit.Option{}
	if err != nil {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Failed to create organization %q.", req.GetName())),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Created organization %q.", res.Result.Organization.Name)),
			audit.WithDetail("organization_id", res.Result.Organization.Id),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionOrganizationCreate, opts...)

	return res, err
}

func (c organizationControllerAudit) UpdateOrganization(ctx context.Context, req *chorus.Organization) (*chorus.UpdateOrganizationReply, error) {
	res, err := c.next.UpdateOrganization(ctx, req)

	opts := []audit.Option{audit.WithDetail("organization_id", req.GetId())}
	if err != nil {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Failed to update organization %q (ID %d).", req.GetName(), req.GetId())),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Updated organization %q (ID %d).", req.GetName(), req.GetId())),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionOrganizationUpdate, opts...)

	return res, err
}

func (c organizationControllerAudit) DeleteOrganization(ctx context.Context, req *chorus.DeleteOrganizationRequest) (*chorus.DeleteOrganizationReply, error) {
	res, err := c.next.DeleteOrganization(ctx, req)

	opts := []audit.Option{audit.WithDetail("organization_id", req.GetId())}
	if err != nil {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Failed to delete organization %d.", req.GetId())),
			audit.WithError(err),
		)
	} else {
		opts = append(opts,
			audit.WithDescription(fmt.Sprintf("Deleted organization %d.", req.GetId())),
		)
	}

	audit.Record(ctx, c.auditWriter, model.AuditActionOrganizationDelete, opts...)

	return res, err
}
