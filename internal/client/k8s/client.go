package k8s

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"k8s.io/utils/ptr"
)

var _ = K8sClienter(&client{})

type K8sClienter interface {
	// Workspace (Namespace) operations
	CreateWorkspace(tenantID uint64, namespace string) error
	DeleteWorkspace(namespace string) error

	// Workbench operations
	CreateWorkbench(workbench *Workbench) error
	UpdateWorkbench(workbench *Workbench) error
	DeleteWorkbench(namespace, workbenchName string) error

	// AppInstance operations
	CreateAppInstance(namespace, workbenchName string, app AppInstance) error
	DeleteAppInstance(namespace, workbenchName string, appInstance AppInstance) error

	// Utility operations
	CreatePortForward(namespace, serviceName string) (uint16, chan struct{}, error)
	PrePullImageOnAllNodes(image string)

	// Watcher operations
	RegisterOnNewWorkbenchHandler(func(workbench Workbench) error) error
	RegisterOnUpdateWorkbenchHandler(func(workbench Workbench) error) error
	RegisterOnDeleteWorkbenchHandler(func(workbench Workbench) error) error
}

type client struct {
	cfg           config.Config
	restConfig    *rest.Config
	k8sClient     *kubernetes.Clientset
	dynamicClient *dynamic.DynamicClient
	gvrCache      map[string]schema.GroupVersionResource
	gvrCacheLock  sync.Mutex

	onNewWorkbench    func(workbench Workbench) error
	onUpdateWorkbench func(workbench Workbench) error
	onDeleteWorkbench func(workbench Workbench) error
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
		gvrCache:      make(map[string]schema.GroupVersionResource),
	}

	if cfg.Clients.K8sClient.IsWatcher {
		go c.watchWorkbenchEvents()
	}

	return c, nil
}

// ----------------------------------------------------------------
// Internal watchers setup
// ----------------------------------------------------------------
func (c *client) watchWorkbenchEvents() {
	factory := dynamicinformer.NewDynamicSharedInformerFactory(c.dynamicClient, 0)

	// Get GVR for Workbench
	workbenchGvr, err := c.getGroupVersionFromKind("Workbench")
	if err != nil {
		logger.TechLog.Error(context.Background(), "Error getting GVR for Workbench", zap.Error(err))
		return
	}

	workbenchInformer := factory.ForResource(workbenchGvr).Informer()
	workbenchInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			c.handleWorkbenchEvent(obj, "added", c.onNewWorkbench)
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			c.handleWorkbenchEvent(newObj, "updated", c.onUpdateWorkbench)
		},
		DeleteFunc: func(obj interface{}) {
			c.handleWorkbenchEvent(obj, "deleted", c.onDeleteWorkbench)
		},
	})

	logger.TechLog.Info(context.Background(), "Starting informers...")
	stopCh := make(chan struct{})
	factory.Start(stopCh)
	factory.WaitForCacheSync(stopCh)

	logger.TechLog.Info(context.Background(), "Informers started and caches synced.")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
	logger.TechLog.Info(context.Background(), "Received interrupt signal, shutting down informers...")
	close(stopCh)

	logger.TechLog.Info(context.Background(), "Stopping informers...")
	factory.Shutdown()
	logger.TechLog.Info(context.Background(), "Informers stopped.")
}

// Generic handler for workbench events
func (c *client) handleWorkbenchEvent(obj any, eventType string, handler func(workbench Workbench) error) {
	logger.TechLog.Debug(context.Background(), fmt.Sprintf("%s workbench", eventType), zap.Any("workbench", obj))

	if obj == nil {
		logger.TechLog.Error(context.Background(), fmt.Sprintf("nil object received during %s workbench event", eventType))
		return
	}

	workbench, err := c.eventInterfaceToWorkbench(obj)
	if err != nil {
		logger.TechLog.Error(context.Background(), "Error converting event interface to Workbench", zap.Error(err))
		return
	}

	if handler != nil {
		if err := handler(workbench); err != nil {
			logger.TechLog.Error(context.Background(), fmt.Sprintf("Error handling %s workbench event", eventType), zap.Error(err))
		}
	}
}

// ----------------------------------------------------------------
// Workspace (Namespace) operations
// ----------------------------------------------------------------
func (c *client) CreateWorkspace(tenantID uint64, namespace string) error {
	return c.syncNamespace(tenantID, namespace)
}

func (c *client) DeleteWorkspace(namespace string) error {
	return c.deleteNamespace(namespace)
}

// ----------------------------------------------------------------
// Workbench operations
// ----------------------------------------------------------------
func (c *client) CreateWorkbench(workbench *Workbench) error {
	k8sWorkbench, err := c.workbenchToK8sWorkbench(workbench)
	if err != nil {
		return fmt.Errorf("error creating workbench: %w", err)
	}

	return c.syncWorkbench(workbench.TenantID, k8sWorkbench, workbench.Namespace)
}

func (c *client) UpdateWorkbench(workbench *Workbench) error {
	k8sWorkbench, err := c.workbenchToK8sWorkbench(workbench)
	if err != nil {
		return fmt.Errorf("error creating workbench: %w", err)
	}

	return c.syncWorkbench(workbench.TenantID, k8sWorkbench, workbench.Namespace)
}

