package provider

import (
	"context"
	"sync"

	v1 "github.com/CHORUS-TRE/chorus-backend/internal/api/v1"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	ctrl_mw "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/middleware"
	"github.com/CHORUS-TRE/chorus-backend/internal/client/filestore"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/request/service"
	service_mw "github.com/CHORUS-TRE/chorus-backend/pkg/request/service/middleware"
)

var requestServiceOnce sync.Once
var requestService service.Requester

func ProvideRequestService() service.Requester {
	requestServiceOnce.Do(func() {
		cfg := ProvideConfig()

		var workspacePrefix string
		for _, storeCfg := range cfg.Services.WorkspaceFileService.Stores {
			workspacePrefix = storeCfg.WorkspacePrefix
			break
		}

		requestService = service.NewRequestService(
			ProvideRequestStore(),
			ProvideRequestSourceFileStore(),
			ProvideRequestStagingFileStore(),
			workspacePrefix,
		)
		requestService = service_mw.Logging(logger.BizLog)(requestService)
		requestService = service_mw.Validation(ProvideValidator())(requestService)
	})
	return requestService
}

var requestControllerOnce sync.Once
var requestController chorus.RequestServiceServer

func ProvideRequestController() chorus.RequestServiceServer {
	requestControllerOnce.Do(func() {
		requestController = v1.NewRequestController(ProvideRequestService())
		requestController = ctrl_mw.RequestAuthorizing(logger.SecLog, ProvideAuthorizer(), ProvideConfig(), ProvideAuthenticator())(requestController)
	})
	return requestController
}

var requestStoreOnce sync.Once
var requestStore service.RequestStore

func ProvideRequestStore() service.RequestStore {
	requestStoreOnce.Do(func() {
		// TODO: Implement request store (postgres)
		// db := ProvideMainDB(WithClient("request-store"), WithMigrations(migration.GetMigration))
		// switch db.Type {
		// case POSTGRES:
		// 	requestStore = postgres.NewRequestStorage(db.DB.GetSqlxDB())
		// default:
		// 	logger.TechLog.Fatal(context.Background(), "unsupported database type: "+db.Type)
		// }
		// requestStore = store_mw.Logging(logger.TechLog)(requestStore)
		logger.TechLog.Warn(context.Background(), "request store not yet implemented")
	})
	return requestStore
}

var requestSourceFileStoreOnce sync.Once
var requestSourceFileStore filestore.FileStore

func ProvideRequestSourceFileStore() filestore.FileStore {
	requestSourceFileStoreOnce.Do(func() {
		fileStores := ProvideFileStores()
		for _, fs := range fileStores {
			requestSourceFileStore = fs
			break
		}
		if requestSourceFileStore == nil {
			logger.TechLog.Fatal(context.Background(), "no file store available for request source files")
		}
	})
	return requestSourceFileStore
}

var requestStagingFileStoreOnce sync.Once
var requestStagingFileStore filestore.FileStore

func ProvideRequestStagingFileStore() filestore.FileStore {
	requestStagingFileStoreOnce.Do(func() {
		fileStores := ProvideFileStores()
		for _, fs := range fileStores {
			requestStagingFileStore = fs
			break
		}
		if requestStagingFileStore == nil {
			logger.TechLog.Fatal(context.Background(), "no file store available for request staging files")
		}
	})
	return requestStagingFileStore
}
