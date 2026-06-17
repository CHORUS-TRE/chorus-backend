package model

import (
	"encoding/json"
	"fmt"
	"time"
)

// NetworkPolicyMode defines the network policy type for a workspace.
type NetworkPolicyMode string

const (
	NetworkPolicyOpen          NetworkPolicyMode = "Open"
	NetworkPolicyAirgapped     NetworkPolicyMode = "Airgapped"
	NetworkPolicyFQDNAllowlist NetworkPolicyMode = "FQDNAllowlist"
)

func (n NetworkPolicyMode) String() string {
	return string(n)
}

// ClipboardMode defines the clipboard direction for a workspace.
type ClipboardMode string

const (
	ClipboardDisabled ClipboardMode = "disabled"
	ClipboardToServer ClipboardMode = "to-server"
	ClipboardToClient ClipboardMode = "to-client"
	ClipboardBoth     ClipboardMode = "both"
)

func (c ClipboardMode) String() string {
	return string(c)
}

// WorkspaceVisibility defines the visibility for a workspace.
type WorkspaceVisibility string

const (
	WorkspaceVisibilityPrivate WorkspaceVisibility = "private"
	WorkspaceVisibilityPublic  WorkspaceVisibility = "public"
)

func (v WorkspaceVisibility) String() string {
	return string(v)
}

// WorkspaceStatus defines the status for a workspace.
type WorkspaceStatus string

const (
	WorkspaceStatusActive   WorkspaceStatus = "active"
	WorkspaceStatusInactive WorkspaceStatus = "inactive"
	WorkspaceStatusDeleted  WorkspaceStatus = "deleted"
)

func (s WorkspaceStatus) String() string {
	return string(s)
}

// Workspace maps an entry in the 'workspaces' database table.
type Workspace struct {
	ID uint64

	TenantID uint64
	UserID   uint64

	Name        string
	ShortName   string
	Description string

	Status WorkspaceStatus

	IsMain bool

	// Network policy fields
	NetworkPolicy        NetworkPolicyMode
	AllowedFQDNs         StringSlice
	NetworkPolicyStatus  string
	NetworkPolicyMessage string

	// Clipboard (workspace-wide default for workbenches)
	Clipboard ClipboardMode

	// Visibility defines whether the workspace is private or public.
	// Public workspaces expose certain information (e.g., name, description, contact information) to users outside the workspace and can be listed in a public catalog.
	Visibility WorkspaceVisibility
	// ContactUserID defines the user to list as contact point when the workspace is public
	ContactUserID *uint64

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// PublicWorkspace represents information made available to users when a workspace is marked as public.
type PublicWorkspace struct {
	ID       uint64
	TenantID uint64

	Name        string
	ShortName   string
	Description string
	Status      WorkspaceStatus

	ContactUsername  string
	ContactFirstName string
	ContactLastName  string
	ContactEmail     string

	CreatedAt time.Time
	UpdatedAt time.Time
}

// JSONMap is a generic map type that handles JSONB serialization for sqlx.
type JSONMap[T any] map[string]T

func (j *JSONMap[T]) Scan(src interface{}) error {
	if src == nil {
		*j = make(JSONMap[T])
		return nil
	}
	var source []byte
	switch v := src.(type) {
	case []byte:
		source = v
	case string:
		source = []byte(v)
	default:
		return fmt.Errorf("unsupported type for JSONMap: %T", src)
	}
	m := make(JSONMap[T])
	if err := json.Unmarshal(source, &m); err != nil {
		return fmt.Errorf("unable to unmarshal JSONMap: %w", err)
	}
	*j = m
	return nil
}

func (j JSONMap[T]) Value() (interface{}, error) {
	if j == nil {
		return "{}", nil
	}
	b, err := json.Marshal(j)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal JSONMap: %w", err)
	}
	return string(b), nil
}

// StringSlice handles PostgreSQL TEXT[] columns.
type StringSlice []string

func (s *StringSlice) Scan(src interface{}) error {
	if src == nil {
		*s = StringSlice{}
		return nil
	}
	switch v := src.(type) {
	case []byte:
		return s.parsePostgresArray(string(v))
	case string:
		return s.parsePostgresArray(v)
	default:
		return fmt.Errorf("unsupported type for StringSlice: %T", src)
	}
}

func (s *StringSlice) parsePostgresArray(str string) error {
	if str == "{}" || str == "" {
		*s = StringSlice{}
		return nil
	}
	// Remove surrounding braces
	str = str[1 : len(str)-1]
	// Simple split by comma (doesn't handle quoted strings with commas)
	result := []string{}
	for _, item := range splitPostgresArray(str) {
		if item != "" {
			// Remove surrounding quotes if present
			if len(item) >= 2 && item[0] == '"' && item[len(item)-1] == '"' {
				item = item[1 : len(item)-1]
			}
			result = append(result, item)
		}
	}
	*s = result
	return nil
}

func splitPostgresArray(s string) []string {
	var result []string
	var current []byte
	inQuotes := false
	for i := 0; i < len(s); i++ {
		switch {
		case s[i] == '"' && !inQuotes:
			inQuotes = true
			current = append(current, s[i])
		case s[i] == '"' && inQuotes:
			inQuotes = false
			current = append(current, s[i])
		case s[i] == ',' && !inQuotes:
			result = append(result, string(current))
			current = current[:0]
		default:
			current = append(current, s[i])
		}
	}
	if len(current) > 0 {
		result = append(result, string(current))
	}
	return result
}

type WorkspaceFilter struct {
	WorkspaceIDsIn *[]uint64
}

func (s Workspace) GetClusterName() string {
	return GetWorkspaceClusterName(s.ID)
}

func GetWorkspaceClusterName(id uint64) string {
	return fmt.Sprintf("workspace%v", id)
}

func GetIDFromClusterName(clusterName string) (uint64, error) {
	var id uint64
	_, err := fmt.Sscanf(clusterName, "workspace%d", &id)
	if err != nil {
		return 0, fmt.Errorf("unable to get workspace ID from cluster name %s: %w", clusterName, err)
	}
	return id, nil
}

func (s Workspace) GetContactUserID() uint64 {
	if s.ContactUserID != nil {
		return *s.ContactUserID
	}
	return 0
}

func (Workspace) IsValidSortType(sortType string) bool {
	validSortTypes := map[string]bool{
		"id":          true,
		"name":        true,
		"shortname":   true,
		"description": true,
		"status":      true,
		"isMain":      true,
		"createdat":   true,
	}

	return validSortTypes[sortType]
}
