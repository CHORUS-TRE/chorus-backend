package provider

import (
	"context"
	"sync"

	v1 "github.com/CHORUS-TRE/chorus-backend/internal/api/v1"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	ctrl_mw "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/middleware"
	"github.com/CHORUS-TRE/chorus-backend/internal/client/filestore"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/approval-request/service"
	service_mw "github.com/CHORUS-TRE/chorus-backend/pkg/approval-request/service/middleware"
)

var approvalRequestServiceOnce sync.Once
var approvalRequestService service.ApprovalRequester

func ProvideApprovalRequestService() service.ApprovalRequester {
	approvalRequestServiceOnce.Do(func() {
		cfg := ProvideConfig()

		var workspacePrefix string
		for _, storeCfg := range cfg.Services.WorkspaceFileService.Stores {
			workspacePrefix = storeCfg.WorkspacePrefix
			break
		}

		approvalRequestService = service.NewApprovalRequestService(
			ProvideApprovalRequestStore(),
			ProvideApprovalRequestSourceFileStore(),
			ProvideApprovalRequestStagingFileStore(),
			workspacePrefix,
		)
		approvalRequestService = service_mw.Logging(logger.BizLog)(approvalRequestService)
		approvalRequestService = service_mw.Validation(ProvideValidator())(approvalRequestService)
	})
	return approvalRequestService
}

var approvalRequestControllerOnce sync.Once
var approvalRequestController chorus.ApprovalRequestServiceServer

func ProvideApprovalRequestController() chorus.ApprovalRequestServiceServer {
	approvalRequestControllerOnce.Do(func() {
		approvalRequestController = v1.NewApprovalRequestController(ProvideApprovalRequestService())
		approvalRequestController = ctrl_mw.ApprovalRequestAuthorizing(logger.SecLog, ProvideAuthorizer(), ProvideConfig(), ProvideAuthenticator())(approvalRequestController)
	})
	return approvalRequestController
}

var approvalRequestStoreOnce sync.Once
var approvalRequestStore service.ApprovalRequestStore

func ProvideApprovalRequestStore() service.ApprovalRequestStore {
	approvalRequestStoreOnce.Do(func() {
		// TODO: Implement approval request store (postgres)
		// db := ProvideMainDB(WithClient("approval-request-store"), WithMigrations(migration.GetMigration))
		// switch db.Type {
		// case POSTGRES:
		// 	approvalRequestStore = postgres.NewApprovalRequestStorage(db.DB.GetSqlxDB())
		// default:
		// 	logger.TechLog.Fatal(context.Background(), "unsupported database type: "+db.Type)
		// }
		// approvalRequestStore = store_mw.Logging(logger.TechLog)(approvalRequestStore)
		logger.TechLog.Warn(context.Background(), "approval request store not yet implemented")
	})
	return approvalRequestStore
}

var approvalRequestSourceFileStoreOnce sync.Once
var approvalRequestSourceFileStore filestore.FileStore

func ProvideApprovalRequestSourceFileStore() filestore.FileStore {
	approvalRequestSourceFileStoreOnce.Do(func() {
		fileStores := ProvideFileStores()
		for _, fs := range fileStores {
			approvalRequestSourceFileStore = fs
			break
		}
		if approvalRequestSourceFileStore == nil {
			logger.TechLog.Fatal(context.Background(), "no file store available for approval request source files")
		}
	})
	return approvalRequestSourceFileStore
}

var approvalRequestStagingFileStoreOnce sync.Once
var approvalRequestStagingFileStore filestore.FileStore

func ProvideApprovalRequestStagingFileStore() filestore.FileStore {
	approvalRequestStagingFileStoreOnce.Do(func() {
		fileStores := ProvideFileStores()
		for _, fs := range fileStores {
			approvalRequestStagingFileStore = fs
			break
		}
		if approvalRequestStagingFileStore == nil {
			logger.TechLog.Fatal(context.Background(), "no file store available for approval request staging files")
		}
	})
	return approvalRequestStagingFileStore
}
