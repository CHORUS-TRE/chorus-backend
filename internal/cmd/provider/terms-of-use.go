package provider

import (
	"context"
	"sync"

	v1 "github.com/CHORUS-TRE/chorus-backend/internal/api/v1"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	ctrl_mw "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/middleware"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/migration"
	"github.com/CHORUS-TRE/chorus-backend/pkg/terms-of-use/service"
	service_mw "github.com/CHORUS-TRE/chorus-backend/pkg/terms-of-use/service/middleware"
	"github.com/CHORUS-TRE/chorus-backend/pkg/terms-of-use/store/postgres"
)

var termsOfUseControllerOnce sync.Once
var termsOfUseController chorus.TermsOfUseServiceServer

func ProvideTermsOfUseController() chorus.TermsOfUseServiceServer {
	termsOfUseControllerOnce.Do(func() {
		termsOfUseController = v1.NewTermsOfUseController(ProvideTermsOfUseService())
		termsOfUseController = ctrl_mw.TermsOfUseAuthorizing(logger.SecLog, ProvideAuthorizer())(termsOfUseController)
		if ProvideConfig().Services.AuditService.Enabled {
			termsOfUseController = ctrl_mw.NewTermsOfUseAuditMiddleware(ProvideAuditWriter())(termsOfUseController)
		}
	})
	return termsOfUseController
}

var termsOfUseServiceOnce sync.Once
var termsOfUseService service.TermsOfUseer

func ProvideTermsOfUseService() service.TermsOfUseer {
	termsOfUseServiceOnce.Do(func() {
		termsOfUseService = service.NewTermsOfUseService(ProvideTermsOfUseStore())
		termsOfUseService = service_mw.Logging(logger.BizLog)(termsOfUseService)
		termsOfUseService = service_mw.Validation()(termsOfUseService)
		termsOfUseService = service_mw.TermsOfUseCaching(logger.TechLog)(termsOfUseService)
	})
	return termsOfUseService
}

var termsOfUseStoreOnce sync.Once
var termsOfUseStore service.TermsOfUseStore

func ProvideTermsOfUseStore() service.TermsOfUseStore {
	termsOfUseStoreOnce.Do(func() {
		db := ProvideMainDB(WithClient("terms-of-use-store"), WithMigrations(migration.GetMigration))
		switch db.Type {
		case POSTGRES:
			termsOfUseStore = postgres.NewTermsOfUseStorage(db.DB.GetSqlxDB())
		default:
			logger.TechLog.Fatal(context.Background(), "unsupported database type: "+db.Type)
		}
	})
	return termsOfUseStore
}
