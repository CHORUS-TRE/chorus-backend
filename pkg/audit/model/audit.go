package model

import "time"

// AuditAction represents the type of action performed
type AuditAction string

const (
	AuditActionCreate AuditAction = "CREATE"
	AuditActionRead   AuditAction = "READ"
	AuditActionUpdate AuditAction = "UPDATE"
	AuditActionDelete AuditAction = "DELETE"
	AuditActionList   AuditAction = "LIST"
)

// AuditResourceType represents the type of resource being audited
type AuditResourceType string

const (
	AuditResourceUser        AuditResourceType = "USER"
	AuditResourceApp         AuditResourceType = "APP"
	AuditResourceAppInstance AuditResourceType = "APP_INSTANCE"
	AuditResourceWorkbench   AuditResourceType = "WORKBENCH"
	AuditResourceWorkspace   AuditResourceType = "WORKSPACE"
)

// AuditEntry represents a single audit log entry
type AuditEntry struct {
	ID            uint64
	TenantID      uint64
	UserID        uint64
	Username      string
	Action        AuditAction       // Type of action performed
	ResourceType  AuditResourceType // Type of the resource
	ResourceID    uint64            // ID of the resource
	WorkspaceID   uint64            // ID of the workspace
	WorkbenchID   uint64            // ID of the workbench
	CorrelationID string            // To correlate with other logs
	Method        string            // gRPC method name
	StatusCode    int               // gRPC status code
	ErrorMessage  string            // Error message if any
	Description   string            // Human readable description of the action
	Details       map[string]any    // JSONB for flexible querying
	CreatedAt     time.Time
}

// AuditFilter for querying audit entries
type AuditFilter struct {
	TenantID     *uint64
	UserID       *uint64
	ResourceType *AuditResourceType
	ResourceID   *uint64
	Action       *AuditAction
	WorkspaceID  *uint64
	WorkbenchID  *uint64
	FromTime     *time.Time
	ToTime       *time.Time
}
