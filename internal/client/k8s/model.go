package k8s

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WorkbenchAppState string

const (
	WorkbenchAppStateRunning WorkbenchAppState = "Running"
	WorkbenchAppStateStopped WorkbenchAppState = "Stopped"
	WorkbenchAppStateKilled  WorkbenchAppState = "Killed"
)

type WorkbenchServer struct {
	Version string `json:"version,omitempty"`
}
type Image struct {
	Registry   string `json:"registry"`
	Repository string `json:"repository"`
	Tag        string `json:"tag,omitempty"`
}
type KioskConfig struct {
	URL string `json:"url"`
}
type WorkbenchApp struct {
	Name        string                       `json:"name"`
	Version     string                       `json:"version,omitempty"`
	State       WorkbenchAppState            `json:"state,omitempty"`
	Image       *Image                       `json:"image,omitempty"`
	ShmSize     *resource.Quantity           `json:"shmSize,omitempty"`
	Resources   *corev1.ResourceRequirements `json:"resources,omitempty"`
	KioskConfig *KioskConfig                 `json:"kioskConfig,omitempty"`
}
type WorkbenchSpec struct {
	Server           WorkbenchServer `json:"server,omitempty"`
	Apps             []WorkbenchApp  `json:"apps,omitempty"`
	ServiceAccount   string          `json:"serviceAccountName,omitempty"`
	ImagePullSecrets []string        `json:"imagePullSecrets,omitempty"`
}

type WorkbenchStatusAppStatus string

const (
	WorkbenchStatusAppStatusUnknown     WorkbenchStatusAppStatus = "Unknown"
	WorkbenchStatusAppStatusRunning     WorkbenchStatusAppStatus = "Running"
	WorkbenchStatusAppStatusComplete    WorkbenchStatusAppStatus = "Complete"
	WorkbenchStatusAppStatusProgressing WorkbenchStatusAppStatus = "Progressing"
	WorkbenchStatusAppStatusFailed      WorkbenchStatusAppStatus = "Failed"
)

type WorkbenchStatusServerStatus string

const (
	WorkbenchStatusServerStatusRunning     WorkbenchStatusServerStatus = "Running"
	WorkbenchStatusServerStatusProgressing WorkbenchStatusServerStatus = "Progressing"
	WorkbenchStatusServerStatusFailed      WorkbenchStatusServerStatus = "Failed"
)

type WorkbenchStatusServer struct {
	Revision int                         `json:"revision"`
	Status   WorkbenchStatusServerStatus `json:"status"`
}

type WorkbenchStatusApp struct {
	Revision int                      `json:"revision"`
	Status   WorkbenchStatusAppStatus `json:"status"`
}

type WorkbenchStatus struct {
	Server WorkbenchStatusServer `json:"server"`
	Apps   []WorkbenchStatusApp  `json:"apps,omitempty"`
}

type Workbench struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              WorkbenchSpec   `json:"spec,omitempty"`
	Status            WorkbenchStatus `json:"status,omitempty"`
}

type WorkbenchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workbench `json:"items"`
}
