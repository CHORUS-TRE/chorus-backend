package k8s

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

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
	Username                string
	UserID                  uint64
	Name                    string
	InitialResolutionWidth  uint32
	InitialResolutionHeight uint32
	Status                  string
	ServerPodStatus         string
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
		appInstance, exists := appsMap[k]
		if !exists {
			logger.TechLog.Warn(context.Background(), "workbench app in status not found in spec apps", zap.String("appUid", k), zap.String("workbenchName", wb.Name))
			continue
		}

		appInstance.K8sStatus = string(app.Status)
	}

	for _, app := range appsMap {
		apps = append(apps, *app)
	}

	tenantIDStr := wb.Labels["chorus-tre.ch/tenant-id"]
	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 64)
	if err != nil {
		return Workbench{}, fmt.Errorf("error parsing tenant ID: %w", err)
	}

	var serverPodStatus string
	if wb.Status.ServerDeployment.ServerPod != nil {
		serverPodStatus = string(wb.Status.ServerDeployment.ServerPod.Status)
	} else {
		serverPodStatus = string(WorkbenchServerContainerStatusUnknown)
	}

	workbench := Workbench{
		Namespace:               wb.Namespace,
		TenantID:                tenantID,
		Name:                    wb.Name,
		Username:                wb.Spec.Server.User,
		UserID:                  c.K8sUserIDToUserID(uint64(wb.Spec.Server.UserID)),
		InitialResolutionWidth:  uint32(wb.Spec.Server.InitialResolutionWidth),
		InitialResolutionHeight: uint32(wb.Spec.Server.InitialResolutionHeight),
		Status:                  string(wb.Status.ServerDeployment.Status),
		ServerPodStatus:         serverPodStatus,
		Apps:                    apps,
	}

	return workbench, nil
}

const userIDOffset uint64 = 1001

func (c *client) UsernameToK8sUser(username string) string {
	name := strings.ToLower(username)
	name = strings.ReplaceAll(name, " ", "_")
	reg := regexp.MustCompile(`[^a-z0-9_]`)
	name = reg.ReplaceAllString(name, "")

	return name
}

func (c *client) K8sUserIDToUserID(userID uint64) uint64 {
	return userID - userIDOffset
}

func (c *client) UserIDToK8sUserID(userID uint64) uint64 {
	return userID + userIDOffset
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
	KioskConfigJWTURL   string
	KioskConfigJWTToken string
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

func (a AppInstance) SanitizedAppName() string {
	name := strings.ToLower(a.AppName)
	re := regexp.MustCompile("[^a-z0-9]+")
	name = re.ReplaceAllString(name, "-")
	if len(name) > 15 {
		name = name[:15]
	}
	name = strings.Trim(name, "-")

	if name == "" {
		name = "unknown"
	}

	return name
}

func (c *client) appInstanceToWorkbenchApp(app AppInstance) WorkbenchApp {
	w := WorkbenchApp{
		Name: fmt.Sprintf("%s-%v", app.SanitizedAppName(), app.ID),
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
	if app.KioskConfigJWTURL != "" {
		if w.KioskConfig == nil {
			w.KioskConfig = &KioskConfig{}
		}
		w.KioskConfig.JWTURL = app.KioskConfigJWTURL
	}
	if app.KioskConfigJWTToken != "" {
		if w.KioskConfig == nil {
			w.KioskConfig = &KioskConfig{}
		}
		w.KioskConfig.JWTToken = app.KioskConfigJWTToken
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
	idStr := w.Name[strings.LastIndex((w.Name), "-")+1:]
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
	ServerDeployment WorkbenchStatusServer         `json:"serverDeployment"`
	Apps             map[string]WorkbenchStatusApp `json:"apps,omitempty"`
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
