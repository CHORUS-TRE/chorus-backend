package audit

import (
	"context"
	"runtime/debug"
	"time"

	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/correlation"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/grpc"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/service"
	"go.uber.org/zap"
)

// Option modifies an AuditEntry
type Option func(entry *model.AuditEntry)

// NewEntry creates an AuditEntry with common fields extracted from context
func NewEntry(ctx context.Context, action model.AuditAction, opts ...Option) *model.AuditEntry {
	entry := &model.AuditEntry{
		Action:    action,
		Details:   model.AuditDetails{},
		CreatedAt: time.Now().UTC(),
	}

	// Auto-extract from context
	if claims, ok := ctx.Value(jwt_model.JWTClaimsContextKey).(*jwt_model.JWTClaims); ok {
		entry.TenantID = claims.TenantID
		entry.UserID = claims.ID
		entry.Username = claims.Username
	}

	if cid, ok := ctx.Value(correlation.CorrelationIDContextKey{}).(string); ok {
		entry.CorrelationID = cid
	}

	// Apply options
	for _, opt := range opts {
		opt(entry)
	}

	return entry
}

// Record is a convenience function that creates an entry and writes it asynchronously
func Record(ctx context.Context, writer service.AuditWriter, action model.AuditAction, opts ...Option) {
	baseCtx := context.WithoutCancel(ctx)

	// Create the audit entry synchronously
	entry := NewEntry(baseCtx, action, opts...)

	// Write the audit entry asynchronously
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.TechLog.Error(baseCtx, "panic while recording audit entry",
					zap.Any("recover", r),
					zap.ByteString("stack", debug.Stack()),
					zap.Any("entry", entry),
				)
			}
		}()

		_, err := writer.Record(baseCtx, entry)
		if err != nil {
			logger.TechLog.Error(baseCtx, "failed to record audit entry",
				zap.Error(err),
				zap.Any("entry", entry),
			)
		}
	}()
}

// WithWorkspaceID sets the workspace ID
func WithWorkspaceID(workspaceID uint64) Option {
	return func(entry *model.AuditEntry) {
		entry.WorkspaceID = workspaceID
	}
}

// WithWorkbenchID sets the workbench ID
func WithWorkbenchID(workbenchID uint64) Option {
	return func(entry *model.AuditEntry) {
		entry.WorkbenchID = workbenchID
	}
}

// WithDescription sets the human readable description
func WithDescription(desc string) Option {
	return func(entry *model.AuditEntry) {
		entry.Description = desc
	}
}

// WithDetail adds a key-value pair to the Details map
func WithDetail(key string, value any) Option {
	return func(entry *model.AuditEntry) {
		entry.Details[key] = value
	}
}

// WithError sets error message and gRPC status code from an error
func WithError(err error) Option {
	return func(entry *model.AuditEntry) {
		entry.Details["error_message"] = err.Error()
		entry.Details["grpc_status_code"] = int(grpc.ErrorCode(err))
	}
}
