package provider

import (
	"context"
	"fmt"
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/minio"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

var minioClientsOnce sync.Once
var minioClients map[string]minio.MinioClienter

func ProvideMinioClients() map[string]minio.MinioClienter {
	minioClientsOnce.Do(func() {
		cfg := ProvideConfig()
		minioClients = make(map[string]minio.MinioClienter)

		for clientName := range cfg.Services.WorkspaceFileService.MinioStores {
			var minioClient minio.MinioClienter
			if !cfg.Services.WorkspaceFileService.MinioStores[clientName].Enabled {
				logger.TechLog.Info(context.Background(), fmt.Sprintf("Minio client '%s' is disabled, using test client", clientName))
				minioClient = minio.NewTestClient()
			} else {
				var err error
				minioClient, err = minio.NewClient(cfg, clientName)
				if err != nil {
					logger.TechLog.Fatal(context.Background(), fmt.Sprintf("unable to provide minio client '%s': '%v'", clientName, err))
				}
			}
			minioClients[clientName] = minioClient
		}
	})
	return minioClients
}
