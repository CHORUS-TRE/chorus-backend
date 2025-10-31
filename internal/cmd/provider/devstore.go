package provider

import (
	"context"
	"sync"

	v1 "github.com/CHORUS-TRE/chorus-backend/internal/api/v1"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	ctrl_mw "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/middleware"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/migration"
	"github.com/CHORUS-TRE/chorus-backend/pkg/devstore/service"
	service_mw "github.com/CHORUS-TRE/chorus-backend/pkg/devstore/service/middleware"
	"github.com/CHORUS-TRE/chorus-backend/pkg/devstore/store/postgres"
)

var devstoreOnce sync.Once
var devstore service.Devstorer

func ProvideDevstore() service.Devstorer {
	devstoreOnce.Do(func() {
		devstore = service.NewDevstoreService(
			ProvideConfig(),
			ProvideDevstoreStore(),
		)

		devstore = service_mw.Logging(logger.BizLog)(devstore)
		devstore = service_mw.Validation(ProvideValidator())(devstore)
	})

	return devstore
}

var devstoreControllerOnce sync.Once
var devstoreController chorus.DevstoreServiceServer

func ProvideDevstoreController() chorus.DevstoreServiceServer {
	devstoreControllerOnce.Do(func() {
		devstoreController = v1.NewDevstoreController(ProvideDevstore())
		devstoreController = ctrl_mw.DevstoreAuthorizing(logger.SecLog, ProvideAuthorizer(), ProvideConfig(), ProvideAuthenticator())(devstoreController)
	})
	return devstoreController
}

var devstoreStoreOnce sync.Once
var devstoreStore service.DevstoreStore

func ProvideDevstoreStore() service.DevstoreStore {
	devstoreStoreOnce.Do(func() {
		db := ProvideMainDB(WithClient("devstore-store"), WithMigrations(migration.GetMigration))
		switch db.Type {
		case POSTGRES:
			devstoreStore = postgres.NewDevstoreStorage(db.DB.GetSqlxDB())
		default:
			logger.TechLog.Fatal(context.Background(), "unsupported database type: "+db.Type)
		}
	})
	return devstoreStore
}
