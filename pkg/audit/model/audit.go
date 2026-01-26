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
	UserID        uint64
	Username      string
	CorrelationID string // To correlate with other logs

	Action      AuditAction  // Type of action performed
	WorkspaceID uint64       // ID of the workspace
	WorkbenchID uint64       // ID of the workbench
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
	UserID      uint64
	Action      AuditAction
	WorkspaceID uint64
	WorkbenchID uint64
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
