package provider

import (
	"context"
	"sync"

	v1 "github.com/CHORUS-TRE/chorus-backend/internal/api/v1"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	ctrl_mw "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/middleware"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/migration"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app/service"
	service_mw "github.com/CHORUS-TRE/chorus-backend/pkg/app/service/middleware"
	store_mw "github.com/CHORUS-TRE/chorus-backend/pkg/app/store/middleware"
	"github.com/CHORUS-TRE/chorus-backend/pkg/app/store/postgres"
	user_model "github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
)

var appOnce sync.Once
var app service.Apper

func ProvideAppService() service.Apper {
	appOnce.Do(func() {
		app = service.NewAppService(
			ProvideAppStore(),
		)
		app = service_mw.Logging(logger.BizLog)(app)
		app = service_mw.Validation(ProvideValidator())(app)
		app = service_mw.AppCaching(logger.TechLog)(app)
	})
	return app
}

var appControllerOnce sync.Once
var appController chorus.AppServiceServer

func ProvideAppController() chorus.AppServiceServer {
	appControllerOnce.Do(func() {
		appController = v1.NewAppController(ProvideAppService())
		appController = ctrl_mw.AppAuthorizing(logger.SecLog, []string{user_model.RoleAuthenticated.String()})(appController)
	})
	return appController
}

var appStoreOnce sync.Once
var appStore service.AppStore

func ProvideAppStore() service.AppStore {
	appStoreOnce.Do(func() {
		db := ProvideMainDB(WithClient("app-store"), WithMigrations(migration.GetMigration))
		switch db.Type {
		case POSTGRES:
			appStore = postgres.NewAppStorage(db.DB.GetSqlxDB())
		default:
			logger.TechLog.Fatal(context.Background(), "unsupported database type: "+db.Type)
		}
		appStore = store_mw.Logging(logger.TechLog)(appStore)
	})
	return appStore
}