func (c *client) DeleteWorkbench(namespace, workbenchName string) error {
	return c.deleteResource(namespace, "Workbench", workbenchName)
}

// ----------------------------------------------------------------
// AppInstance operations
// ----------------------------------------------------------------
func (c *client) CreateAppInstance(namespace, workbenchName string, appInstance AppInstance) error {
	app := c.appInstanceToK8sWorkbenchApp(appInstance)

	gvr, err := c.getGroupVersionFromKind("Workbench")
	if err != nil {
		return fmt.Errorf("failed to get gvr from kind - %s", err)
	}

	patch := []map[string]interface{}{
		{
			"op":    "add",
			"path":  "/spec/apps/" + appInstance.UID(),
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
		logger.TechLog.Info(context.Background(), "not found")
		patch = []map[string]interface{}{
			{
				"op":   "add",
				"path": "/spec/apps",
				// "value": []map[string]interface{}{app},
				"value": map[string]interface{}{appInstance.UID(): app},
			},
		}

	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("error marshalling patch: %w", err)
	}

	logger.TechLog.Debug(context.Background(), "create app instance update patchBytes", zap.String("patchBytes", string(patchBytes)))

	_, err = c.dynamicClient.Resource(gvr).Namespace(namespace).Patch(context.Background(), workbenchName, types.JSONPatchType, patchBytes, v1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("error applying patch create appInstance (%s): %w", string(patchBytes), err)
	}

	return nil
}

func (c *client) DeleteAppInstance(namespace, workbenchName string, appInstance AppInstance) error {

	gvr, err := c.getGroupVersionFromKind("Workbench")
	if err != nil {
		return fmt.Errorf("failed to get gvr from kind - %s", err)
	}

	patch := []map[string]interface{}{
		{
			"op":   "remove",
			"path": fmt.Sprintf("/spec/apps/%s", appInstance.UID()),
		},
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("error marshalling patch: %w", err)
	}

	logger.TechLog.Debug(context.Background(), "delete app instance patchBytes", zap.String("patchBytes", string(patchBytes)))

	_, err = c.dynamicClient.Resource(gvr).Namespace(namespace).Patch(context.Background(), workbenchName, types.JSONPatchType, patchBytes, v1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("error applying patch delete app instance (%s): %w", string(patchBytes), err)
	}

	return nil
}

// ----------------------------------------------------------------
// Utility operations
// ----------------------------------------------------------------
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

func (c *client) PrePullImageOnAllNodes(image string) {
	err := c.syncImagePullSecret("default")
	if err != nil {
		logger.TechLog.Error(context.Background(), "failed to sync image pull secret",
			zap.String("image", image),
			zap.Error(err),
		)
		return
	}

	nodeList, err := c.k8sClient.CoreV1().Nodes().List(context.Background(), v1.ListOptions{})
	if err != nil {
		logger.TechLog.Error(context.Background(), "failed to list nodes while pre-pulling image",
			zap.String("image", image),
			zap.Error(err),
		)

		return
	}

	for _, node := range nodeList.Items {
		job := &batchv1.Job{
			ObjectMeta: v1.ObjectMeta{
				GenerateName: "prepull-",
				Namespace:    "default",
			},
			Spec: batchv1.JobSpec{
				TTLSecondsAfterFinished: ptr.To(int32(60)),
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						NodeName:      node.Name,
						RestartPolicy: corev1.RestartPolicyNever,
						Containers: []corev1.Container{
							{
								Name:    "puller",
								Image:   image,
								Command: []string{"bash", "-c", "exit"},
							},
						},
						ImagePullSecrets: []corev1.LocalObjectReference{
							{
								Name: c.cfg.Clients.K8sClient.ImagePullSecretName,
							},
						},
					},
				},
			},
			TypeMeta: v1.TypeMeta{},
			Status:   batchv1.JobStatus{},
		}

		_, err := c.k8sClient.BatchV1().Jobs("default").Create(context.Background(), job, v1.CreateOptions{})
		if err != nil {
			logger.TechLog.Error(context.Background(), "failed to create job for pre-pulling image",
				zap.String("image", image),
				zap.String("node", node.Name),
				zap.Error(err),
			)
		} else {
			logger.TechLog.Info(context.Background(), "successfully created job for pre-pulling image",
				zap.String("image", image),
				zap.String("node", node.Name),
			)
		}
	}
}

// ----------------------------------------------------------------
// Watcher registration methods
// ----------------------------------------------------------------
func (c *client) RegisterOnNewWorkbenchHandler(handler func(workbench Workbench) error) error {
	c.onNewWorkbench = handler
	return nil
}
func (c *client) RegisterOnUpdateWorkbenchHandler(handler func(workbench Workbench) error) error {
	c.onUpdateWorkbench = handler
	return nil
}
func (c *client) RegisterOnDeleteWorkbenchHandler(handler func(workbench Workbench) error) error {
	c.onDeleteWorkbench = handler
	return nil
}
