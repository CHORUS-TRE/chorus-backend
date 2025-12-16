package provider

import (
	"context"
	"fmt"
	"sync"

	v1 "github.com/CHORUS-TRE/chorus-backend/internal/api/v1"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	ctrl_mw "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/middleware"
	"github.com/CHORUS-TRE/chorus-backend/internal/client/diskblockstore"
	"github.com/CHORUS-TRE/chorus-backend/internal/client/minio"
	miniorawclient "github.com/CHORUS-TRE/chorus-backend/internal/client/minio/raw-client"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace-file/service"
	service_mw "github.com/CHORUS-TRE/chorus-backend/pkg/workspace-file/service/middleware"
)

var workspaceFileOnce sync.Once
var workspaceFile service.WorkspaceFiler

func ProvideWorkspaceFile() service.WorkspaceFiler {
	workspaceFileOnce.Do(func() {
		var err error
		workspaceFile, err = service.NewWorkspaceFileService(
			ProvideWorkspaceFileStores(),
			ProvideConfig().Services.WorkspaceFileService.Stores,
		)
		if err != nil {
			logger.TechLog.Fatal(context.Background(), "failed to create workspace file service: "+err.Error())
		}
		workspaceFile = service_mw.Logging(logger.BizLog)(workspaceFile)
		workspaceFile = service_mw.Validation(ProvideValidator())(workspaceFile)
		workspaceFile = service_mw.WorkspaceCaching(logger.TechLog)(workspaceFile)
	})
	return workspaceFile
}

var workspaceFileControllerOnce sync.Once
var workspaceFileController chorus.WorkspaceFileServiceServer

func ProvideWorkspaceFileController() chorus.WorkspaceFileServiceServer {
	workspaceFileControllerOnce.Do(func() {
		workspaceFileController = v1.NewWorkspaceFileController(ProvideWorkspaceFile())
		workspaceFileController = ctrl_mw.WorkspaceFileAuthorizing(logger.SecLog, ProvideAuthorizer(), ProvideConfig(), ProvideAuthenticator())(workspaceFileController)
	})
	return workspaceFileController
}

var workspaceFileStoresOnce sync.Once
var workspaceFileStores map[string]service.WorkspaceFileStore

func ProvideWorkspaceFileStores() map[string]service.WorkspaceFileStore {
	workspaceFileStoresOnce.Do(func() {
		config := ProvideConfig()
		workspaceFileStores = make(map[string]service.WorkspaceFileStore)

		for storeName, storeCfg := range config.Services.WorkspaceFileService.Stores {
			blockStoreName := storeCfg.BlockStoreName
			blockStore, ok := config.Clients.BlockStores[blockStoreName]
			if !ok {
				logger.TechLog.Fatal(context.Background(), "block store not found: "+blockStoreName+" for workspace file store: "+storeName)
			}

			switch blockStore.Type {
			case "minio":
				if blockStore.MinioConfig == nil {
					logger.TechLog.Fatal(context.Background(), "minio_config is required for minio block store type: "+blockStoreName)
				}

				var minioClient miniorawclient.MinioClienter
				if !blockStore.MinioConfig.Enabled {
					logger.TechLog.Info(context.Background(), fmt.Sprintf("Minio block store '%s' is disabled, using test client", blockStoreName))
					minioClient = miniorawclient.NewTestClient()
				} else {
					clientCfg := miniorawclient.GetMinioClientConfigFromBlockStore(blockStoreName, blockStore.MinioConfig)
					var err error
					minioClient, err = miniorawclient.NewClient(clientCfg)
					if err != nil {
						logger.TechLog.Fatal(context.Background(), fmt.Sprintf("unable to create minio client for block store '%s': '%v'", blockStoreName, err))
					}
				}

				fileStore, err := minio.NewMinioFileStorage(minioClient)
				if err != nil {
					logger.TechLog.Fatal(context.Background(), "failed to create minio workspace file store: "+err.Error())
				}
				workspaceFileStores[storeName] = fileStore

			case "disk":
				if blockStore.DiskConfig == nil {
					logger.TechLog.Fatal(context.Background(), "disk_config is required for disk block store type: "+blockStoreName)
				}
				if blockStore.DiskConfig.BasePath == "" {
					logger.TechLog.Fatal(context.Background(), "disk_config.base_path is required for disk block store: "+blockStoreName)
				}
				fileStore, err := diskblockstore.NewDiskFileStorage(blockStore.DiskConfig.BasePath)
				if err != nil {
					logger.TechLog.Fatal(context.Background(), "failed to create disk workspace file store: "+err.Error())
				}
				workspaceFileStores[storeName] = fileStore

			default:
				logger.TechLog.Fatal(context.Background(), "unsupported block store type: "+blockStore.Type+" for block store: "+blockStoreName)
			}
		}
	})
	return workspaceFileStores
}
