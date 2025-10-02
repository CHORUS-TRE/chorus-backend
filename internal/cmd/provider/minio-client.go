package provider

import (
	"context"
	"fmt"
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/minio"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

var minioClientOnce sync.Once
var minioClient minio.MinioClienter

func ProvideMinioClient() minio.MinioClienter {
	minioClientOnce.Do(func() {
		cfg := ProvideConfig()
		if !cfg.Clients.MinioClient.Enabled {
			logger.TechLog.Info(context.Background(), "Minio client is disabled, using test client")
			minioClient = minio.NewTestClient()
		} else {
			var err error
			minioClient, err = minio.NewClient(cfg)
			if err != nil {
				logger.TechLog.Fatal(context.Background(), fmt.Sprintf("unable to provide minio client: '%v'", err))
			}
		}
	})
	return minioClient
}
