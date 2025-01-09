package helm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"go.uber.org/zap"
	helmaction "helm.sh/helm/v3/pkg/action"
	helmchart "helm.sh/helm/v3/pkg/chart"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/jsonmergepatch"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type HelmClienter interface {
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
	AppName     string
	AppRegistry string
	AppImage    string
	AppVersion  string
}

func appToMap(app AppInstance) map[string]interface{} {
	m := map[string]interface{}{
		"app":  app.AppName,
		"name": app.AppName,
	}
	if app.AppVersion != "" {
		m["version"] = app.AppVersion
	}

	if app.AppRegistry != "" {
		if app.AppVersion == "" {
			m["image"] = map[string]string{
				"registry":   app.AppRegistry,
				"repository": app.AppImage,
			}
		} else {
			m["image"] = map[string]string{
				"registry":   app.AppRegistry,
				"repository": app.AppImage,
				"tag":        app.AppVersion,
			}
		}
	}

	return m
}

func debug(format string, v ...interface{}) {
	logger.TechLog.Debug(context.Background(), fmt.Sprintf(format, v...))
}

func NewClient(cfg config.Config) (*client, error) {
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

func getK8sConfig(cfg config.Config) (*rest.Config, error) {
	if cfg.Clients.HelmClient.KubeConfig != "" {
		return getK8sConfigFromKubeConfig(cfg)
	}
	if cfg.Clients.HelmClient.Token != "" {
		return getK8sConfigFromServiceAccount(cfg)
	}

	return nil, errors.New("no config for helm client found")
}

func getK8sConfigFromKubeConfig(cfg config.Config) (*rest.Config, error) {
	config, err := clientcmd.Load(([]byte)(cfg.Clients.HelmClient.KubeConfig))
	if err != nil {
		return nil, fmt.Errorf("error loading kubeconfig: %w", err)
	}

	clientConfig := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{})

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("error getting restconfig: %w", err)
	}

	return restConfig, nil

}

func getK8sConfigFromServiceAccount(cfg config.Config) (*rest.Config, error) {
	token := cfg.Clients.HelmClient.Token
	caCert := cfg.Clients.HelmClient.CA
	apiServer := cfg.Clients.HelmClient.APIServer

	restConfig := &rest.Config{
		Host:        apiServer,
		BearerToken: token,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: []byte(caCert),
		},
	}

	return restConfig, nil
}

func (c *client) renderTemplate(namespace, workbenchName string, apps []AppInstance) (string, error) {
	actionConfig := helmaction.Configuration{}
	client := helmaction.NewInstall(&actionConfig)
	client.DryRun = true
	client.ReleaseName = workbenchName
	client.Namespace = namespace
	client.ClientOnly = true

	appMaps := []map[string]interface{}{}
	for _, app := range apps {
		appMaps = append(appMaps, appToMap(app))
	}

	vals := map[string]interface{}{
		"namespace": namespace,
		"name":      workbenchName,
		"apps":      appMaps,
	}
	if len(c.cfg.Clients.HelmClient.ImagePullSecrets) != 0 {
		dockerConfig, err := EncodeRegistriesToDockerJSON(c.cfg.Clients.HelmClient.ImagePullSecrets)
		if err != nil {
			return "", fmt.Errorf("unable to encode registries: %w", err)
		}
		vals["imagePullSecret"] = map[string]string{
			"name":             "image-pull-secret",
			"dockerConfigJson": dockerConfig,
		}
	}

	release, err := client.Run(c.chart, vals)
	if err != nil {
		return "", fmt.Errorf("error rendering Helm template: %w", err)
	}

	return release.Manifest, nil
}

