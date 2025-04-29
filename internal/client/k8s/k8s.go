package k8s

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	jsonpatch "github.com/evanphx/json-patch"
	"go.uber.org/zap"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

const (
	DEFAULT_POLL_INTERVAL = 500 * time.Millisecond
)

func (c *client) syncWorkbench(tenantID uint64, workbench Workbench, namespace string) error {
	logger.TechLog.Debug(context.Background(), "syncing workbench",
		zap.String("namespace", namespace), zap.Any("workbench", workbench), zap.Uint64("tenantID", tenantID),
	)

	kind := "Workbench"
	name := workbench.Name

	err := c.syncNamespace(tenantID, namespace)
	if err != nil {
		return fmt.Errorf("error syncing namespace: %w", err)
	}

	err = c.syncImagePullSecret(namespace)
	if err != nil {
		// TODO fix
		// return fmt.Errorf("error syncing image pull secret: %w", err)
		fmt.Println("error syncing image pull secret: %w", err)
	}

	return c.syncResource(workbench, kind, name, namespace, "spec")
}

func (c *client) syncResource(spec interface{}, kind, name, namespace, specFieldName string) error {
	logger.TechLog.Debug(context.Background(), "syncing resource",
		zap.String("namespace", namespace), zap.Any("kind", kind),
	)

	gvr, err := c.getGroupVersionFromKind(kind)
	if err != nil {
		return fmt.Errorf("failed to get gvr from kind - %s", err)
	}

	rawSpecs, err := c.interfaceToMapInterface(spec)
	if err != nil {
		return fmt.Errorf("error converting spec to map: %w", err)
	}

	fmt.Println("rawSpecs", rawSpecs)

	existing, err := c.dynamicClient.Resource(gvr).Namespace(namespace).Get(context.Background(), name, v1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		logger.TechLog.Info(context.Background(), "Missing resource, creating",
			zap.String("namespace", namespace), zap.String("kind", kind), zap.String("name", name),
		)

		_, err := c.dynamicClient.Resource(gvr).Namespace(namespace).Create(context.Background(), &unstructured.Unstructured{Object: rawSpecs}, v1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("error creating resource: %w", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("error retrieving resource: %w", err)
	}

	existingSpec, exists := existing.Object[specFieldName]
	if !exists {
		return fmt.Errorf(specFieldName+" field not found in resource kind: %s, name: %s", kind, name)
	}

	desiredSpecBytes, err := json.Marshal(rawSpecs[specFieldName])
	if err != nil {
		return fmt.Errorf("error marshalling desired spec: %w", err)
	}
	existingSpecBytes, err := json.Marshal(existingSpec)
	if err != nil {
		return fmt.Errorf("error marshalling existing spec: %w", err)
	}

	patch, err := jsonpatch.CreateMergePatch(existingSpecBytes, desiredSpecBytes)
	if err != nil {
		return fmt.Errorf("error calculating patch: %w", err)
	}

	if len(patch) > 0 && string(patch) != "{}" {
		updatedSpec := map[string]interface{}{
			specFieldName: json.RawMessage(desiredSpecBytes),
		}

		patch, err := json.Marshal(updatedSpec)
		if err != nil {
			return fmt.Errorf("error marshalling patch: %w", err)
		}

		logger.TechLog.Info(context.Background(), "Resource not in the correct state, applying patch",
			zap.String("namespace", namespace), zap.String("kind", kind), zap.String("name", name),
			zap.String("patch", string(patch)),
		)

		_, err = c.dynamicClient.Resource(gvr).Namespace(namespace).Patch(context.Background(), name, types.MergePatchType, patch, v1.PatchOptions{})
		if err != nil {
			return fmt.Errorf("error applying patch: %w", err)
		}
	}

	return nil
}

func (c *client) syncNamespace(tenantID uint64, namespace string) error {
	gvr, err := c.getGroupVersionFromKind("Namespace")
	if err != nil {
		return fmt.Errorf("failed to get gvr from kind - %s", err)
	}

	_, err = c.dynamicClient.Resource(gvr).Get(context.Background(), namespace, v1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		logger.TechLog.Info(context.Background(), "Namespace missing, creating",
			zap.String("namespace", namespace),
		)

		rawObj := map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"name": namespace,
				"labels": map[string]interface{}{
					"chorus-tre.ch/created-by": "chorus-backend",
					"chorus-tre.ch/tenant-id":  fmt.Sprintf("%d", tenantID),
				},
			},
		}

		_, err := c.dynamicClient.Resource(gvr).Create(context.Background(), &unstructured.Unstructured{Object: rawObj}, v1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("error creating namespace: %w", err)
		}

		return nil
	}
	if err != nil {
		return fmt.Errorf("error retrieving namespace: %w", err)
	}

	return nil
}

