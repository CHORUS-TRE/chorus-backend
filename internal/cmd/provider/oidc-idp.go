package provider

import (
	"context"
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/migration"
	"github.com/CHORUS-TRE/chorus-backend/pkg/oidc-idp/service"
	"github.com/CHORUS-TRE/chorus-backend/pkg/oidc-idp/store/middleware"
	"github.com/CHORUS-TRE/chorus-backend/pkg/oidc-idp/store/postgres"
	"github.com/luikyv/go-oidc/pkg/goidc"
	"go.uber.org/zap"
)

var oidcidpServiceOnce sync.Once
var oidcidpService service.OIDCProviderService

func ProvideOIDCIDPService() service.OIDCProviderService {
	oidcidpServiceOnce.Do(func() {
		var err error
		oidcidpService, err = service.NewOIDCProviderService(ProvideConfig(), ProvideAuthorizer(), ProvideOIDCIDPAuthnSessionManager(), ProvideOIDCIDPClientManager(), ProvideOIDCIDPLogoutSessionManager(), ProvideOIDCIDPGrantSessionManager(), ProvideUser())
		if err != nil {
			logger.TechLog.Fatal(context.Background(), "unable to instantiate OIDC IDP service", zap.Error(err))
		}
		// oidcidpService = service.Logging(logger.BizLog)(oidcidpService)
		// oidcidpService = service.Validation(ProvideValidator())(oidcidpService)
	})
	return oidcidpService
}

var clientManagerOnce sync.Once
var clientManager goidc.ClientManager

func ProvideOIDCIDPClientManager() goidc.ClientManager {
	clientManagerOnce.Do(func() {
		var err error
		clientManager, err = service.NewClientManager(ProvideConfig())
		if err != nil {
			logger.TechLog.Fatal(context.Background(), "unable to instantiate OIDC IDP client manager", zap.Error(err))
		}
	})
	return clientManager
}

var authnSessionManagerOnce sync.Once
var authnSessionManager goidc.AuthnSessionManager

func ProvideOIDCIDPAuthnSessionManager() goidc.AuthnSessionManager {
	authnSessionManagerOnce.Do(func() {
		db := ProvideMainDB(WithClient("default-store"), WithMigrations(migration.GetMigration))
		switch db.Type {
		case POSTGRES:
			authnSessionManager = postgres.NewAuthnSessionManager(db.DB.GetSqlxDB())
		default:
			logger.TechLog.Fatal(context.Background(), "unsupported database type: "+db.Type)
		}
		authnSessionManager = middleware.AuthnLogging(logger.BizLog)(authnSessionManager)
	})
	return authnSessionManager
}

var logoutSessionManagerOnce sync.Once
var logoutSessionManager goidc.LogoutSessionManager

func ProvideOIDCIDPLogoutSessionManager() goidc.LogoutSessionManager {
	logoutSessionManagerOnce.Do(func() {
		db := ProvideMainDB(WithClient("default-store"), WithMigrations(migration.GetMigration))
		switch db.Type {
		case POSTGRES:
			logoutSessionManager = postgres.NewLogoutSessionManager(db.DB.GetSqlxDB())
		default:
			logger.TechLog.Fatal(context.Background(), "unsupported database type: "+db.Type)
		}
		logoutSessionManager = middleware.LogoutLogging(logger.BizLog)(logoutSessionManager)
	})
	return logoutSessionManager
}

var grantSessionManagerOnce sync.Once
var grantSessionManager goidc.GrantSessionManager

func ProvideOIDCIDPGrantSessionManager() goidc.GrantSessionManager {
	grantSessionManagerOnce.Do(func() {
		db := ProvideMainDB(WithClient("default-store"), WithMigrations(migration.GetMigration))
		switch db.Type {
		case POSTGRES:
			grantSessionManager = postgres.NewGrantSessionManager(db.DB.GetSqlxDB())
		default:
			logger.TechLog.Fatal(context.Background(), "unsupported database type: "+db.Type)
		}
		grantSessionManager = middleware.GrantLogging(logger.BizLog)(grantSessionManager)
	})
	return grantSessionManager
}
