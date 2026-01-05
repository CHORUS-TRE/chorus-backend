package provider

import (
	"context"
	"fmt"
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/diskfilestore"
	"github.com/CHORUS-TRE/chorus-backend/internal/client/filestore"
	"github.com/CHORUS-TRE/chorus-backend/internal/client/minio"
	miniorawclient "github.com/CHORUS-TRE/chorus-backend/internal/client/minio/raw-client"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

var fileStoresOnce sync.Once
var fileStores map[string]filestore.FileStore

// ProvideFileStores initializes and returns a map of file stores.
// Each file store is configured based on the configuration in the config file.
// Supported types are "minio" and "disk".
func ProvideFileStores() map[string]filestore.FileStore {
	fileStoresOnce.Do(func() {
		config := ProvideConfig()
		fileStores = make(map[string]filestore.FileStore)

		for fileStoreName, fileStoreCfg := range config.Clients.FileStores {
			switch fileStoreCfg.Type {
			case "minio":
				var minioClient miniorawclient.MinioClienter
				if !fileStoreCfg.MinioConfig.Enabled {
					logger.TechLog.Info(context.Background(), fmt.Sprintf("Minio file store '%s' is disabled, using test client", fileStoreName))
					minioClient = miniorawclient.NewTestClient()
				} else {
					clientCfg := miniorawclient.GetMinioClientConfigFromFileStore(fileStoreName, fileStoreCfg.MinioConfig)
					var err error
					minioClient, err = miniorawclient.NewClient(clientCfg)
					if err != nil {
						logger.TechLog.Fatal(context.Background(), fmt.Sprintf("unable to create minio client for file store '%s': '%v'", fileStoreName, err))
					}
				}

				fileStore, err := minio.NewMinioFileStorage(minioClient)
				if err != nil {
					logger.TechLog.Fatal(context.Background(), fmt.Sprintf("failed to create minio file store '%s': %v", fileStoreName, err))
				}
				fileStores[fileStoreName] = fileStore

			case "disk":
				if fileStoreCfg.DiskConfig.BasePath == "" {
					logger.TechLog.Fatal(context.Background(), fmt.Sprintf("disk_config.base_path is required for disk file store: %s", fileStoreName))
				}
				fileStore, err := diskfilestore.NewDiskFileStorage(fileStoreCfg.DiskConfig.BasePath)
				if err != nil {
					logger.TechLog.Fatal(context.Background(), fmt.Sprintf("failed to create disk file store '%s': %v", fileStoreName, err))
				}
				fileStores[fileStoreName] = fileStore

			default:
				logger.TechLog.Fatal(context.Background(), fmt.Sprintf("unsupported file store type '%s' for file store: %s", fileStoreCfg.Type, fileStoreName))
			}

			logger.TechLog.Info(context.Background(), fmt.Sprintf("Initialized file store '%s' of type '%s'", fileStoreName, fileStoreCfg.Type))
		}
	})
	return fileStores
}
