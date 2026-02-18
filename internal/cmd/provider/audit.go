package provider

import (
	"context"
	"sync"

	v1 "github.com/CHORUS-TRE/chorus-backend/internal/api/v1"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/migration"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/service"
	service_mw "github.com/CHORUS-TRE/chorus-backend/pkg/audit/service/middleware"
	store_mw "github.com/CHORUS-TRE/chorus-backend/pkg/audit/store/middleware"
	"github.com/CHORUS-TRE/chorus-backend/pkg/audit/store/postgres"
)

var auditControllerOnce sync.Once
var auditController chorus.AuditServiceServer

func ProvideAuditController() chorus.AuditServiceServer {
	auditControllerOnce.Do(func() {
		auditController = v1.NewAuditController(ProvideAuditService())
	})
	return auditController
}

var auditServiceOnce sync.Once
var auditService service.Auditer

func ProvideAuditService() service.Auditer {
	auditServiceOnce.Do(func() {
		cfg := ProvideConfig()

		if !cfg.Services.AuditService.Enabled {
			logger.TechLog.Info(context.Background(), "Audit service is disabled")
			auditService = service.NewNoOpAuditer()
			return
		}

		auditStore := ProvideAuditStore()
		auditService = service.NewAuditService(auditStore)

		auditService = service_mw.Logging(logger.BizLog)(auditService)
	})
	return auditService
}

var auditStoreOnce sync.Once
var auditStore service.AuditStore

func ProvideAuditStore() service.AuditStore {
	auditStoreOnce.Do(func() {
		db := ProvideAuditDB(WithClient("audit-store"), WithMigrations(migration.GetAuditMigration))
		switch db.Type {
		case POSTGRES:
			auditStore = postgres.NewAuditStorage(db.DB.GetSqlxDB())
		default:
			logger.TechLog.Fatal(context.Background(), "unsupported database type for audit service: "+db.Type)
		}

		auditStore = store_mw.Logging(logger.TechLog)(auditStore)
	})
	return auditStore
}

// ProvideAuditWriter returns the audit service as an AuditWriter for middleware
func ProvideAuditWriter() service.AuditWriter {
	return ProvideAuditService()
}
