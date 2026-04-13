package model

import "time"

// WorkspaceSvc maps an entry in the 'workspace_services' database table.
type WorkspaceSvc struct {
	ID          uint64
	TenantID    uint64
	WorkspaceID uint64
	Name        string

	// Spec (desired state)
	State                  string
	ChartRegistry          string
	ChartRepository        string
	ChartTag               string
	Values                 JSONMap[any] `db:"valuesoverride"`
	CredentialsSecretName  string
	CredentialsPaths       StringSlice
	ConnectionInfoTemplate string
	ComputedValues         JSONMap[string]

	// Status (observed from K8s)
	Status         string
	StatusMessage  string
	ConnectionInfo string
	SecretName     string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type WorkspaceSvcFilter struct {
	WorkspaceIDsIn *[]uint64
}

type WorkspaceSvcStatusUpdate struct {
	Status         string
	StatusMessage  string
	ConnectionInfo string
	SecretName     string
}

func (WorkspaceSvc) IsValidSortType(sortType string) bool {
	validSortTypes := map[string]bool{
		"id":        true,
		"name":      true,
		"state":     true,
		"status":    true,
		"createdat": true,
	}
	return validSortTypes[sortType]
}
