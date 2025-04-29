package k8s

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"go.uber.org/zap"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

type K8sClienter interface {
	CreateWorkspace(tenantID uint64, namespace string) error
	DeleteWorkspace(namespace string) error
	CreateWorkbench(tenantID uint64, namespace, workbenchName string) error
	UpdateWorkbench(tenantID uint64, namespace, workbenchName string, apps []AppInstance) error
	CreatePortForward(namespace, serviceName string) (uint16, chan struct{}, error)
	CreateAppInstance(namespace, workbenchName string, app AppInstance) error
	DeleteAppInstance(namespace, workbenchName string, appInstance AppInstance) error
	DeleteWorkbench(namespace, workbenchName string) error

	WatchOnNewWorkbench(func(workbench *Workbench) error) error
	WatchOnUpdateWorkbench(func(oldWorkbench, newWorkbench *Workbench) error) error
	WatchOnDeleteWorkbench(func(workbench *Workbench) error) error
}

type client struct {
	cfg           config.Config
	restConfig    *rest.Config
	k8sClient     *kubernetes.Clientset
	dynamicClient *dynamic.DynamicClient
	gvrCache      map[string]schema.GroupVersionResource
	gvrCacheLock  sync.Mutex

	onNewWorkbench    func(workbench *Workbench) error
	onUpdateWorkbench func(oldWorkbench, newWorkbench *Workbench) error
	onDeleteWorkbench func(workbench *Workbench) error
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
		go c.watch()
	}

	return c, nil
}

