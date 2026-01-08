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
	InitialResolutionWidth  int    `json:"initialResolutionWidth,omitempty"`
	InitialResolutionHeight int    `json:"initialResolutionHeight,omitempty"`
	Version                 string `json:"version,omitempty"`
	User                    string `json:"user,omitempty"`
	UserID                  int    `json:"userid,omitempty"`
}
type Image struct {
	Registry   string `json:"registry"`
	Repository string `json:"repository"`
	Tag        string `json:"tag,omitempty"`
}
type KioskConfig struct {
	URL      string `json:"url"`
	JWTURL   string `json:"jwtUrl,omitempty"`
	JWTToken string `json:"jwtToken,omitempty"`
}
type InitContainerConfig struct {
	Version string `json:"version,omitempty"`
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
	Server           WorkbenchServer         `json:"server,omitempty"`
	InitContainer    *InitContainerConfig    `json:"initContainer,omitempty"`
	Apps             map[string]WorkbenchApp `json:"apps,omitempty"`
	ServiceAccount   string                  `json:"serviceAccountName,omitempty"`
	ImagePullSecrets []string                `json:"imagePullSecrets,omitempty"`
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

type WorkbenchServerContainerStatus string

const (
	WorkbenchServerContainerStatusWaiting     WorkbenchServerContainerStatus = "Waiting"
	WorkbenchServerContainerStatusStarting    WorkbenchServerContainerStatus = "Starting"
	WorkbenchServerContainerStatusReady       WorkbenchServerContainerStatus = "Ready"
	WorkbenchServerContainerStatusFailing     WorkbenchServerContainerStatus = "Failing"
	WorkbenchServerContainerStatusRestarting  WorkbenchServerContainerStatus = "Restarting"
	WorkbenchServerContainerStatusTerminating WorkbenchServerContainerStatus = "Terminating"
	WorkbenchServerContainerStatusTerminated  WorkbenchServerContainerStatus = "Terminated"
	WorkbenchServerContainerStatusUnknown     WorkbenchServerContainerStatus = "Unknown"
)

type WorkbenchServerPodHealth struct {
	Status       WorkbenchServerContainerStatus `json:"status"`
	Ready        bool                           `json:"ready"`
	RestartCount int32                          `json:"restartCount"`
	Message      string                         `json:"message,omitempty"`
}

type WorkbenchStatusServer struct {
	Revision  int                         `json:"revision"`
	Status    WorkbenchStatusServerStatus `json:"status"`
	ServerPod *WorkbenchServerPodHealth   `json:"serverPod,omitempty"`
}

type WorkbenchStatusApp struct {
	Revision int                      `json:"revision"`
	Status   WorkbenchStatusAppStatus `json:"status"`
}

type WorkbenchStatus struct {
	ObservedGeneration int64                         `json:"observedGeneration,omitempty"`
	ServerDeployment   WorkbenchStatusServer         `json:"serverDeployment"`
	Apps               map[string]WorkbenchStatusApp `json:"apps,omitempty"`
}

type K8sWorkbench struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              WorkbenchSpec   `json:"spec,omitempty"`
	Status            WorkbenchStatus `json:"status,omitempty"`
}

type WorkbenchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []K8sWorkbench `json:"items"`
}

type Namespace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              corev1.NamespaceSpec   `json:"spec,omitempty"`
	Status            corev1.NamespaceStatus `json:"status,omitempty"`
}
