package provider

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/zap"

	v1 "github.com/CHORUS-TRE/chorus-backend/internal/api/v1"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	ctrl_mw "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/middleware"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/internal/migration"
	grpc_mw "github.com/CHORUS-TRE/chorus-backend/internal/protocol/grpc/middleware"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authentication/service"
	service_mw "github.com/CHORUS-TRE/chorus-backend/pkg/authentication/service/middleware"
	store_mw "github.com/CHORUS-TRE/chorus-backend/pkg/authentication/store/middleware"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authentication/store/postgres"
)

var authenticatorOnce sync.Once
var authenticator service.Authenticator

func ProvideAuthenticator() service.Authenticator {
	cfg := ProvideConfig()

	authenticatorOnce.Do(func() {
		authService, err := service.NewAuthenticationService(cfg, ProvideUser(), ProvideAuthenticationStore(), ProvideDaemonEncryptionKey())
		if err != nil {
			logger.TechLog.Fatal(context.Background(), fmt.Sprintf("unable to provide authentication service: %v", err))
		}
		authenticator = service_mw.Logging(logger.SecLog)(authService)
	})
	return authenticator
}

var authenticationControllerOnce sync.Once
var authenticationController chorus.AuthenticationServiceServer

func ProvideAuthenticationController() chorus.AuthenticationServiceServer {
	authenticationControllerOnce.Do(func() {
		authenticationController = v1.NewAuthenticationController(ProvideAuthenticator(), ProvideAuthorizer(), ProvideConfig())
		if ProvideConfig().Services.AuditService.Enabled {
			authenticationController = ctrl_mw.NewAuthenticationAuditMiddleware(ProvideAuditWriter())(authenticationController)
		}
	})
	return authenticationController
}

var authenticationStoreOnce sync.Once
var authenticationStore service.AuthenticationStore

func ProvideAuthenticationStore() service.AuthenticationStore {
	authenticationStoreOnce.Do(func() {
		db := ProvideMainDB(WithClient("authentication-store"), WithMigrations(migration.GetMigration))
		switch db.Type {
		case POSTGRES:
			authenticationStore = postgres.NewAuthenticationStorage(db.DB.GetSqlxDB())
		default:
			logger.TechLog.Fatal(context.Background(), "unsupported database type: "+db.Type)
		}
		authenticationStore = store_mw.Logging(logger.TechLog)(authenticationStore)
	})
	return authenticationStore
}

var clientWhitelisterOnce sync.Once
var clientWhitelister grpc_mw.ClientWhitelister

func ProvideClientWhitelister() grpc_mw.ClientWhitelister {
	clientWhitelisterOnce.Do(func() {
		var err error
		clientWhitelister, err = grpc_mw.NewIPWhitelister(ProvideConfig())
		if err != nil {
			logger.TechLog.Logger.Fatal("unable to create client whitelister", zap.Error(err))
		}
	})
	return clientWhitelister
}
