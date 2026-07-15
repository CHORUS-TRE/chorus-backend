package provider

import (
	"context"
	"sync"

	v1 "github.com/CHORUS-TRE/chorus-backend/internal/api/v1"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	ctrl_mw "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/middleware"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/migration"
	"github.com/CHORUS-TRE/chorus-backend/pkg/platform-settings/service"
	service_mw "github.com/CHORUS-TRE/chorus-backend/pkg/platform-settings/service/middleware"
	store_mw "github.com/CHORUS-TRE/chorus-backend/pkg/platform-settings/store/middleware"
	"github.com/CHORUS-TRE/chorus-backend/pkg/platform-settings/store/postgres"
)

var platformSettingsOnce sync.Once
var platformSettings service.PlatformSettingser

func ProvidePlatformSettings() service.PlatformSettingser {
	platformSettingsOnce.Do(func() {
		platformSettings = service.NewPlatformSettingsService(ProvidePlatformSettingsStore())
		platformSettings = service_mw.Logging(logger.BizLog)(platformSettings)
		platformSettings = service_mw.Validation()(platformSettings)
		platformSettings = service_mw.PlatformSettingsCaching(logger.TechLog)(platformSettings)
	})
	return platformSettings
}

var platformSettingsControllerOnce sync.Once
var platformSettingsController chorus.PlatformSettingsServiceServer

func ProvidePlatformSettingsController() chorus.PlatformSettingsServiceServer {
	platformSettingsControllerOnce.Do(func() {
		platformSettingsController = v1.NewPlatformSettingsController(ProvidePlatformSettings())
		platformSettingsController = ctrl_mw.PlatformSettingsAuthorizing(logger.SecLog, ProvideAuthorizer())(platformSettingsController)
		if ProvideConfig().Services.AuditService.Enabled {
			platformSettingsController = ctrl_mw.NewPlatformSettingsAuditMiddleware(ProvideAuditWriter())(platformSettingsController)
		}
	})
	return platformSettingsController
}

var platformSettingsStoreOnce sync.Once
var platformSettingsStore service.PlatformSettingsStore

func ProvidePlatformSettingsStore() service.PlatformSettingsStore {
	platformSettingsStoreOnce.Do(func() {
		db := ProvideMainDB(WithClient("platform-settings-store"), WithMigrations(migration.GetMigration))
		switch db.Type {
		case POSTGRES:
			platformSettingsStore = postgres.NewPlatformSettingsStorage(db.DB.GetSqlxDB())
		default:
			logger.TechLog.Fatal(context.Background(), "unsupported database type: "+db.Type)
		}
		platformSettingsStore = store_mw.Logging(logger.TechLog)(platformSettingsStore)
	})
	return platformSettingsStore
}
