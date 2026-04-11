package k8s

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ----------------------------------------------------------------
// Workspace CRD K8s types
// Mirrors the operator's Workspace CRD (default.chorus-tre.ch/v1alpha1)
// ----------------------------------------------------------------

// WorkspaceServiceState defines the desired lifecycle state
type WorkspaceServiceState string

const (
	WorkspaceServiceStateRunning WorkspaceServiceState = "Running"
	WorkspaceServiceStateStopped WorkspaceServiceState = "Stopped"
	WorkspaceServiceStateDeleted WorkspaceServiceState = "Deleted"
)

// WorkspaceServiceChart identifies a Helm chart in an OCI registry
type WorkspaceServiceChart struct {
	Registry   string `json:"registry,omitempty"`
	Repository string `json:"repository,omitempty"`
	Tag        string `json:"tag"`
}

// WorkspaceServiceCredentials configures auto-generated password injection
type WorkspaceServiceCredentials struct {
	SecretName string   `json:"secretName"`
	Paths      []string `json:"paths,omitempty"`
}

// WorkspaceK8sService defines a Helm-chart-based service in the workspace namespace
type WorkspaceK8sService struct {
	State                  WorkspaceServiceState        `json:"state,omitempty"`
	Chart                  WorkspaceServiceChart        `json:"chart"`
	Values                 *apiextensionsv1.JSON        `json:"values,omitempty"`
	Credentials            *WorkspaceServiceCredentials `json:"credentials,omitempty"`
	ConnectionInfoTemplate string                       `json:"connectionInfoTemplate,omitempty"`
	ComputedValues         map[string]string            `json:"computedValues,omitempty"`
}

// WorkspaceSpec defines the desired state of the Workspace CRD
type WorkspaceSpec struct {
	NetworkPolicy string                         `json:"networkPolicy"`
	AllowedFQDNs  []string                       `json:"allowedFQDNs,omitempty"`
	Services      map[string]WorkspaceK8sService `json:"services,omitempty"`
}

// NetworkPolicyStatus is the observed state of the workspace network policy
type NetworkPolicyStatus struct {
	Status  string `json:"status,omitempty"`
	Message string `json:"message,omitempty"`
}

// WorkspaceStatusServiceStatus is the observed state of a workspace service
type WorkspaceStatusServiceStatus string

const (
	WorkspaceStatusServiceProgressing WorkspaceStatusServiceStatus = "Progressing"
	WorkspaceStatusServiceRunning     WorkspaceStatusServiceStatus = "Running"
	WorkspaceStatusServiceStopped     WorkspaceStatusServiceStatus = "Stopped"
	WorkspaceStatusServiceDeleted     WorkspaceStatusServiceStatus = "Deleted"
	WorkspaceStatusServiceFailed      WorkspaceStatusServiceStatus = "Failed"
)

// WorkspaceStatusService is the observed status of a workspace service
type WorkspaceStatusService struct {
	Status         WorkspaceStatusServiceStatus `json:"status"`
	Message        string                       `json:"message,omitempty"`
	ConnectionInfo string                       `json:"connectionInfo,omitempty"`
	SecretName     string                       `json:"secretName,omitempty"`
}

// WorkspaceStatus defines the observed state of the Workspace CRD
type K8sWorkspaceStatus struct {
	ObservedGeneration int64                             `json:"observedGeneration,omitempty"`
	NetworkPolicy      NetworkPolicyStatus               `json:"networkPolicy,omitempty"`
	Conditions         []metav1.Condition                `json:"conditions,omitempty"`
	Services           map[string]WorkspaceStatusService `json:"services,omitempty"`
}

// K8sWorkspace is the full Workspace CRD object
type K8sWorkspace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              WorkspaceSpec      `json:"spec,omitempty"`
	Status            K8sWorkspaceStatus `json:"status,omitempty"`
}

// K8sWorkspaceList contains a list of Workspaces
type K8sWorkspaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []K8sWorkspace `json:"items"`
}

// ----------------------------------------------------------------
// Internal workspace input/output for K8s client
// ----------------------------------------------------------------

// WorkspaceInput is the internal model used to create/update a Workspace CRD
type WorkspaceInput struct {
	TenantID      uint64
	Namespace     string
	NetworkPolicy string
	AllowedFQDNs  []string
	Clipboard     string
	Services      map[string]WorkspaceK8sService
}

// WorkspaceOutput is the internal model received from Workspace CRD watch events
type WorkspaceOutput struct {
	CurrentGeneration  int64
	ObservedGeneration int64
	Namespace          string
	TenantID           uint64

	NetworkPolicyStatus  string
	NetworkPolicyMessage string

	ServiceStatuses map[string]WorkspaceServiceStatusOutput
}

// WorkspaceServiceStatusOutput is the observed status of a workspace service
type WorkspaceServiceStatusOutput struct {
	Status         string
	Message        string
	ConnectionInfo string
	SecretName     string
}