func (c *client) watch() {
	factory := dynamicinformer.NewDynamicSharedInformerFactory(c.dynamicClient, 0)

	// namespaceGvr, err := c.getGroupVersionFromKind("Namespace")
	// if err != nil {
	// 	logger.TechLog.Error(context.Background(), "Error getting GVR for namespace:", zap.Error(err))
	// 	return
	// }
	// namespaceInformer := factory.ForResource(namespaceGvr).Informer()
	// namespaceInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
	// 	AddFunc: func(obj interface{}) {
	// 		logger.TechLog.Info(context.Background(), "Watcher event: added namespace:")
	// 		namespace, err := EventInterfaceToNamespace(obj)
	// 		if err != nil {
	// 			logger.TechLog.Error(context.Background(), "Error converting to Namespace:", zap.Error(err))
	// 			return
	// 		}
	// 	},
	// 	UpdateFunc: func(oldObj, newObj interface{}) {
	// 		logger.TechLog.Info(context.Background(), "Watcher event: updated namespace:")
	// 		newNamespace, err := EventInterfaceToNamespace(newObj)
	// 		if err != nil {
	// 			logger.TechLog.Error(context.Background(), "Error converting to Namespace:", zap.Error(err))
	// 			return
	// 		}
	// 		oldNamespace, err := EventInterfaceToNamespace(oldObj)
	// 		if err != nil {
	// 			logger.TechLog.Error(context.Background(), "Error converting to Namespace:", zap.Error(err))
	// 			return
	// 		}
	// 	},
	// 	DeleteFunc: func(obj interface{}) {
	// 		logger.TechLog.Info(context.Background(), "Watcher event: deleted namespace:")
	// 		namespace, err := EventInterfaceToNamespace(obj)
	// 		if err != nil {
	// 			logger.TechLog.Error(context.Background(), "Error converting to Namespace:", zap.Error(err))
	// 			return
	// 		}
	// 	},
	// })

	workbenchGvr, err := c.getGroupVersionFromKind("Workbench")
	if err != nil {
		logger.TechLog.Error(context.Background(), "Error getting GVR for Workbench", zap.Error(err))
		return
	}
	workbenchInformer := factory.ForResource(workbenchGvr).Informer()
	workbenchInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			logger.TechLog.Debug(context.Background(), "added workbench", zap.Any("workbench", obj))
			workbench, err := EventInterfaceToWorkbench(obj)
			if err != nil {
				logger.TechLog.Error(context.Background(), "Error converting to Workbench:", zap.Error(err))
				return
			}
			if c.onNewWorkbench != nil {
				if err := c.onNewWorkbench(workbench); err != nil {
					logger.TechLog.Error(context.Background(), "Error handling new workbench:", zap.Error(err))
				}
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			logger.TechLog.Debug(context.Background(), "updated workbench", zap.Any("newWorkbench", newObj), zap.Any("oldWorkbench", oldObj))
			newWorkbench, err := EventInterfaceToWorkbench(newObj)
			if err != nil {
				logger.TechLog.Error(context.Background(), "Error converting to Workbench:", zap.Error(err))
				return
			}
			oldWorkbench, err := EventInterfaceToWorkbench(oldObj)
			if err != nil {
				logger.TechLog.Error(context.Background(), "Error converting to Workbench:", zap.Error(err))
				return
			}
			if c.onUpdateWorkbench != nil {
				if err := c.onUpdateWorkbench(oldWorkbench, newWorkbench); err != nil {
					logger.TechLog.Error(context.Background(), "Error handling updated workbench:", zap.Error(err))
				}
			}
		},
		DeleteFunc: func(obj interface{}) {
			logger.TechLog.Debug(context.Background(), "deleted workbench", zap.Any("workbench", obj))
			workbench, err := EventInterfaceToWorkbench(obj)
			if err != nil {
				logger.TechLog.Error(context.Background(), "Error converting to Workbench:", zap.Error(err))
				return
			}
			if c.onDeleteWorkbench != nil {
				if err := c.onDeleteWorkbench(workbench); err != nil {
					logger.TechLog.Error(context.Background(), "Error handling deleted workbench:", zap.Error(err))
				}
			}
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

func (c *client) WatchOnNewWorkbench(handler func(workbench *Workbench) error) error {
	c.onNewWorkbench = handler
	return nil
}
func (c *client) WatchOnUpdateWorkbench(handler func(oldWorkbench, newWorkbench *Workbench) error) error {
	c.onUpdateWorkbench = handler
	return nil
}
func (c *client) WatchOnDeleteWorkbench(handler func(workbench *Workbench) error) error {
	c.onDeleteWorkbench = handler
	return nil
}

func (c *client) makeWorkbench(tenantID uint64, namespace, workbenchName string, apps []AppInstance) (Workbench, error) {
	workbench := Workbench{
		TypeMeta: v1.TypeMeta{
			Kind:       "Workbench",
			APIVersion: "default.chorus-tre.ch/v1alpha1",
		},
		ObjectMeta: v1.ObjectMeta{
			Name:      workbenchName,
			Namespace: namespace,
			Labels: map[string]string{
				"chorus-tre.ch/created-by": "chorus-backend",
				"chorus-tre.ch/tenant-id":  fmt.Sprintf("%d", tenantID),
			},
		},
		Spec: WorkbenchSpec{
			Apps: []WorkbenchApp{},
		},
	}

	for _, app := range apps {
		workbenchApp := c.appInstanceToWorkbenchApp(app)
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

func (c *client) CreateWorkspace(tenantID uint64, namespace string) error {
	return c.syncNamespace(tenantID, namespace)
}

func (c *client) DeleteWorkspace(namespace string) error {
	return c.deleteNamespace(namespace)
}

func (c *client) CreateWorkbench(tenantID uint64, namespace, workbenchName string) error {
	workbench, err := c.makeWorkbench(tenantID, namespace, workbenchName, []AppInstance{})
	if err != nil {
		return fmt.Errorf("error creating workbench: %w", err)
	}

	return c.syncWorkbench(tenantID, workbench, namespace)
}

func (c *client) UpdateWorkbench(tenantID uint64, namespace, workbenchName string, apps []AppInstance) error {
	workbench, err := c.makeWorkbench(tenantID, namespace, workbenchName, apps)
	if err != nil {
		return fmt.Errorf("error creating workbench: %w", err)
	}

	return c.syncWorkbench(tenantID, workbench, namespace)
}

func (c *client) CreateAppInstance(namespace, workbenchName string, appInstance AppInstance) error {
	app := c.appInstanceToWorkbenchApp(appInstance)

	gvr, err := c.getGroupVersionFromKind("Workbench")
	if err != nil {
		return fmt.Errorf("failed to get gvr from kind - %s", err)
	}

	patch := []map[string]interface{}{
		{
			"op":    "add",
			"path":  "/spec/apps/-",
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
				"value": []interface{}{app},
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
		return fmt.Errorf("error applying patch: %w", err)
	}

	return nil
}

func (c *client) DeleteAppInstance(namespace, workbenchName string, appInstance AppInstance) error {
	app := c.appInstanceToWorkbenchApp(appInstance)

	// Fetch the current workbench
	gvr, err := c.getGroupVersionFromKind("Workbench")
	if err != nil {
		return fmt.Errorf("failed to get gvr from kind - %s", err)
	}

	workbench, err := c.dynamicClient.Resource(gvr).Namespace(namespace).Get(context.Background(), workbenchName, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get workbench: %w", err)
	}

	// Find the index
	apps, found, err := unstructured.NestedSlice(workbench.Object, "spec", "apps")
	if err != nil || !found {
		return fmt.Errorf("apps field not found: %w", err)
	}

	indexToRemove := -1
	for i, a := range apps {
		appMap, ok := a.(map[string]interface{})
		if !ok {
			continue
		}
		if appMap["name"] == app.Name {
			indexToRemove = i
			break
		}
	}

	if indexToRemove == -1 {
		return fmt.Errorf("app instance %s not found", app.Name)
	}

	patch := []map[string]interface{}{
		{
			"op":   "remove",
			"path": fmt.Sprintf("/spec/apps/%d", indexToRemove),
		},
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("error marshalling patch: %w", err)
	}

	logger.TechLog.Debug(context.Background(), "delete app instance patchBytes", zap.String("patchBytes", string(patchBytes)))

	_, err = c.dynamicClient.Resource(gvr).Namespace(namespace).Patch(context.Background(), workbenchName, types.JSONPatchType, patchBytes, v1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("error applying patch: %w", err)
	}

	return nil
}

func (c *client) DeleteWorkbench(namespace, workbenchName string) error {
	return c.deleteResource(namespace, "Workbench", workbenchName)
}
