package model

import "time"

// ServiceInstanceState defines the desired lifecycle state of a workspace service instance.
type ServiceInstanceState string

const (
	ServiceInstanceStateRunning ServiceInstanceState = "Running"
	ServiceInstanceStateStopped ServiceInstanceState = "Stopped"
	ServiceInstanceStateDeleted ServiceInstanceState = "Deleted"
)

func (s ServiceInstanceState) String() string {
	return string(s)
}

// ServiceInstanceStatus defines the observed status of a workspace service instance.
type ServiceInstanceStatus string

const (
	ServiceInstanceStatusProgressing ServiceInstanceStatus = "Progressing"
	ServiceInstanceStatusRunning     ServiceInstanceStatus = "Running"
	ServiceInstanceStatusStopped     ServiceInstanceStatus = "Stopped"
	ServiceInstanceStatusDeleted     ServiceInstanceStatus = "Deleted"
	ServiceInstanceStatusFailed      ServiceInstanceStatus = "Failed"
)

func (s ServiceInstanceStatus) String() string {
	return string(s)
}

// WorkspaceServiceInstance maps an entry in the 'workspace_services' database table.
type WorkspaceServiceInstance struct {
	ID          uint64
	TenantID    uint64
	WorkspaceID uint64
	Name        string

	// Spec (desired state)
	State                  ServiceInstanceState
	ChartRegistry          string
	ChartRepository        string
	ChartTag               string
	Values                 JSONMap[any] `db:"valuesoverride"`
	CredentialsSecretName  string
	CredentialsPaths       StringSlice
	ConnectionInfoTemplate string
	ComputedValues         JSONMap[string]

	// Status (observed from K8s)
	Status         ServiceInstanceStatus
	StatusMessage  string
	ConnectionInfo string
	SecretName     string

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type WorkspaceServiceInstanceFilter struct {
	WorkspaceIDsIn *[]uint64
}

type WorkspaceServiceInstanceStatusUpdate struct {
	Status         string
	StatusMessage  string
	ConnectionInfo string
	SecretName     string
}

func (WorkspaceServiceInstance) IsValidSortType(sortType string) bool {
	validSortTypes := map[string]bool{
		"id":        true,
		"name":      true,
		"state":     true,
		"status":    true,
		"createdat": true,
	}
	return validSortTypes[sortType]
}
