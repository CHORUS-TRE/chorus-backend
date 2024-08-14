package helm

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/pkg/errors"
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
		return nil, errors.Wrapf(err, "Error loading Helm chart: %v", err)
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
		return nil, errors.Wrapf(err, "Error loading kubeconfig: %v", err)
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
		return nil, errors.Wrapf(err, "Error initializing Helm configuration: %v", err)
	}

	return actionConfig, nil

}

func (c *client) CreatePortForward(namespace, serviceName string) (uint16, chan struct{}, error) {
	helmConfig, err := c.getConfig(namespace)
	if err != nil {
		return 0, nil, errors.Wrap(err, "unable to get config")
	}

	config, err := helmConfig.RESTClientGetter.ToRESTConfig()
	if err != nil {
		return 0, nil, errors.Wrap(err, "unable to convert to rest config")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return 0, nil, errors.Wrap(err, "unable to get clienset")
	}

	pods, err := clientset.CoreV1().Pods(namespace).List(context.Background(), v1.ListOptions{
		LabelSelector: fmt.Sprintf("workbench=%s", serviceName),
	})
	if err != nil {
		return 0, nil, errors.Wrap(err, "unable to get pods")
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
		return 0, nil, errors.Wrap(err, "unable to get spdy round tripper")
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL())

	stopChan := make(chan struct{}, 1)
	readyChan := make(chan struct{})
	out, errOut := io.Discard, io.Discard

	pf, err := portforward.New(dialer, ports, stopChan, readyChan, out, errOut)
	if err != nil {
		return 0, nil, errors.Wrap(err, "unable to create the port forwarder")
	}

	go func() {
		if err := pf.ForwardPorts(); err != nil {
			logger.TechLog.Error(context.Background(), "portforwarding error", zap.Error(err))
		}
	}()

	<-readyChan

	forwardedPorts, err := pf.GetPorts()
	if err != nil {
		return 0, nil, errors.Wrap(err, "unable to get ports")
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
		return errors.Wrapf(err, "Unable to get config: %v", err)
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
		return errors.Wrapf(err, "Failed to install workbench: %v", err)
	}

	return nil
}

func (c *client) CreateAppInstance(namespace, workbenchName, appName, appImage string) error {
	actionConfig, err := c.getConfig(namespace)
	if err != nil {
		return errors.Wrapf(err, "Unable to get config: %v", err)
	}

	get := helmaction.NewGet(actionConfig)
	release, err := get.Run(workbenchName)
	if err != nil {
		return errors.Wrapf(err, "Failed to get release: %v", err)
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
		return errors.Wrapf(err, "Failed to add app to workbench: %v", err)
	}

	return nil
}

func (c *client) DeleteApp(namespace, workbenchName, appName string) error {
	actionConfig, err := c.getConfig(namespace)
	if err != nil {
		return errors.Wrapf(err, "Unable to get config: %v", err)
	}

	get := helmaction.NewGet(actionConfig)
	release, err := get.Run(workbenchName)
	if err != nil {
		return errors.Wrapf(err, "Failed to get release: %v", err)
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
		return errors.Wrapf(err, "Failed to delete app from workbench: %v", err)
	}

	return nil
}

func (c *client) DeleteWorkbench(namespace, workbenchName string) error {
	actionConfig, err := c.getConfig(namespace)
	if err != nil {
		return errors.Wrapf(err, "Unable to get config: %v", err)
	}

	uninstall := helmaction.NewUninstall(actionConfig)
	_, err = uninstall.Run(workbenchName)
	if err != nil {
		return errors.Wrapf(err, "Failed to delete workbench: %v", err)
	}

	return nil
}
