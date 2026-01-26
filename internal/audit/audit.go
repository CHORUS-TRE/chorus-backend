package audit

import (
	"context"
	"time"

	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/correlation"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/service"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
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
	// Create the audit entry synchronously
	entry := NewEntry(ctx, action, opts...)

	// Write the audit entry asynchronously
	go func() {
		_, err := writer.Record(context.Background(), entry)
		if err != nil {
			logger.TechLog.Error(context.Background(), "failed to record audit entry",
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

// WithMethod sets the gRPC method name in entry details map
func WithGRPCMethod(method string) Option {
	return WithDetail("grpc_method", method)
}

// WithStatusCode sets the gRPC status code in entry details map
func WithGRPCStatusCode(code codes.Code) Option {
	return WithDetail("grpc_status_code", int(code))
}

// WithErrorMessage sets the error message in entry details map
func WithErrorMessage(errMsg string) Option {
	return WithDetail("error_message", errMsg)
}
