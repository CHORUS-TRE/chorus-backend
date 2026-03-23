package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// AuditEntry represents a single audit log entry
type AuditEntry struct {
	ID uint64

	TenantID      uint64
	ActorID       uint64 // Who performed the action (from JWT)
	ActorUsername string // Username of the actor (from JWT)
	CorrelationID string // To correlate with other logs

	Action      AuditAction  // Type of action performed
	WorkspaceID uint64       // Context: workspace acted upon
	WorkbenchID uint64       // Context: workbench acted upon
	UserID      uint64       // Context: user acted upon
	Description string       // Human readable description of the action
	Details     AuditDetails // JSONB for flexible querying

	CreatedAt time.Time
}

type AuditDetails map[string]any

func (d *AuditDetails) Scan(value any) error {
	if value == nil {
		*d = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unable to scan AuditDetails: %v", value)
	}
	var details map[string]any
	if err := json.Unmarshal(bytes, &details); err != nil {
		return fmt.Errorf("unable to unmarshal AuditDetails: %w", err)
	}
	*d = details
	return nil
}

func (d AuditDetails) Value() (driver.Value, error) {
	if d == nil {
		return nil, nil
	}
	return json.Marshal(d)
}

// AuditFilter for querying audit entries
type AuditFilter struct {
	ActorID     uint64
	UserID      uint64
	WorkspaceID uint64
	WorkbenchID uint64
	Action      AuditAction
	FromTime    time.Time
	ToTime      time.Time
}

func (AuditEntry) IsValidSortType(sortType string) bool {
	switch sortType {
	case "createdat":
		return true
	default:
		return false
	}
}
