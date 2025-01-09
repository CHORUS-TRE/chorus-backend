package helm

import (
	"errors"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

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
