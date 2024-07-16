package provider

import (
	"context"
	"sync"

	v1 "github.com/CHORUS-TRE/chorus-backend/internal/api/v1"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	ctrl_mw "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/middleware"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/migration"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/service"
	service_mw "github.com/CHORUS-TRE/chorus-backend/pkg/user/service/middleware"
	store_mw "github.com/CHORUS-TRE/chorus-backend/pkg/user/store/middleware"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/store/postgres"
)

var userOnce sync.Once
var user service.Userer

func ProvideUser() service.Userer {
	cfg := ProvideConfig()

	userOnce.Do(func() {
		user = service.NewUserService(
			cfg.Daemon.TOTP.NumRecoveryCodes,
			ProvideDaemonEncryptionKey(),
			ProvideUserStore(),
			ProvideMailer(),
		)
		user = service_mw.Logging(logger.BizLog)(user)
		user = service_mw.Validation(ProvideValidator())(user)
		user = service_mw.UserCaching(logger.TechLog)(user)
	})
	return user
}

var userControllerOnce sync.Once
var userController chorus.UserServiceServer

func ProvideUserController() chorus.UserServiceServer {
	userControllerOnce.Do(func() {
		userController = v1.NewUserController(ProvideUser())
		userController = ctrl_mw.UserAuthorizing(logger.SecLog, []string{model.RoleAdmin.String()})(userController)
	})
	return userController
}

var userStoreOnce sync.Once
var userStore service.UserStore

func ProvideUserStore() service.UserStore {
	userStoreOnce.Do(func() {
		db := ProvideMainDB(WithClient("user-store"), WithMigrations(migration.GetMigration))
		switch db.Type {
		case POSTGRES:
			userStore = postgres.NewUserStorage(db.DB.GetSqlxDB())
		default:
			logger.TechLog.Fatal(context.Background(), "unsupported database type: "+db.Type)
		}
		userStore = store_mw.Logging(logger.TechLog)(userStore)
	})
	return userStore
}
