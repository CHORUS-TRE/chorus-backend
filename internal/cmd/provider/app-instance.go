package provider

import (
	"context"
	"sync"

	v1 "github.com/CHORUS-TRE/chorus-backend/internal/api/v1"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	ctrl_mw "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/middleware"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/migration"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app-instance/service"
	service_mw "github.com/CHORUS-TRE/chorus-backend/pkg/app-instance/service/middleware"
	store_mw "github.com/CHORUS-TRE/chorus-backend/pkg/app-instance/store/middleware"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app-instance/store/postgres"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
)

var appInstanceOnce sync.Once
var appInstance service.AppInstanceer

func ProvideAppInstance() service.AppInstanceer {
	appInstanceOnce.Do(func() {
		appInstance = service.NewAppInstanceService(
			ProvideAppInstanceStore(),
			ProvideHelmClient(),
			ProvideAppService(),
		)
		appInstance = service_mw.Logging(logger.BizLog)(appInstance)
		appInstance = service_mw.Validation(ProvideValidator())(appInstance)
		appInstance = service_mw.AppInstanceCaching(logger.TechLog)(appInstance)
	})
	return appInstance
}

var appInstanceControllerOnce sync.Once
var appInstanceController chorus.AppInstanceServiceServer

func ProvideAppInstanceController() chorus.AppInstanceServiceServer {
	appInstanceControllerOnce.Do(func() {
		appInstanceController = v1.NewAppInstanceController(ProvideAppInstance())
		appInstanceController = ctrl_mw.AppInstanceAuthorizing(logger.SecLog, []string{user_model.RoleAuthenticated.String()})(appInstanceController)
	})
	return appInstanceController
}

var appInstanceStoreOnce sync.Once
var appInstanceStore service.AppInstanceStore

func ProvideAppInstanceStore() service.AppInstanceStore {
	appInstanceStoreOnce.Do(func() {
		db := ProvideMainDB(WithClient("appInstance-store"), WithMigrations(migration.GetMigration))
		switch db.Type {
		case POSTGRES:
			appInstanceStore = postgres.NewAppInstanceStorage(db.DB.GetSqlxDB())
		default:
			logger.TechLog.Fatal(context.Background(), "unsupported database type: "+db.Type)
		}
		appInstanceStore = store_mw.Logging(logger.TechLog)(appInstanceStore)
	})
	return appInstanceStore
}
