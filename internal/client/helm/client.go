package helm

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"go.uber.org/zap"
	helmaction "helm.sh/helm/v3/pkg/action"
	helmchart "helm.sh/helm/v3/pkg/chart"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

type HelmClienter interface {
	CreateWorkbench(namespace, workbenchName string) error
	CreatePortForward(namespace, serviceName string) (uint16, chan struct{}, error)
	CreateAppInstance(namespace, workbenchName, appName, appImage string) error
	DeleteApp(namespace, workbenchName, appName string) error
	DeleteWorkbench(namespace, workbenchName string) error
}

type client struct {
	cfg   config.Config
	chart *helmchart.Chart
}

func debug(format string, v ...interface{}) {
	logger.TechLog.Debug(context.Background(), fmt.Sprintf(format, v...))
}

func NewClient(cfg config.Config) (*client, error) {
	chart, err := GetHelmChart()
	if err != nil {
		return nil, fmt.Errorf("Error loading Helm chart: %w", err)
	}

	c := &client{
		chart: chart,
		cfg:   cfg,
	}
	return c, nil
}

func (c *client) getConfig(namespace string) (*helmaction.Configuration, error) {
	config, err := clientcmd.Load(([]byte)(c.cfg.Clients.HelmClient.KubeConfig))
	if err != nil {
		return nil, fmt.Errorf("Error loading kubeconfig: %w", err)
	}

	clientConfig := clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{})

	configFlags := &genericclioptions.ConfigFlags{
		Namespace: &namespace,
	}

	configFlags.WrapConfigFn = func(cfg *rest.Config) *rest.Config {
		clientConfig, err := clientConfig.ClientConfig()
		if err != nil {
			fmt.Printf("Error getting client config: %v\n", err)
			os.Exit(1)
		}
		return clientConfig
	}

	actionConfig := new(helmaction.Configuration)
	if err := actionConfig.Init(configFlags, namespace, "secret", debug); err != nil {
		return nil, fmt.Errorf("Error initializing Helm configuration: %w", err)
	}

	return actionConfig, nil

}

func (c *client) CreatePortForward(namespace, serviceName string) (uint16, chan struct{}, error) {
	helmConfig, err := c.getConfig(namespace)
	if err != nil {
		return 0, nil, fmt.Errorf("unable to get config: %w", err)
	}

	config, err := helmConfig.RESTClientGetter.ToRESTConfig()
	if err != nil {
		return 0, nil, fmt.Errorf("unable to convert to rest config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return 0, nil, fmt.Errorf("unable to get clienset: %w", err)
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(context.Background(), v1.ListOptions{
		LabelSelector: fmt.Sprintf("workbench=%s", serviceName),
	})
	if err != nil {
		return 0, nil, fmt.Errorf("unable to get pods: %w", err)
	}

	if len(pods.Items) == 0 {
		return 0, nil, errors.New("No pods found for the service")
	}

	podName := pods.Items[0].Name
	ports := []string{"0:8080"}

	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(podName).
		SubResource("portforward")

	transport, upgrader, err := spdy.RoundTripperFor(config)
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

func (c *client) CreateWorkbench(namespace, workbenchName string) error {
	actionConfig, err := c.getConfig(namespace)
	if err != nil {
		return fmt.Errorf("Unable to get config: %w", err)
	}

	install := helmaction.NewInstall(actionConfig)
	install.CreateNamespace = true
	install.Namespace = namespace
	install.ReleaseName = workbenchName
	vals := map[string]interface{}{
		"name": workbenchName,
		"apps": []map[string]string{},
	}

	_, err = install.Run(c.chart, vals)
	if err != nil {
		return fmt.Errorf("Failed to install workbench: %w", err)
	}

	return nil
}

func (c *client) CreateAppInstance(namespace, workbenchName, appName, appImage string) error {
	actionConfig, err := c.getConfig(namespace)
	if err != nil {
		return fmt.Errorf("Unable to get config: %w", err)
	}

	get := helmaction.NewGet(actionConfig)
	release, err := get.Run(workbenchName)
	if err != nil {
		return fmt.Errorf("Failed to get release: %w", err)
	}

	app := map[string]string{
		"app":  appName,
		"name": appName,
	}

	vals := release.Config
	vals["apps"] = append(vals["apps"].([]interface{}), app)

	upgrade := helmaction.NewUpgrade(actionConfig)
	upgrade.Namespace = namespace
	_, err = upgrade.Run(workbenchName, c.chart, vals)
	if err != nil {
		return fmt.Errorf("Failed to add app to workbench: %w", err)
	}

	return nil
}

func (c *client) DeleteApp(namespace, workbenchName, appName string) error {
	actionConfig, err := c.getConfig(namespace)
	if err != nil {
		return fmt.Errorf("Unable to get config: %w", err)
	}

	get := helmaction.NewGet(actionConfig)
	release, err := get.Run(workbenchName)
	if err != nil {
		return fmt.Errorf("Failed to get release: %w", err)
	}

	vals := release.Config
	apps := vals["apps"].([]interface{})
	for i, app := range apps {
		if app.(map[string]interface{})["name"] == appName {
			vals["apps"] = append(apps[:i], apps[i+1:]...)
			break
		}
	}

	upgrade := helmaction.NewUpgrade(actionConfig)
	upgrade.Namespace = namespace
	_, err = upgrade.Run(workbenchName, c.chart, vals)
	if err != nil {
		return fmt.Errorf("Failed to delete app from workbench: %w", err)
	}

	return nil
}

func (c *client) DeleteWorkbench(namespace, workbenchName string) error {
	actionConfig, err := c.getConfig(namespace)
	if err != nil {
		return fmt.Errorf("Unable to get config: %w", err)
	}

	uninstall := helmaction.NewUninstall(actionConfig)
	_, err = uninstall.Run(workbenchName)
	if err != nil {
		return fmt.Errorf("Failed to delete workbench: %w", err)
	}

	return nil
}
