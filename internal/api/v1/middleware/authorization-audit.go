package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/audit"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/service"
)

var _ chorus.AuthorizationServiceServer = (*authorizationControllerAudit)(nil)

type authorizationControllerAudit struct {
	next        chorus.AuthorizationServiceServer
	auditWriter service.AuditWriter
}

func NewAuthorizationAuditMiddleware(auditWriter service.AuditWriter) func(chorus.AuthorizationServiceServer) chorus.AuthorizationServiceServer {
	return func(next chorus.AuthorizationServiceServer) chorus.AuthorizationServiceServer {
		return &authorizationControllerAudit{
			auditWriter: auditWriter,
			next:        next,
		}
	}
}

func (c authorizationControllerAudit) ListRoles(ctx context.Context, req *chorus.ListRolesRequest) (*chorus.ListRolesReply, error) {
	res, err := c.next.ListRoles(ctx, req)

	if err != nil {
		audit.Record(ctx, c.auditWriter, model.AuditActionRoleList,
			audit.WithDescription("Failed to list roles."),
			audit.WithError(err),
		)
	}

	return res, err
}

func (c authorizationControllerAudit) ListPermissions(ctx context.Context, req *chorus.ListPermissionsRequest) (*chorus.ListPermissionsReply, error) {
	res, err := c.next.ListPermissions(ctx, req)

	if err != nil {
		audit.Record(ctx, c.auditWriter, model.AuditActionPermissionList,
			audit.WithDescription("Failed to list permissions."),
			audit.WithError(err),
		)
	}

	return res, err
}

func (c authorizationControllerAudit) CreateDynamicRole(ctx context.Context, req *chorus.CreateDynamicRoleRequest) (*chorus.CreateDynamicRoleReply, error) {
	res, err := c.next.CreateDynamicRole(ctx, req)

	baseOptions := []audit.Option{
		audit.WithDetail("role_name", req.Role.Name),
		audit.WithDetail("role_scope", req.Role.Scope),
		audit.WithDetail("validation_context", req.ValidationContext),
	}

	if err != nil {
		audit.Record(ctx, c.auditWriter, model.AuditActionRoleCreate,
			audit.WithDescription("Failed to create dynamic role."),
			audit.WithOptions(baseOptions...),
			audit.WithError(err),
		)
	} else {
		audit.Record(ctx, c.auditWriter, model.AuditActionRoleCreate,
			audit.WithDescription("Created dynamic role."),
			audit.WithOptions(baseOptions...),
		)
	}

	return res, err
}
