package k8s

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ----------------------------------------------------------------
// Converters from internal types to K8s types
// ----------------------------------------------------------------

func (c *client) workbenchToK8sWorkbench(workbench *Workbench) (K8sWorkbench, error) {
	// Construct K8s Workbench
	k8sWorkbench := K8sWorkbench{
		TypeMeta: v1.TypeMeta{
			Kind:       "Workbench",
			APIVersion: "default.chorus-tre.ch/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      workbench.Name,
			Namespace: workbench.Namespace,
			Labels: map[string]string{
				"chorus-tre.ch/created-by": "chorus-backend",
				"chorus-tre.ch/tenant-id":  fmt.Sprintf("%d", workbench.TenantID),
			},
		},
		Spec: WorkbenchSpec{
			Server: WorkbenchServer{
				InitialResolutionWidth:  int(workbench.InitialResolutionWidth),
				InitialResolutionHeight: int(workbench.InitialResolutionHeight),
			},
			Apps: map[string]WorkbenchApp{},
		},
	}

	// Convert Workbench Apps
	for _, app := range workbench.Apps {
		workbenchApp := c.appInstanceToK8sWorkbenchApp(app)
		k8sWorkbench.Spec.Apps[app.UID()] = workbenchApp
	}

	// Add optional fields from config
	if len(c.cfg.Clients.K8sClient.ImagePullSecrets) != 0 {
		k8sWorkbench.Spec.ImagePullSecrets = []string{c.cfg.Clients.K8sClient.ImagePullSecretName}
	}

	if c.cfg.Clients.K8sClient.ServerVersion != "" {
		k8sWorkbench.Spec.Server.Version = c.cfg.Clients.K8sClient.ServerVersion
	}

	if c.cfg.Clients.K8sClient.InitContainerVersion != "" {
		k8sWorkbench.Spec.InitContainer = &InitContainerConfig{
			Version: c.cfg.Clients.K8sClient.InitContainerVersion,
		}
	}

	// Add user details if configured and not empty
	username := workbench.SanitizedUsername()
	if c.cfg.Clients.K8sClient.AddUserDetails && username != "" {
		k8sWorkbench.Spec.Server.User = username
		k8sWorkbench.Spec.Server.UserID = int(workbench.UserID + workbenchUserIDOffset)
	}

	return k8sWorkbench, nil
}

func (c *client) appInstanceToK8sWorkbenchApp(app AppInstance) WorkbenchApp {
	w := WorkbenchApp{
		Name: fmt.Sprintf("%s-%v", app.SanitizedAppName(), app.ID),
	}

	if app.AppTag != "" {
		w.Version = app.AppTag
	}

	if app.K8sState != "" {
		w.State = WorkbenchAppState(app.K8sState)
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
			w.Image.Tag = defaultImageTag
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

// ----------------------------------------------------------------
// Converters from K8s types to internal types
// ----------------------------------------------------------------

func (c *client) k8sWorkbenchToWorkbench(wb K8sWorkbench) (Workbench, error) {
	// Convert Workbench Apps
	apps := make([]AppInstance, 0, len(wb.Spec.Apps))
	appsMap := make(map[string]*AppInstance, len(wb.Spec.Apps))
	for k, app := range wb.Spec.Apps {
		appInstance, err := c.k8sWorkbenchAppToAppInstance(app)
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

	// Parse tenant ID
	tenantIDStr := wb.Labels["chorus-tre.ch/tenant-id"]
	tenantID, err := strconv.ParseUint(tenantIDStr, 10, 64)
	if err != nil {
		return Workbench{}, fmt.Errorf("error parsing tenant ID: %w", err)
	}

	// Get server pod status
	var serverPodStatus string
	if wb.Status.ServerDeployment.ServerPod != nil {
		serverPodStatus = string(wb.Status.ServerDeployment.ServerPod.Status)
	} else {
		serverPodStatus = string(WorkbenchServerContainerStatusUnknown)
	}

	// Construct Workbench
	workbench := Workbench{
		Namespace:               wb.Namespace,
		TenantID:                tenantID,
		Name:                    wb.Name,
		Username:                wb.Spec.Server.User,
		UserID:                  uint64(wb.Spec.Server.UserID) - workbenchUserIDOffset,
		InitialResolutionWidth:  uint32(wb.Spec.Server.InitialResolutionWidth),
		InitialResolutionHeight: uint32(wb.Spec.Server.InitialResolutionHeight),
		Status:                  string(wb.Status.ServerDeployment.Status),
		ServerPodStatus:         serverPodStatus,
		Apps:                    apps,
	}

	return workbench, nil
}

func (c *client) k8sWorkbenchAppToAppInstance(w WorkbenchApp) (AppInstance, error) {
	// Parse app instance ID from name
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

// ----------------------------------------------------------------
// Converters from event interface to internal types
// ----------------------------------------------------------------

// Converts event object to Workbench
func (c *client) eventInterfaceToWorkbench(obj any) (Workbench, error) {
	if obj == nil {
		return Workbench{}, fmt.Errorf("nil object received in event")
	}

	k8sWorkbench, err := eventInterfaceToK8sWorkbench(obj)
	if err != nil {
		return Workbench{}, fmt.Errorf("error converting to Workbench: %w", err)
	}

	workbench, err := c.k8sWorkbenchToWorkbench(*k8sWorkbench)
	if err != nil {
		return Workbench{}, fmt.Errorf("error converting to Workbench: %w", err)
	}
	return workbench, nil
}

// ----------------------------------------------------------------
// Converters from unstructured to K8s types
// ----------------------------------------------------------------

func eventInterfaceToK8sWorkbench(a any) (*K8sWorkbench, error) {
	u, ok := a.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("expected unstructured.Unstructured, got %T", a)
	}
	return unstructuredToK8sWorkbench(u)
}

func unstructuredToK8sWorkbench(u *unstructured.Unstructured) (*K8sWorkbench, error) {
	var wb K8sWorkbench
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &wb)
	if err != nil {
		return nil, err
	}
	return &wb, nil
}

// func eventInterfaceToNamespace(a any) (*Namespace, error) {
// 	u, ok := a.(*unstructured.Unstructured)
// 	if !ok {
// 		return nil, fmt.Errorf("expected unstructured.Unstructured, got %T", a)
// 	}
// 	return unstructuredToNamespace(u)
// }

// func unstructuredToNamespace(u *unstructured.Unstructured) (*Namespace, error) {
// 	var ns Namespace
// 	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &ns)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &ns, nil
// }