func (c *client) deleteNamespace(namespace string) error {
	gvr, err := c.getGroupVersionFromKind("Namespace")
	if err != nil {
		return fmt.Errorf("failed to get gvr from kind - %s", err)
	}

	_, err = c.dynamicClient.Resource(gvr).Get(context.Background(), namespace, v1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		logger.TechLog.Info(context.Background(), "Namespace already deleted",
			zap.String("namespace", namespace),
		)
		return nil
	}
	if err != nil {
		return fmt.Errorf("error retrieving namespace: %w", err)
	}
	logger.TechLog.Info(context.Background(), "Deleting namespace",
		zap.String("namespace", namespace),
	)

	err = c.dynamicClient.Resource(gvr).Delete(context.Background(), namespace, v1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("error deleting namespace: %w", err)
	}

	// Wait for the namespace to be deleted
	for {
		_, err = c.dynamicClient.Resource(gvr).Get(context.Background(), namespace, v1.GetOptions{})
		if k8serrors.IsNotFound(err) {
			logger.TechLog.Info(context.Background(), "Namespace deleted",
				zap.String("namespace", namespace),
			)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving namespace: %w", err)
		}
		logger.TechLog.Info(context.Background(), "Waiting for namespace to be deleted",
			zap.String("namespace", namespace),
		)
		time.Sleep(DEFAULT_POLL_INTERVAL)
	}
}

func (c *client) deleteResource(namespace, kind, name string) error {
	gvr, err := c.getGroupVersionFromKind(kind)
	if err != nil {
		return fmt.Errorf("failed to get gvr from kind - %s", err)
	}

	_, err = c.dynamicClient.Resource(gvr).Namespace(namespace).Get(context.Background(), name, v1.GetOptions{})
	if k8serrors.IsNotFound(err) {
		logger.TechLog.Info(context.Background(), "Resource already deleted",
			zap.String("namespace", namespace), zap.String("kind", kind), zap.String("name", name),
		)
		return nil
	}
	if err != nil {
		return fmt.Errorf("error retrieving resource: %w", err)
	}
	logger.TechLog.Info(context.Background(), "Deleting resource",
		zap.String("namespace", namespace), zap.String("kind", kind), zap.String("name", name),
	)

	err = c.dynamicClient.Resource(gvr).Namespace(namespace).Delete(context.Background(), name, v1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("error deleting resource: %w", err)
	}

	// Wait for the resource to be deleted
	for {
		_, err = c.dynamicClient.Resource(gvr).Namespace(namespace).Get(context.Background(), name, v1.GetOptions{})
		if k8serrors.IsNotFound(err) {
			logger.TechLog.Info(context.Background(), "Resource deleted",
				zap.String("namespace", namespace), zap.String("kind", kind), zap.String("name", name),
			)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving resource: %w", err)
		}

		logger.TechLog.Info(context.Background(), "Waiting for resource to be deleted",
			zap.String("namespace", namespace), zap.String("kind", kind), zap.String("name", name),
		)
		time.Sleep(DEFAULT_POLL_INTERVAL)
	}
}

func (c *client) syncImagePullSecret(namespace string) error {
	if len(c.cfg.Clients.K8sClient.ImagePullSecrets) == 0 {
		return nil
	}

	secretName := c.cfg.Clients.K8sClient.ImagePullSecretName

	dockerConfig, err := EncodeRegistriesToDockerJSON(c.cfg.Clients.K8sClient.ImagePullSecrets)
	if err != nil {
		return fmt.Errorf("unable to encode registries: %w", err)
	}

	dockerConfigBase64 := base64.StdEncoding.EncodeToString([]byte(dockerConfig))

	spec := map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Secret",
		"metadata": map[string]interface{}{
			"name": secretName,
		},
		"type": "kubernetes.io/dockerconfigjson",
		"data": map[string]interface{}{
			".dockerconfigjson": dockerConfigBase64,
		},
	}

	return c.syncResource(spec, "Secret", secretName, namespace, "data")
}

func (c *client) interfaceToMapInterface(i interface{}) (map[string]interface{}, error) {
	raw, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	err = json.Unmarshal(raw, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (c *client) getGroupVersionFromKind(kindName string) (schema.GroupVersionResource, error) {
	c.gvrCacheLock.Lock()
	defer c.gvrCacheLock.Unlock()
	if cachedGvr, ok := c.gvrCache[kindName]; ok {
		return cachedGvr, nil
	}

	discoveryClient := discovery.NewDiscoveryClientForConfigOrDie(c.restConfig)

	apiResourceLists, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		return schema.GroupVersionResource{}, err
	}

	for _, apiResourceList := range apiResourceLists {
		for _, resource := range apiResourceList.APIResources {
			if resource.Kind == kindName {
				group, version := getGroupVersion(apiResourceList.GroupVersion)
				gvr := schema.GroupVersionResource{Group: group, Version: version, Resource: resource.Name}
				c.gvrCache[kindName] = gvr
				return gvr, nil
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
