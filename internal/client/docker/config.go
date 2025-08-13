package docker

import (
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
)

type DockerRegistryConfig struct {
	Registry string // e.g., "harbor.example.com"
	Username string
	Password string
}

type DockerClientConfig struct {
	Registries map[string]DockerRegistryConfig
}

func getDockerClientConfig(cfg config.Config) (DockerClientConfig, error) {
	registries := make(map[string]DockerRegistryConfig)
	imagePullSecrets := cfg.Clients.K8sClient.ImagePullSecrets // Get the registries from k8sClient ImagePullSecrets

	for _, secret := range imagePullSecrets {
		registryCfg := DockerRegistryConfig{
			Registry: secret.Registry,
			Username: secret.Username,
			Password: secret.Password,
		}

		registries[secret.Registry] = registryCfg
	}

	return DockerClientConfig{
		Registries: registries,
	}, nil
}
