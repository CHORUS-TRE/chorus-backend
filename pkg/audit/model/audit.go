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
	Action        AuditAction
	ResourceType  AuditResourceType
	ResourceID    uint64
	CorrelationID string
	Method        string         // gRPC method name
	StatusCode    int            // gRPC status code
	ErrorMessage  string         // Error message if any
	Details       map[string]any // JSONB for flexible querying
	CreatedAt     time.Time
}

// AuditFilter for querying audit entries
type AuditFilter struct {
	TenantID     *uint64
	UserID       *uint64
	ResourceType *AuditResourceType
	ResourceID   *uint64
	Action       *AuditAction
	FromTime     *time.Time
	ToTime       *time.Time
	Limit        int
	Offset       int
}
