package k8s

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

const appInstanceNamePrefix = "app-instance-"

type AppInstance struct {
	ID      uint64
	AppName string

	AppRegistry string
	AppImage    string
	AppTag      string

	ShmSize        string
	KioskConfigURL string
	MaxCPU         string
	MinCPU         string
	MaxMemory      string
	MinMemory      string
	// IconURL        string
}

func (c *client) appToApp(app AppInstance) WorkbenchApp {
	w := WorkbenchApp{
		Name: fmt.Sprintf("%s%v", appInstanceNamePrefix, app.ID),
	}

	if app.AppTag != "" {
		w.Version = app.AppTag
	}

	if app.AppRegistry == "" {
		w.Image = &Image{
			Registry:   c.cfg.Clients.K8sClient.DefaultRegistry,
			Repository: c.cfg.Clients.K8sClient.DefaultRepository + "/" + app.AppImage,
			Tag:        app.AppTag,
		}
	} else {
		w.Image = &Image{
			Registry:   app.AppRegistry,
			Repository: app.AppImage,
		}
		if app.AppTag == "" {
			w.Image.Tag = "latest"
		} else {
			w.Image.Tag = app.AppTag
		}
	}

	if app.ShmSize != "" {
		shmSize := resource.MustParse(app.ShmSize)
		w.ShmSize = &shmSize
	}
	if app.KioskConfigURL != "" {
		w.KioskConfig = &KioskConfig{
			URL: app.KioskConfigURL,
		}
	}

	if app.MaxCPU != "" || app.MinCPU != "" || app.MaxMemory != "" || app.MinMemory != "" {
		w.Resources = &corev1.ResourceRequirements{}
		if app.MaxCPU != "" {
			if w.Resources.Limits == nil {
				w.Resources.Limits = corev1.ResourceList{}
			}
			w.Resources.Limits["cpu"] = resource.MustParse(app.MaxCPU)
		}
		if app.MinCPU != "" {
			if w.Resources.Requests == nil {
				w.Resources.Requests = corev1.ResourceList{}
			}
			w.Resources.Requests["cpu"] = resource.MustParse(app.MinCPU)
		}
		if app.MaxMemory != "" {
			if w.Resources.Limits == nil {
				w.Resources.Limits = corev1.ResourceList{}
			}
			w.Resources.Limits["memory"] = resource.MustParse(app.MaxMemory)
		}
		if app.MinMemory != "" {
			if w.Resources.Requests == nil {
				w.Resources.Requests = corev1.ResourceList{}
			}
			w.Resources.Requests["memory"] = resource.MustParse(app.MinMemory)
		}
	}

	return w
}

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

type Namespace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              corev1.NamespaceSpec   `json:"spec,omitempty"`
	Status            corev1.NamespaceStatus `json:"status,omitempty"`
}

func EventInterfaceToWorkbench(a any) (*Workbench, error) {
	u, ok := a.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("expected unstructured.Unstructured, got %T", a)
	}
	return UnstructuredToWorkbench(u)
}

func UnstructuredToWorkbench(u *unstructured.Unstructured) (*Workbench, error) {
	var wb Workbench
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &wb)
	if err != nil {
		return nil, err
	}
	return &wb, nil
}

func EventInterfaceToNamespace(a any) (*Namespace, error) {
	u, ok := a.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("expected unstructured.Unstructured, got %T", a)
	}
	return UnstructuredToNamespace(u)
}

func UnstructuredToNamespace(u *unstructured.Unstructured) (*Namespace, error) {
	var ns Namespace
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &ns)
	if err != nil {
		return nil, err
	}
	return &ns, nil
}
