package ociregistry

import (
	"net/url"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
)

type RegistryConfig struct {
	Registry string // e.g., "harbor.example.com"
	Username string
	Password string
}

type ClientConfig struct {
	Registries map[string]RegistryConfig
}

func getClientConfig(cfg config.Config) (ClientConfig, error) {
	registries := make(map[string]RegistryConfig)

	harborCfg := cfg.Clients.HarborClient
	if harborCfg.Enabled {
		if u, err := url.Parse(harborCfg.URL); err == nil && u.Host != "" {
			registries[u.Host] = RegistryConfig{
				Registry: u.Host,
				Username: harborCfg.Username,
				Password: harborCfg.Password.PlainText(),
			}
		}
	}

	return ClientConfig{
		Registries: registries,
	}, nil
}
