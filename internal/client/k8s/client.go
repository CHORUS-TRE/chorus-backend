package k8s

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	helmchart "helm.sh/helm/v3/pkg/chart"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type K8sClienter interface {
	CreateWorkbench(namespace, workbenchName string) error
	UpdateWorkbench(namespace, workbenchName string, apps []AppInstance) error
	CreatePortForward(namespace, serviceName string) (uint16, chan struct{}, error)
	CreateAppInstance(namespace, workbenchName string, app AppInstance) error
	DeleteAppInstance(namespace, workbenchName string, appInstance AppInstance) error
	DeleteWorkbench(namespace, workbenchName string) error
}

type client struct {
	cfg           config.Config
	chart         *helmchart.Chart
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

func appToMap(app AppInstance) map[string]interface{} {
	m := map[string]interface{}{
		"app":  app.AppName,
		"name": app.AppName,
	}
	if app.AppTag != "" {
		m["version"] = app.AppTag
	}

	if app.AppRegistry != "" {
		if app.AppTag == "" {
			m["image"] = map[string]string{
				"registry":   app.AppRegistry,
				"repository": app.AppImage,
			}
		} else {
			m["image"] = map[string]string{
				"registry":   app.AppRegistry,
				"repository": app.AppImage,
				"tag":        app.AppTag,
			}
		}
	}

	if app.ShmSize != "" {
		m["shmSize"] = app.ShmSize
	}
	if app.KioskConfigURL != "" {
		m["kioskConfig"] = map[string]string{
			"url": app.KioskConfigURL,
		}
	}
	if app.MaxCPU != "" || app.MinCPU != "" || app.MaxMemory != "" || app.MinMemory != "" {
		m["resources"] = map[string]map[string]string{}
		if app.MaxCPU != "" {
			m["resources"].(map[string]map[string]string)["limits"] = map[string]string{
				"cpu": app.MaxCPU,
			}
		}
		if app.MinCPU != "" {
			m["resources"].(map[string]map[string]string)["requests"] = map[string]string{
				"cpu": app.MinCPU,
			}
		}
		if app.MaxMemory != "" {
			if _, ok := m["resources"]; !ok {
				m["resources"] = map[string]map[string]string{}
			}
			m["resources"].(map[string]map[string]string)["limits"] = map[string]string{
				"memory": app.MaxMemory,
			}
		}
		if app.MinMemory != "" {
			if _, ok := m["resources"]; !ok {
				m["resources"] = map[string]map[string]string{}
			}
			m["resources"].(map[string]map[string]string)["requests"] = map[string]string{
				"memory": app.MinMemory,
			}
		}
	}

	return m
}

func NewClient(cfg config.Config) (*client, error) {
	// clientcmd.SetLogger(&clientcmd.DefaultLogger{Verbosity: 10})

	chart, err := GetHelmChart()
	if err != nil {
		return nil, fmt.Errorf("error loading Helm chart: %w", err)
	}

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
		chart:         chart,
		cfg:           cfg,
		restConfig:    restConfig,
		k8sClient:     k8sClient,
		dynamicClient: dynamicClient,
	}
	return c, nil
}

func (c *client) renderWorkbenchTemplate(namespace, workbenchName string, apps []AppInstance) (string, error) {
	appMaps := []map[string]interface{}{}
	for _, app := range apps {
		appMaps = append(appMaps, appToMap(app))
	}

	vals := map[string]interface{}{
		"namespace": namespace,
		"name":      workbenchName,
		"apps":      appMaps,
	}
	if len(c.cfg.Clients.K8sClient.ImagePullSecrets) != 0 {
		dockerConfig, err := EncodeRegistriesToDockerJSON(c.cfg.Clients.K8sClient.ImagePullSecrets)
		if err != nil {
			return "", fmt.Errorf("unable to encode registries: %w", err)
		}
		vals["imagePullSecret"] = map[string]string{
			"name":             "image-pull-secret",
			"dockerConfigJson": dockerConfig,
		}
	}

	return c.renderTemplate(namespace, workbenchName, vals)
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

func (c *client) CreateWorkbench(namespace, workbenchName string) error {
	manifest, err := c.renderWorkbenchTemplate(namespace, workbenchName, []AppInstance{})
	if err != nil {
		return fmt.Errorf("error rendering template: %w", err)
	}
	err = c.applyManifest(manifest, namespace)
	if err != nil {
		return fmt.Errorf("error applying manifest: %w", err)
	}

	return nil
}

func (c *client) UpdateWorkbench(namespace, workbenchName string, apps []AppInstance) error {
	manifest, err := c.renderWorkbenchTemplate(namespace, workbenchName, apps)
	if err != nil {
		return fmt.Errorf("error rendering template: %w", err)
	}
	err = c.applyManifest(manifest, namespace)
	if err != nil {
		return fmt.Errorf("error applying manifest: %w", err)
	}

	return nil
}

func (c *client) CreateAppInstance(namespace, workbenchName string, appInstance AppInstance) error {
	app := appToMap(appInstance)

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
				"op":    "add",
				"path":  "/spec/apps",
				"value": []map[string]interface{}{app},
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
	app := appToMap(appInstance)

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
	gvr, err := c.getGroupVersionFromKind("Workbench")
	if err != nil {
		return fmt.Errorf("failed to get gvr from kind - %s", err)
	}

	err = c.dynamicClient.Resource(gvr).Namespace(namespace).Delete(context.Background(), workbenchName, v1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("error deleting workbench: %w", err)
	}

	return nil
}
