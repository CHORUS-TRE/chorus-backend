package k8s

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type K8sClienter interface {
	CreateWorkspace(namespace string) error
	DeleteWorkspace(namespace string) error
	CreateWorkbench(namespace, workbenchName string) error
	UpdateWorkbench(namespace, workbenchName string, apps []AppInstance) error
	CreatePortForward(namespace, serviceName string) (uint16, chan struct{}, error)
	CreateAppInstance(namespace, workbenchName string, app AppInstance) error
	DeleteAppInstance(namespace, workbenchName string, appInstance AppInstance) error
	DeleteWorkbench(namespace, workbenchName string) error
}

type client struct {
	cfg           config.Config
	restConfig    *rest.Config
	k8sClient     *kubernetes.Clientset
	dynamicClient *dynamic.DynamicClient
}

type AppInstance struct {
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

func appToApp(app AppInstance) WorkbenchApp {
	w := WorkbenchApp{
		Name: app.AppName,
	}

	if app.AppTag != "" {
		w.Version = app.AppTag
	}

	if app.AppRegistry != "" {
		if app.AppTag == "" {
			w.Image = &Image{
				Registry:   app.AppRegistry,
				Repository: app.AppImage,
			}
		} else {
			w.Image = &Image{
				Registry:   app.AppRegistry,
				Repository: app.AppImage,
				Tag:        app.AppTag,
			}
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
			w.Resources.Limits = corev1.ResourceList{
				"cpu": resource.MustParse(app.MaxCPU),
			}
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

func NewClient(cfg config.Config) (*client, error) {
	// clientcmd.SetLogger(&clientcmd.DefaultLogger{Verbosity: 10})

	restConfig, err := getK8sConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("error getting k8s config: %w", err)
	}

	k8sClient, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating k8s client: %w", err)
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating k8s client: %w", err)
	}

	c := &client{
		cfg:           cfg,
		restConfig:    restConfig,
		k8sClient:     k8sClient,
		dynamicClient: dynamicClient,
	}
	return c, nil
}

func (c *client) makeWorkbench(namespace, workbenchName string, apps []AppInstance) (Workbench, error) {
	workbench := Workbench{
		TypeMeta: v1.TypeMeta{
			Kind:       "Workbench",
			APIVersion: "default.chorus-tre.ch/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      workbenchName,
			Namespace: namespace,
		},
		Spec: WorkbenchSpec{
			Apps: []WorkbenchApp{},
		},
	}

	for _, app := range apps {
		workbenchApp := appToApp(app)
		workbench.Spec.Apps = append(workbench.Spec.Apps, workbenchApp)
	}

	if len(c.cfg.Clients.K8sClient.ImagePullSecrets) != 0 {
		workbench.Spec.ImagePullSecrets = []string{c.cfg.Clients.K8sClient.ImagePullSecretName}
	}

	if c.cfg.Clients.K8sClient.ServerVersion != "" {
		workbench.Spec.Server = WorkbenchServer{
			Version: c.cfg.Clients.K8sClient.ServerVersion,
		}
	}

	return workbench, nil
}

func EncodeRegistriesToDockerJSON(entries []config.ImagePullSecret) (string, error) {
	auths := make(map[string]map[string]string)

	for _, entry := range entries {
		auth := base64.StdEncoding.EncodeToString([]byte(entry.Username + ":" + entry.Password))

		auths[entry.Registry] = map[string]string{
			"auth": auth,
		}
	}

	result := map[string]map[string]map[string]string{
		"auths": auths,
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonData), nil
}

func (c *client) CreateWorkspace(namespace string) error {
	return c.syncNamespace(namespace)
}

func (c *client) DeleteWorkspace(namespace string) error {
	return c.deleteNamespace(namespace)
}

func (c *client) CreateWorkbench(namespace, workbenchName string) error {
	workbench, err := c.makeWorkbench(namespace, workbenchName, []AppInstance{})
	if err != nil {
		return fmt.Errorf("error creating workbench: %w", err)
	}

	return c.syncWorkbench(workbench, namespace)
}

func (c *client) UpdateWorkbench(namespace, workbenchName string, apps []AppInstance) error {
	workbench, err := c.makeWorkbench(namespace, workbenchName, apps)
	if err != nil {
		return fmt.Errorf("error creating workbench: %w", err)
	}

	return c.syncWorkbench(workbench, namespace)
}

func (c *client) CreateAppInstance(namespace, workbenchName string, appInstance AppInstance) error {
	app := appToApp(appInstance)

	gvr, err := c.getGroupVersionFromKind("Workbench")
	if err != nil {
		return fmt.Errorf("failed to get gvr from kind - %s", err)
	}

	patch := []map[string]interface{}{
		{
			"op":    "add",
			"path":  "/apps/-",
			"value": app,
		},
	}

	resource, err := c.dynamicClient.Resource(gvr).Namespace(namespace).Get(context.Background(), workbenchName, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to fetch workbench resource: %w", err)
	}

	// Check if the "apps" field exists
	unstructuredContent := resource.UnstructuredContent()
	spec, found, err := unstructured.NestedMap(unstructuredContent, "spec")
	if err != nil || !found {
		return fmt.Errorf("failed to retrieve spec: %w", err)
	}

	_, found = spec["apps"]
	if !found {
		fmt.Println("not found")
		patch = []map[string]interface{}{
			{
				"op":   "add",
				"path": "/spec/apps",
				// "value": []map[string]interface{}{app},
				"value": app,
			},
		}

	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("error marshalling patch: %w", err)
	}

	fmt.Println("dumping patchBytes", string(patchBytes))
	_, err = c.dynamicClient.Resource(gvr).Namespace(namespace).Patch(context.Background(), workbenchName, types.JSONPatchType, patchBytes, v1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("error applying patch: %w", err)
	}

	return nil
}

func (c *client) DeleteAppInstance(namespace, workbenchName string, appInstance AppInstance) error {
	app := appToApp(appInstance)

	patch := map[string]interface{}{
		"op":    "remove",
		"path":  "/apps/-",
		"value": app,
	}
	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("error marshalling patch: %w", err)
	}

	gvr, err := c.getGroupVersionFromKind("Workbench")
	if err != nil {
		return fmt.Errorf("failed to get gvr from kind - %s", err)
	}

	fmt.Println("dumping patchBytes")
	fmt.Println("patchBytes", string(patchBytes))

	_, err = c.dynamicClient.Resource(gvr).Namespace(namespace).Patch(context.Background(), workbenchName, types.JSONPatchType, patchBytes, v1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("error applying patch: %w", err)
	}

	return nil
}

func (c *client) DeleteWorkbench(namespace, workbenchName string) error {
	return c.deleteResource(namespace, "Workbench", workbenchName)
}
