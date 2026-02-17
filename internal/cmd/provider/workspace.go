package provider

import (
	"context"
	"sync"

	v1 "github.com/CHORUS-TRE/chorus-backend/internal/api/v1"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	ctrl_mw "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/middleware"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/migration"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/service"
	service_mw "github.com/CHORUS-TRE/chorus-backend/pkg/workspace/service/middleware"
	store_mw "github.com/CHORUS-TRE/chorus-backend/pkg/workspace/store/middleware"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workspace/store/postgres"
)

var workspaceControllerOnce sync.Once
var workspaceController chorus.WorkspaceServiceServer

func ProvideWorkspaceController() chorus.WorkspaceServiceServer {
	workspaceControllerOnce.Do(func() {
		workspaceController = v1.NewWorkspaceController(ProvideWorkspaceService())
		workspaceController = ctrl_mw.WorkspaceAuthorizing(logger.SecLog, ProvideAuthorizer(), ProvideConfig(), ProvideAuthenticator())(workspaceController)
		if ProvideConfig().Services.AuditService.Enabled {
			workspaceController = ctrl_mw.NewWorkspaceAuditMiddleware(ProvideAuditWriter())(workspaceController)
		}
	})
	return workspaceController
}

var workspaceServiceOnce sync.Once
var workspaceService service.Workspaceer

func ProvideWorkspaceService() service.Workspaceer {
	workspaceServiceOnce.Do(func() {
		workspaceService = service.NewWorkspaceService(
			ProvideConfig(),
			ProvideWorkspaceStore(),
			ProvideK8sClient(),
			ProvideWorkbench(),
			ProvideUser(),
			ProvideNotificationStore(),
		)
		workspaceService = service_mw.Logging(logger.BizLog)(workspaceService)
		workspaceService = service_mw.Validation(ProvideValidator())(workspaceService)
		workspaceService = service_mw.WorkspaceCaching(logger.TechLog)(workspaceService)
	})
	return workspaceService
}

var workspaceStoreOnce sync.Once
var workspaceStore service.WorkspaceStore

func ProvideWorkspaceStore() service.WorkspaceStore {
	workspaceStoreOnce.Do(func() {
		db := ProvideMainDB(WithClient("workspace-store"), WithMigrations(migration.GetMigration))
		switch db.Type {
		case POSTGRES:
			workspaceStore = postgres.NewWorkspaceStorage(db.DB.GetSqlxDB())
		default:
			logger.TechLog.Fatal(context.Background(), "unsupported database type: "+db.Type)
		}
		workspaceStore = store_mw.Logging(logger.TechLog)(workspaceStore)
	})
	return workspaceStore
}
