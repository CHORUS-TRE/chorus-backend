package docker

import (
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
)

var _ DockerClienter = &client{}

type DockerClienter interface {
	ImageExists(imageRef string, username string, password string) (bool, error)
}

type client struct {
	cfg             config.Config
	dockerClientCfg DockerClientConfig
}

func NewClient(cfg config.Config) (*client, error) {
	clientCfg, err := getDockerClientConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("error getting docker config: %w", err)
	}

	return &client{
		cfg:             cfg,
		dockerClientCfg: clientCfg,
	}, nil
}

// ImageExists checks whether an image exists in registry
func (c *client) ImageExists(imageRef string, username string, password string) (bool, error) {
	// Parse image reference
	ref, err := name.ParseReference(imageRef, name.WeakValidation)
	if err != nil {
		return false, fmt.Errorf("invalid docker image reference: %w", err)
	}

	registry := ref.Context().RegistryStr()
	authenticator, err := c.getRegistryAuth(registry, username, password)
	if err != nil {
		return false, fmt.Errorf("failed to get docker registry auth: %w", err)
	}

	// Use GET request directly since some registries (like Harbor) don't support HEAD properly
	_, err = remote.Get(ref, remote.WithAuth(authenticator))
	if err != nil {
		// Check if it's a "not found" error
		if isNotFoundError(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if docker image exists: %w", err)
	}

	return true, nil
}

func (c *client) getRegistryAuth(registry string, username string, password string) (authn.Authenticator, error) {
	if registry == "" {
		return nil, fmt.Errorf("docker registry hostname cannot be empty")
	}

	// If credentials are provided, use them
	if username != "" && password != "" {
		return authn.FromConfig((authn.AuthConfig{
			Username: username,
			Password: password,
		})), nil
	}

	// Check if registry is configured
	cfg, found := c.dockerClientCfg.Registries[registry]
	if found && cfg.Username != "" && cfg.Password != "" {
		return authn.FromConfig(authn.AuthConfig{
			Username: cfg.Username,
			Password: cfg.Password,
		}), nil
	}

	// Fallback to anonymous access
	return authn.Anonymous, nil
}

// isNotFoundError checks if the error indicates the image was not found
func isNotFoundError(err error) bool {
	if terr, ok := err.(*transport.Error); ok {
		return terr.StatusCode == 404 // HTTP 404 Not Found
	}
	return false
}
