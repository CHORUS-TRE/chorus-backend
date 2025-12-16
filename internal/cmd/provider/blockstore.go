package provider

import (
	"context"
	"fmt"
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/client/blockstore"
	"github.com/CHORUS-TRE/chorus-backend/internal/client/diskblockstore"
	"github.com/CHORUS-TRE/chorus-backend/internal/client/minio"
	miniorawclient "github.com/CHORUS-TRE/chorus-backend/internal/client/minio/raw-client"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
)

var blockStoresOnce sync.Once
var blockStores map[string]blockstore.BlockStore

// ProvideBlockStores initializes and returns a map of block stores.
// Each block store is configured based on the configuration in the config file.
// Supported types are "minio" and "disk".
func ProvideBlockStores() map[string]blockstore.BlockStore {
	blockStoresOnce.Do(func() {
		config := ProvideConfig()
		blockStores = make(map[string]blockstore.BlockStore)

		for blockStoreName, blockStoreCfg := range config.Clients.BlockStores {
			switch blockStoreCfg.Type {
			case "minio":
				var minioClient miniorawclient.MinioClienter
				if !blockStoreCfg.MinioConfig.Enabled {
					logger.TechLog.Info(context.Background(), fmt.Sprintf("Minio block store '%s' is disabled, using test client", blockStoreName))
					minioClient = miniorawclient.NewTestClient()
				} else {
					clientCfg := miniorawclient.GetMinioClientConfigFromBlockStore(blockStoreName, blockStoreCfg.MinioConfig)
					var err error
					minioClient, err = miniorawclient.NewClient(clientCfg)
					if err != nil {
						logger.TechLog.Fatal(context.Background(), fmt.Sprintf("unable to create minio client for block store '%s': '%v'", blockStoreName, err))
					}
				}

				fileStore, err := minio.NewMinioFileStorage(minioClient)
				if err != nil {
					logger.TechLog.Fatal(context.Background(), fmt.Sprintf("failed to create minio block store '%s': %v", blockStoreName, err))
				}
				blockStores[blockStoreName] = fileStore

			case "disk":
				if blockStoreCfg.DiskConfig.BasePath == "" {
					logger.TechLog.Fatal(context.Background(), fmt.Sprintf("disk_config.base_path is required for disk block store: %s", blockStoreName))
				}
				fileStore, err := diskblockstore.NewDiskFileStorage(blockStoreCfg.DiskConfig.BasePath)
				if err != nil {
					logger.TechLog.Fatal(context.Background(), fmt.Sprintf("failed to create disk block store '%s': %v", blockStoreName, err))
				}
				blockStores[blockStoreName] = fileStore

			default:
				logger.TechLog.Fatal(context.Background(), fmt.Sprintf("unsupported block store type '%s' for block store: %s", blockStoreCfg.Type, blockStoreName))
			}

			logger.TechLog.Info(context.Background(), fmt.Sprintf("Initialized block store '%s' of type '%s'", blockStoreName, blockStoreCfg.Type))
		}
	})
	return blockStores
}
