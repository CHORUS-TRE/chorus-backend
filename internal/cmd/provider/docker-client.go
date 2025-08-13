package provider

import (
	"context"
	"fmt"
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/docker"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

var dockerClientOnce sync.Once
var dockerClient docker.DockerClienter

func ProvideDockerClient() docker.DockerClienter {
	dockerClientOnce.Do(func() {
		cfg := ProvideConfig()
		if !cfg.Clients.DockerClient.Enabled {
			logger.TechLog.Info(context.Background(), "Docker client is disabled, using test client")
			dockerClient = docker.NewTestClient()
		} else {
			var err error
			dockerClient, err = docker.NewClient(cfg)
			if err != nil {
				logger.TechLog.Fatal(context.Background(), fmt.Sprintf("unable to provide k8s client: '%v'", err))
			}
		}
	})
	return dockerClient
}