func (c *client) applyManifest(manifest, namespace string) error {
	decoder := yaml.NewYAMLOrJSONDecoder(bytes.NewReader([]byte(manifest)), 4096)
	for {
		var rawObj map[string]interface{}
		if err := decoder.Decode(&rawObj); err != nil {
			break
		}

		kind, ok := rawObj["kind"].(string)
		if !ok {
			continue
		}

		name, _ := rawObj["metadata"].(map[string]interface{})["name"].(string)

		gvr, err := c.getGroupVersionFromKind(kind)
		if err != nil {
			return fmt.Errorf("failed to get gvr from kind - %s", err)
		}

		if kind == "Namespace" {
			_, err := c.dynamicClient.Resource(gvr).Get(context.Background(), name, v1.GetOptions{})
			if k8serrors.IsNotFound(err) {
				logger.TechLog.Info(context.Background(), "Namespace missing, creating",
					zap.String("name", name),
					zap.String("manifest", manifest),
				)

				_, err := c.dynamicClient.Resource(gvr).Create(context.Background(), &unstructured.Unstructured{Object: rawObj}, v1.CreateOptions{})
				if err != nil {
					return fmt.Errorf("error creating namespace: %w", err)
				}
			} else if err != nil {
				return fmt.Errorf("error retrieving namespace: %w", err)
			}
			continue
		}

		existing, err := c.dynamicClient.Resource(gvr).Namespace(namespace).Get(context.Background(), name, v1.GetOptions{})
		if k8serrors.IsNotFound(err) {
			logger.TechLog.Info(context.Background(), "Missing resource, creating",
				zap.String("namespace", namespace), zap.String("kind", kind), zap.String("name", name),
				zap.String("manifest", manifest),
			)

			_, err := c.dynamicClient.Resource(gvr).Namespace(namespace).Create(context.Background(), &unstructured.Unstructured{Object: rawObj}, v1.CreateOptions{})
			if err != nil {
				return fmt.Errorf("error creating resource: %w", err)
			}
			continue
		}

		if err != nil {
			return fmt.Errorf("error retrieving resource: %w", err)
		}

		spec, hasSpec := rawObj["spec"]
		if !hasSpec {
			continue
		}

		existingSpec, exists := existing.Object["spec"]
		if !exists || !hasSpec {
			return fmt.Errorf("spec field not found in resource kind: %s, name: %s", kind, name)
		}

		desiredSpecBytes, _ := json.Marshal(spec)
		existingSpecBytes, _ := json.Marshal(existingSpec)

		patch, err := jsonmergepatch.CreateThreeWayJSONMergePatch(existingSpecBytes, desiredSpecBytes, existingSpecBytes)
		if err != nil {
			return fmt.Errorf("error calculating patch: %w", err)
		}

		if len(patch) > 0 && string(patch) != "{}" {
			updatedSpec := map[string]interface{}{
				"spec": json.RawMessage(desiredSpecBytes),
			}

			patch, _ := json.Marshal(updatedSpec)

			logger.TechLog.Info(context.Background(), "Resource not in the correct state, applying patch",
				zap.String("namespace", namespace), zap.String("kind", kind), zap.String("name", name),
				zap.String("manifest", manifest), zap.String("patch", string(patch)),
			)

			_, err := c.dynamicClient.Resource(gvr).Namespace(namespace).Patch(context.Background(), name, types.MergePatchType, patch, v1.PatchOptions{})
			if err != nil {
				return fmt.Errorf("error applying patch: %w", err)
			}
		}
	}

	return nil
}

func (c *client) getGroupVersionFromKind(kindName string) (schema.GroupVersionResource, error) {
	discoveryClient := discovery.NewDiscoveryClientForConfigOrDie(c.restConfig)

	apiResourceLists, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		return schema.GroupVersionResource{}, err
	}

	for _, apiResourceList := range apiResourceLists {
		for _, resource := range apiResourceList.APIResources {
			if resource.Kind == kindName {
				group, version := getGroupVersion(apiResourceList.GroupVersion)
				return schema.GroupVersionResource{Group: group, Version: version, Resource: resource.Name}, nil
			}
		}
	}
	return schema.GroupVersionResource{}, nil
}

func getGroupVersion(groupVersion string) (string, string) {
	if strings.Contains(groupVersion, "/") {
		arr := strings.Split(groupVersion, "/")
		return arr[0], arr[1]
	}
	return "", groupVersion
}

func (c *client) CreatePortForward(namespace, serviceName string) (uint16, chan struct{}, error) {
	pods, err := c.k8sClient.CoreV1().Pods(namespace).List(context.Background(), v1.ListOptions{
		LabelSelector: fmt.Sprintf("workbench=%s", serviceName),
	})
	if err != nil {
		return 0, nil, fmt.Errorf("unable to get pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return 0, nil, errors.New("no pods found for the service")
	}

	podName := pods.Items[0].Name
	ports := []string{"0:8080"}

	req := c.k8sClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(podName).
		SubResource("portforward")

	transport, upgrader, err := spdy.RoundTripperFor(c.restConfig)
	if err != nil {
		return 0, nil, fmt.Errorf("unable to get spdy round tripper: %w", err)
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL())

	stopChan := make(chan struct{}, 1)
	readyChan := make(chan struct{})
	out, errOut := io.Discard, io.Discard

	pf, err := portforward.New(dialer, ports, stopChan, readyChan, out, errOut)
	if err != nil {
		return 0, nil, fmt.Errorf("unable to create the port forwarder: %w", err)
	}

	go func() {
		if err := pf.ForwardPorts(); err != nil {
			// todo check if err is ErrLostConnectionToPod
			// if so recreacte portforward
			logger.TechLog.Error(context.Background(), "portforwarding error", zap.Error(err))
		}
	}()

	<-readyChan

	forwardedPorts, err := pf.GetPorts()
	if err != nil {
		return 0, nil, fmt.Errorf("unable to get ports: %w", err)
	}
	if len(forwardedPorts) != 1 {
		return 0, nil, errors.New("not right number of forwarded ports")
	}
	port := forwardedPorts[0]

	return port.Local, stopChan, nil
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
	manifest, err := c.renderTemplate(namespace, workbenchName, []AppInstance{})
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
	manifest, err := c.renderTemplate(namespace, workbenchName, apps)
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

	patch := map[string]interface{}{
		"op":    "add",
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
