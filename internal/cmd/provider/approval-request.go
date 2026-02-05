package provider

import (
	"context"
	"fmt"
	"sync"

	v1 "github.com/CHORUS-TRE/chorus-backend/internal/api/v1"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	ctrl_mw "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/middleware"
	"github.com/CHORUS-TRE/chorus-backend/internal/client/filestore"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/migration"
	"github.com/CHORUS-TRE/chorus-backend/pkg/approval-request/service"
	service_mw "github.com/CHORUS-TRE/chorus-backend/pkg/approval-request/service/middleware"
	store_mw "github.com/CHORUS-TRE/chorus-backend/pkg/approval-request/store/middleware"
	"github.com/CHORUS-TRE/chorus-backend/pkg/approval-request/store/postgres"
)

var approvalRequestServiceOnce sync.Once
var approvalRequestService service.ApprovalRequester

func ProvideApprovalRequestService() service.ApprovalRequester {
	approvalRequestServiceOnce.Do(func() {
		cfg := ProvideConfig()

		approvalRequestConfig := service.ApprovalRequestConfig{
			RequireDataManagerApproval: cfg.Services.ApprovalRequestService.RequireDataManagerApproval,
		}

		approvalRequestService = service.NewApprovalRequestService(
			ProvideApprovalRequestStore(),
			ProvideWorkspaceFile(),
			ProvideApprovalRequestStagingFileStore(cfg.Services.ApprovalRequestService.StagingFileStoreName),
			ProvideNotificationStore(),
			ProvideAuthorizer(),
			approvalRequestConfig,
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
		approvalRequestController = ctrl_mw.ApprovalRequestAuthorizing(logger.SecLog, ProvideAuthorizer(), ProvideConfig(), ProvideAuthenticator(), ProvideApprovalRequestStore())(approvalRequestController)
	})
	return approvalRequestController
}

var approvalRequestStoreOnce sync.Once
var approvalRequestStore service.ApprovalRequestStore

func ProvideApprovalRequestStore() service.ApprovalRequestStore {
	approvalRequestStoreOnce.Do(func() {
		db := ProvideMainDB(WithClient("approval-request-store"), WithMigrations(migration.GetMigration))
		switch db.Type {
		case POSTGRES:
			approvalRequestStore = postgres.NewApprovalRequestStorage(db.DB.GetSqlxDB())
		default:
			logger.TechLog.Fatal(context.Background(), "unsupported database type: "+db.Type)
		}
		approvalRequestStore = store_mw.Logging(logger.TechLog)(approvalRequestStore)
		logger.TechLog.Warn(context.Background(), "approval request store not yet implemented")
	})
	return approvalRequestStore
}

var approvalRequestStagingFileStoreOnce sync.Once
var approvalRequestStagingFileStore filestore.FileStore

func ProvideApprovalRequestStagingFileStore(stagingStoreName string) filestore.FileStore {
	approvalRequestStagingFileStoreOnce.Do(func() {
		fileStores := ProvideFileStores()
		if stagingStoreName == "" {
			logger.TechLog.Fatal(context.Background(), "staging file store not specified: "+stagingStoreName)
		}

		if store, ok := fileStores[stagingStoreName]; ok {
			approvalRequestStagingFileStore = store
			fmt.Println("Using staging file store:", stagingStoreName)
		} else {
			logger.TechLog.Fatal(context.Background(), "staging file store not found: "+stagingStoreName)
		}

		if approvalRequestStagingFileStore == nil {
			logger.TechLog.Fatal(context.Background(), "no file store available for approval request staging files")
		}
	})
	return approvalRequestStagingFileStore
}
