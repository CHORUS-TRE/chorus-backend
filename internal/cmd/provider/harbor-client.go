package provider

import (
	"context"
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/harbor"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

var harborClientOnce sync.Once
var harborClient harbor.HarborClient

func ProvideHarborClient() harbor.HarborClient {
	harborClientOnce.Do(func() {
		cfg := ProvideConfig()

		if !cfg.Clients.HarborClient.Enabled {
			logger.TechLog.Info(context.Background(), "Harbor client is disabled, using noop client")
			harborClient = &harbor.HarborNoopClient{}
			return
		}

		harborClient = harbor.NewHarborClient(cfg, ProvideDockerClient())
	})
	return harborClient
}
