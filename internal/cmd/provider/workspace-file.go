package provider

import (
	"context"
	"sync"

	v1 "github.com/CHORUS-TRE/chorus-backend/internal/api/v1"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	ctrl_mw "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/middleware"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace-file/service"
	service_mw "github.com/CHORUS-TRE/chorus-backend/pkg/workspace-file/service/middleware"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace-file/store/minio"
)

var workspaceFileOnce sync.Once
var workspaceFile service.WorkspaceFiler

func ProvideWorkspaceFile() service.WorkspaceFiler {
	workspaceFileOnce.Do(func() {
		workspaceFile = service.NewWorkspaceFileService(
			ProvideWorkspaceFileStores(),
		)
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
		minioClients := ProvideMinioClients()
		workspaceFileStores = make(map[string]service.WorkspaceFileStore)

		// Minio file stores
		for storeName, minioClient := range minioClients {
			minioStore, err := minio.NewMinioFileStorage(storeName, minioClient)
			if err != nil {
				logger.TechLog.Fatal(context.Background(), "failed to create minio file store: "+err.Error())
			}
			workspaceFileStores[storeName] = minioStore
		}

		// Additional file stores can be added here
	})
	return workspaceFileStores
}
