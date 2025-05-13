package k8s

import (
	"context"
	"fmt"
	"strconv"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type Workbench struct {
	Namespace               string
	TenantID                uint64
	WorkbenchName           string
	InitialResolutionWidth  uint32
	InitialResolutionHeight uint32
	Status                  string
	Apps                    []AppInstance
}

func (c *client) K8sWorkbenchToWorkbench(wb K8sWorkbench) (Workbench, error) {
	apps := make([]AppInstance, 0, len(wb.Spec.Apps))
	appsMap := make(map[string]*AppInstance, len(wb.Spec.Apps))
	for k, app := range wb.Spec.Apps {
		appInstance, err := c.workbenchAppToAppInstance(app)
		if err != nil {
			return Workbench{}, fmt.Errorf("error converting to AppInstance: %w", err)
		}
		appsMap[k] = &appInstance
	}

	for k, app := range wb.Status.Apps {
		appsMap[k].K8sStatus = string(app.Status)
	}

	for _, app := range appsMap {
		apps = append(apps, *app)
	}

	tenantIDStr := wb.Labels["chorus-tre.ch/tenant-id"]
	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 64)
	if err != nil {
		return Workbench{}, fmt.Errorf("error parsing tenant ID: %w", err)
	}

	workbench := Workbench{
		TenantID:                tenantID,
		Namespace:               wb.Namespace,
		WorkbenchName:           wb.Name,
		InitialResolutionWidth:  uint32(wb.Spec.Server.InitialResolutionWidth),
		InitialResolutionHeight: uint32(wb.Spec.Server.InitialResolutionHeight),
		Status:                  string(wb.Status.Server.Status),
		Apps:                    apps,
	}

	return workbench, nil
}

const appInstanceNamePrefix = "app-instance-"

type AppInstance struct {
	ID      uint64
	AppName string

	AppRegistry string
	AppImage    string
	AppTag      string

	K8sState  string
	K8sStatus string

	ShmSize             string
	KioskConfigURL      string
	MaxCPU              string
	MinCPU              string
	MaxMemory           string
	MinMemory           string
	MaxEphemeralStorage string
	MinEphemeralStorage string
	// IconURL        string
}

func (a AppInstance) UID() string {
	return fmt.Sprintf("%s%v", appInstanceNamePrefix, a.ID)
}

func (c *client) appInstanceToWorkbenchApp(app AppInstance) WorkbenchApp {
	w := WorkbenchApp{
		Name: app.UID(),
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

	if app.MaxCPU != "" || app.MinCPU != "" || app.MaxMemory != "" || app.MinMemory != "" || app.MaxEphemeralStorage != "" || app.MinEphemeralStorage != "" {
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
		if app.MaxEphemeralStorage != "" {
			if w.Resources.Limits == nil {
				w.Resources.Limits = corev1.ResourceList{}
			}
			w.Resources.Limits["ephemeral-storage"] = resource.MustParse(app.MaxEphemeralStorage)
		}
		if app.MinEphemeralStorage != "" {
			if w.Resources.Requests == nil {
				w.Resources.Requests = corev1.ResourceList{}
			}
			w.Resources.Requests["ephemeral-storage"] = resource.MustParse(app.MinEphemeralStorage)
		}
	}

	return w
}

func (c *client) workbenchAppToAppInstance(w WorkbenchApp) (AppInstance, error) {
	idStr := w.Name[len(appInstanceNamePrefix):]
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		logger.TechLog.Error(context.Background(), "failed to parse app instance ID", zap.Any("workbenchApp", w), zap.Error(err))
		err = fmt.Errorf("failed to parse app instance ID %s: %w", idStr, err)
		return AppInstance{}, err
	}

	app := AppInstance{
		ID:      id,
		AppName: w.Name,
	}

	if w.Image != nil {
		app.AppRegistry = w.Image.Registry
		app.AppImage = w.Image.Repository
		app.AppTag = w.Image.Tag
	}

	app.K8sState = string(w.State)

	if w.ShmSize != nil {
		app.ShmSize = w.ShmSize.String()
	}
	if w.KioskConfig != nil {
		app.KioskConfigURL = w.KioskConfig.URL
	}

	if w.Resources != nil {
		if w.Resources.Limits != nil {
			if cpu, ok := w.Resources.Limits["cpu"]; ok {
				app.MaxCPU = cpu.String()
			}
			if memory, ok := w.Resources.Limits["memory"]; ok {
				app.MaxMemory = memory.String()
			}
			if ephemeralStorage, ok := w.Resources.Limits["ephemeral-storage"]; ok {
				app.MaxEphemeralStorage = ephemeralStorage.String()
			}
		}
		if w.Resources.Requests != nil {
			if cpu, ok := w.Resources.Requests["cpu"]; ok {
				app.MinCPU = cpu.String()
			}
			if memory, ok := w.Resources.Requests["memory"]; ok {
				app.MinMemory = memory.String()
			}
			if ephemeralStorage, ok := w.Resources.Requests["ephemeral-storage"]; ok {
				app.MinEphemeralStorage = ephemeralStorage.String()
			}
		}
	}

	return app, nil
}

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
	Server           WorkbenchServer         `json:"server,omitempty"`
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

type WorkbenchStatusServer struct {
	Revision int                         `json:"revision"`
	Status   WorkbenchStatusServerStatus `json:"status"`
}

type WorkbenchStatusApp struct {
	Revision int                      `json:"revision"`
	Status   WorkbenchStatusAppStatus `json:"status"`
}

type WorkbenchStatus struct {
	Server WorkbenchStatusServer         `json:"server"`
	Apps   map[string]WorkbenchStatusApp `json:"apps,omitempty"`
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

func EventInterfaceToWorkbench(a any) (*K8sWorkbench, error) {
	u, ok := a.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("expected unstructured.Unstructured, got %T", a)
	}
	return UnstructuredToWorkbench(u)
}

func UnstructuredToWorkbench(u *unstructured.Unstructured) (*K8sWorkbench, error) {
	var wb K8sWorkbench
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
