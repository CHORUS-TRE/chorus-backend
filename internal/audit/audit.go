package audit

import (
	"context"
	"time"

	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/utils/correlation"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/service"
	"google.golang.org/grpc/codes"
)

// Option modifies an AuditEntry
type Option func(entry *model.AuditEntry)

// NewEntry creates an AuditEntry with common fields extracted from context
func NewEntry(ctx context.Context, action model.AuditAction, resourceType model.AuditResourceType, resourceID uint64, opts ...Option) *model.AuditEntry {
	entry := &model.AuditEntry{
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		CreatedAt:    time.Now().UTC(),
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

// Record is a convenience function that creates an entry and writes it
func Record(ctx context.Context, writer service.AuditWriter, action model.AuditAction, resourceType model.AuditResourceType, resourceID uint64, opts ...Option) error {
	entry := NewEntry(ctx, action, resourceType, resourceID, opts...)
	return writer.Record(ctx, entry)
}

func WithWorkspaceID(workspaceID uint64) Option {
	return func(entry *model.AuditEntry) {
		entry.WorkspaceID = workspaceID
	}
}

func WithWorkbenchID(workbenchID uint64) Option {
	return func(entry *model.AuditEntry) {
		entry.WorkbenchID = workbenchID
	}
}

// WithMethod sets the gRPC method name
func WithMethod(method string) Option {
	return func(entry *model.AuditEntry) {
		entry.Method = method
	}
}

// WithStatusCode sets the gRPC status code
func WithStatusCode(code codes.Code) Option {
	return func(entry *model.AuditEntry) {
		entry.StatusCode = int(code)
	}
}

// WithErrorMessage sets the error message
func WithErrorMessage(errMsg string) Option {
	return func(entry *model.AuditEntry) {
		entry.ErrorMessage = errMsg
	}
}

// WithDescription sets the human readable description
func WithDescription(desc string) Option {
	return func(entry *model.AuditEntry) {
		entry.Description = desc
	}
}

// WithDetails sets the details map
func WithDetails(details map[string]any) Option {
	return func(entry *model.AuditEntry) {
		entry.Details = details
	}
}
