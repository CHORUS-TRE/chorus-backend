package k8s

import (
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func getK8sConfig(cfg config.Config) (restConfig *rest.Config, err error) {
	if cfg.Clients.K8sClient.InClusterConfigEnabled {
		restConfig, err = rest.InClusterConfig()
	} else {
		restConfig, err = getK8sConfigFromKubeConfig(cfg)
	}
	if err != nil {
		return nil, err
	}

	if cfg.Clients.K8sClient.InsecureTLS {
		restConfig.TLSClientConfig.Insecure = true
		restConfig.TLSClientConfig.CAData = nil
		restConfig.TLSClientConfig.CAFile = ""
	}

	return restConfig, nil
}

func getK8sConfigFromKubeConfig(cfg config.Config) (*rest.Config, error) {
	config, err := clientcmd.LoadFromFile(cfg.Clients.K8sClient.KubeConfig)
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
