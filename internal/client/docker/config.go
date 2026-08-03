package docker

import (
	"net/url"

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

	harborCfg := cfg.Clients.HarborClient
	if harborCfg.Enabled {
		if u, err := url.Parse(harborCfg.URL); err == nil && u.Host != "" {
			registries[u.Host] = DockerRegistryConfig{
				Registry: u.Host,
				Username: harborCfg.Username,
				Password: harborCfg.Password.PlainText(),
			}
		}
	}

	return DockerClientConfig{
		Registries: registries,
	}, nil
}
