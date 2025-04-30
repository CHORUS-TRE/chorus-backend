package provider

import (
	"context"
	"fmt"
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/k8s"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

var k8sClientOnce sync.Once
var k8sClient k8s.K8sClienter

func ProvideK8sClient() k8s.K8sClienter {
	k8sClientOnce.Do(func() {
		cfg := ProvideConfig()
		if cfg.Clients.K8sClient.KubeConfig == "" && cfg.Clients.K8sClient.Token == "" {
			k8sClient = k8s.NewTestClient()
		} else {
			var err error
			k8sClient, err = k8s.NewClient(cfg)
			if err != nil {
				logger.TechLog.Fatal(context.Background(), fmt.Sprintf("unable to provide k8s client: '%v'", err))
			}
		}
	})
	return k8sClient
}
