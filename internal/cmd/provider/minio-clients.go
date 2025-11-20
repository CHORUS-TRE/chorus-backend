package provider

import (
	"context"
	"fmt"
	"sync"

	miniorawclient "github.com/CHORUS-TRE/chorus-backend/internal/client/minio/raw-client"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

var minioClientsOnce sync.Once
var minioClients map[string]miniorawclient.MinioClienter

func ProvideMinioClients() map[string]miniorawclient.MinioClienter {
	minioClientsOnce.Do(func() {
		cfg := ProvideConfig()
		minioClients = make(map[string]miniorawclient.MinioClienter)

		for clientName, _ := range cfg.Clients.MinioClients {
			var minioClient miniorawclient.MinioClienter
			if !cfg.Clients.MinioClients[clientName].Enabled {
				logger.TechLog.Info(context.Background(), fmt.Sprintf("Minio client '%s' is disabled, using test client", clientName))
				minioClient = miniorawclient.NewTestClient()
			} else {
				var err error
				minioClient, err = miniorawclient.NewClient(cfg, clientName)
				if err != nil {
					logger.TechLog.Fatal(context.Background(), fmt.Sprintf("unable to provide minio client '%s': '%v'", clientName, err))
				}
			}
			minioClients[clientName] = minioClient
		}
	})
	return minioClients
}
