package provider

import (
	"context"
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/migration"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/tenant/store/postgres"

	"github.com/CHORUS-TRE/chorus-backend/pkg/tenant/service"
	service_mw "github.com/CHORUS-TRE/chorus-backend/pkg/tenant/service/middleware"
)

var tenantStoreOnce sync.Once
var tenantStore service.TenantStore

func ProvideTenantStore() service.TenantStore {
	tenantStoreOnce.Do(func() {
		db := ProvideMainDB(WithClient("default-store"), WithMigrations(migration.GetMigration))
		switch db.Type {
		case POSTGRES:
			tenantStore = postgres.NewTenantStorage(db.DB.GetSqlxDB())
		default:
			logger.TechLog.Fatal(context.Background(), "unsupported database type: "+db.Type)
		}
	})
	return tenantStore
}

var tenanterOnce sync.Once
var tenanter service.Tenanter

func ProvideTenanter() service.Tenanter {
	tenanterOnce.Do(func() {

		tenanter = service.NewTenantService(
			ProvideTenantStore(),
			ProvideConfig(),
		)
		tenanter = service_mw.Logging(logger.BizLog)(tenanter)
		tenanter = service_mw.Validation(ProvideValidator())(tenanter)
		tenanter = service_mw.TenantCaching(logger.TechLog)(tenanter)
	})
	return tenanter
}
