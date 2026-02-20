package provider

import (
	"context"
	"sync"

	v1 "github.com/CHORUS-TRE/chorus-backend/internal/api/v1"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	ctrl_mw "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/middleware"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/migration"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/service"
	service_mw "github.com/CHORUS-TRE/chorus-backend/pkg/workbench/service/middleware"
	store_mw "github.com/CHORUS-TRE/chorus-backend/pkg/workbench/store/middleware"
	"github.com/CHORUS-TRE/chorus-backend/pkg/workbench/store/postgres"
)

var workbenchOnce sync.Once
var workbench service.Workbencher

func ProvideWorkbench() service.Workbencher {
	workbenchOnce.Do(func() {
		workbench = service.NewWorkbenchService(
			ProvideConfig(),
			ProvideWorkbenchStore(),
			ProvideK8sClient(),
			ProvideAppService(),
			ProvideUser(),
			ProvideAuthenticator(),
			ProvideNotificationStore(),
		)
		workbench = service_mw.Logging(logger.BizLog)(workbench)
		workbench = service_mw.Validation(ProvideValidator())(workbench)
		workbench = service_mw.WorkbenchCaching(logger.TechLog)(workbench)
	})
	return workbench
}

var workbenchControllerOnce sync.Once
var workbenchController chorus.WorkbenchServiceServer

func ProvideWorkbenchController() chorus.WorkbenchServiceServer {
	workbenchControllerOnce.Do(func() {
		workbenchController = v1.NewWorkbenchController(ProvideWorkbench(), ProvideAuditService())
		workbenchController = ctrl_mw.WorkbenchAuthorizing(logger.SecLog, ProvideAuthorizer(), ProvideConfig(), ProvideAuthenticator())(workbenchController)
		if ProvideConfig().Services.AuditService.Enabled {
			workbenchController = ctrl_mw.NewWorkbenchAuditMiddleware(ProvideAuditWriter())(workbenchController)
		}
	})
	return workbenchController
}

var workbenchStoreOnce sync.Once
var workbenchStore service.WorkbenchStore

func ProvideWorkbenchStore() service.WorkbenchStore {
	workbenchStoreOnce.Do(func() {
		db := ProvideMainDB(WithClient("workbench-store"), WithMigrations(migration.GetMigration))
		switch db.Type {
		case POSTGRES:
			workbenchStore = postgres.NewWorkbenchStorage(db.DB.GetSqlxDB())
		default:
			logger.TechLog.Fatal(context.Background(), "unsupported database type: "+db.Type)
		}
		workbenchStore = store_mw.Logging(logger.TechLog)(workbenchStore)
	})
	return workbenchStore
}
