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
	ImageExists(imageRef string) (bool, error)
	GetLabels(imageRef string) (map[string]string, error)
	Credentials() (username, password string)
	Host() string
}

type registryConfig struct {
	host     string // bare hostname, e.g. "harbor.example.com"
	username string
	password string
}

type client struct {
	cfg config.Config
	reg registryConfig
}

func NewClient(cfg config.Config) (*client, error) {
	ociCfg := cfg.Clients.OCIClient
	return &client{
		cfg: cfg,
		reg: registryConfig{
			host:     ociCfg.Host,
			username: ociCfg.Username,
			password: ociCfg.Password.PlainText(),
		},
	}, nil
}

func (c *client) Host() string {
	return c.reg.host
}

func (c *client) Credentials() (string, string) {
	return c.reg.username, c.reg.password
}

func (c *client) ImageExists(imageRef string) (bool, error) {
	ref, err := name.ParseReference(imageRef, name.WeakValidation)
	if err != nil {
		return false, fmt.Errorf("invalid image reference: %w", err)
	}

	_, err = remote.Get(ref, remote.WithAuth(c.getRegistryAuth()))
	if err != nil {
		if terr, ok := err.(*transport.Error); ok && terr.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("failed to check if image exists: %w", err)
	}

	return true, nil
}

func (c *client) GetLabels(imageRef string) (map[string]string, error) {
	ref, err := name.ParseReference(imageRef, name.WeakValidation)
	if err != nil {
		return nil, fmt.Errorf("invalid image reference: %w", err)
	}

	desc, err := remote.Get(ref, remote.WithAuth(c.getRegistryAuth()))
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

func (c *client) getRegistryAuth() authn.Authenticator {
	if c.reg.username == "" || c.reg.password == "" {
		return authn.Anonymous
	}
	return authn.FromConfig(authn.AuthConfig{Username: c.reg.username, Password: c.reg.password})
}
