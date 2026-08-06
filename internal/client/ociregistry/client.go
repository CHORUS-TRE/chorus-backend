package ociregistry

import (
	"context"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	"go.uber.org/zap"
)

var _ OCIClienter = &client{}

type OCIClienter interface {
	ImageExists(imageRef string, username string, password string) (bool, error)
	GetLabels(imageRef string, username string, password string) (map[string]string, error)
}

type client struct {
	cfg       config.Config
	clientCfg ClientConfig
}

func NewClient(cfg config.Config) (*client, error) {
	clientCfg, err := getClientConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("error getting oci client config: %w", err)
	}

	return &client{
		cfg:       cfg,
		clientCfg: clientCfg,
	}, nil
}

// ImageExists checks whether an image exists in registry
func (c *client) ImageExists(imageRef string, username string, password string) (bool, error) {
	// Parse image reference
	ref, err := name.ParseReference(imageRef, name.WeakValidation)
	if err != nil {
		return false, fmt.Errorf("invalid image reference: %w", err)
	}

	registry := ref.Context().RegistryStr()
	authenticator, err := c.getRegistryAuth(registry, username, password)
	if err != nil {
		return false, fmt.Errorf("failed to get registry auth: %w", err)
	}

	_, err = remote.Get(ref, remote.WithAuth(authenticator))
	if err != nil {
		if terr, ok := err.(*transport.Error); ok && terr.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if image exists: %w", err)
	}

	return true, nil
}

// GetLabels retrieves the OCI image config labels for the given image reference.
func (c *client) GetLabels(imageRef string, username, password string) (map[string]string, error) {
	ref, err := name.ParseReference(imageRef, name.WeakValidation)
	if err != nil {
		return nil, fmt.Errorf("invalid image reference: %w", err)
	}

	registry := ref.Context().RegistryStr()
	authenticator, err := c.getRegistryAuth(registry, username, password)
	if err != nil {
		return nil, fmt.Errorf("failed to get registry auth: %w", err)
	}

	desc, err := remote.Get(ref, remote.WithAuth(authenticator))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch image descriptor: %w", err)
	}

	img, err := desc.Image()
	if err != nil {
		return nil, fmt.Errorf("failed to get image from descriptor: %w", err)
	}

	cfgFile, err := img.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("failed to get image config: %w", err)
	}

	logger.TechLog.Debug(context.Background(), "fetched image config labels", zap.String("imageRef", imageRef), zap.Int64("nb_labels", int64(len(cfgFile.Config.Labels))))

	return cfgFile.Config.Labels, nil
}

func (c *client) getRegistryAuth(registry string, username string, password string) (authn.Authenticator, error) {
	if registry == "" {
		return nil, fmt.Errorf("registry hostname cannot be empty")
	}

	// If credentials are provided, use them
	if username != "" && password != "" {
		return authn.FromConfig((authn.AuthConfig{
			Username: username,
			Password: password,
		})), nil
	}

	// Check if registry is configured
	cfg, found := c.clientCfg.Registries[registry]
	if found && cfg.Username != "" && cfg.Password != "" {
		return authn.FromConfig(authn.AuthConfig{
			Username: cfg.Username,
			Password: cfg.Password,
		}), nil
	}

	// Fallback to anonymous access
	return authn.Anonymous, nil
}
