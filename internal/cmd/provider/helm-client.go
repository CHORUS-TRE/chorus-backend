package provider

import (
	"context"
	"fmt"
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/helm"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

var helmClientOnce sync.Once
var helmClient helm.HelmClienter

func ProvideHelmClient() helm.HelmClienter {
	helmClientOnce.Do(func() {
		cfg := ProvideConfig()
		if cfg.Clients.HelmClient.KubeConfig == "" && cfg.Clients.HelmClient.Token == "" {
			helmClient = helm.NewTestClient()
		} else {
			var err error
			helmClient, err = helm.NewClient(cfg)
			if err != nil {
				logger.TechLog.Fatal(context.Background(), fmt.Sprintf("unable to provide helm cleint: '%v'", err))
			}
		}
	})
	return helmClient
}
