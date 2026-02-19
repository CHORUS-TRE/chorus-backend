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
