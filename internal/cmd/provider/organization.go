package provider

import (
	"context"
	"sync"

	v1 "github.com/CHORUS-TRE/chorus-backend/internal/api/v1"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	ctrl_mw "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/middleware"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/migration"
	"github.com/CHORUS-TRE/chorus-backend/pkg/organization/service"
	service_mw "github.com/CHORUS-TRE/chorus-backend/pkg/organization/service/middleware"
	store_mw "github.com/CHORUS-TRE/chorus-backend/pkg/organization/store/middleware"
	"github.com/CHORUS-TRE/chorus-backend/pkg/organization/store/postgres"
)

var organizationControllerOnce sync.Once
var organizationController chorus.OrganizationServiceServer

func ProvideOrganizationController() chorus.OrganizationServiceServer {
	organizationControllerOnce.Do(func() {
		organizationController = v1.NewOrganizationController(ProvideOrganizationService())
		organizationController = ctrl_mw.OrganizationAuthorizing(logger.SecLog, ProvideAuthorizer())(organizationController)
		if ProvideConfig().Services.AuditService.Enabled {
			organizationController = ctrl_mw.NewOrganizationAuditMiddleware(ProvideAuditWriter())(organizationController)
		}
	})
	return organizationController
}

var organizationServiceOnce sync.Once
var organizationService service.Organizationer

func ProvideOrganizationService() service.Organizationer {
	organizationServiceOnce.Do(func() {
		organizationService = service.NewOrganizationService(ProvideOrganizationStore())
		organizationService = service_mw.Logging(logger.BizLog)(organizationService)
		organizationService = service_mw.Validation(ProvideValidator())(organizationService)
		organizationService = service_mw.OrganizationCaching(logger.TechLog)(organizationService)
	})
	return organizationService
}

var organizationStoreOnce sync.Once
var organizationStore service.OrganizationStore

func ProvideOrganizationStore() service.OrganizationStore {
	organizationStoreOnce.Do(func() {
		db := ProvideMainDB(WithClient("organization-store"), WithMigrations(migration.GetMigration))
		switch db.Type {
		case POSTGRES:
			organizationStore = postgres.NewOrganizationStorage(db.DB.GetSqlxDB())
		default:
			logger.TechLog.Fatal(context.Background(), "unsupported database type: "+db.Type)
		}
		organizationStore = store_mw.Logging(logger.TechLog)(organizationStore)
	})
	return organizationStore
}
