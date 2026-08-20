package provider

import (
	"context"
	"sync"

	v1 "github.com/CHORUS-TRE/chorus-backend/internal/api/v1"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	ctrl_mw "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/middleware"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	authorization_service "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service/middleware"
	store_mw "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/store/middleware"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authorization/store/postgres"
	"go.uber.org/zap"
)

var authorizationPolicyOnce sync.Once
var authorizationPolicy authorization_service.Authorizer

func ProvideAuthorizer() authorization_service.Authorizer {
	authorizationPolicyOnce.Do(func() {
		var err error
		authorizationPolicy, err = authorization_service.NewAuthorizationService(context.Background(), ProvideConfig(), ProvideAuthorizationStore())
		if err != nil {
			logger.TechLog.Fatal(context.Background(), "failed to create authorization policy", zap.Error(err))
		}
		authorizationPolicy = middleware.Logging(logger.BizLog)(authorizationPolicy)
		authorizationPolicy = middleware.Validation(ProvideValidator())(authorizationPolicy)
		authorizationPolicy = middleware.AuthorizationCaching(logger.TechLog)(authorizationPolicy)
	})
	return authorizationPolicy
}

var authorizationControllerOnce sync.Once
var authorizationController chorus.AuthorizationServiceServer

func ProvideAuthorizationController() chorus.AuthorizationServiceServer {
	authorizationControllerOnce.Do(func() {
		authorizationController = v1.NewAuthorizationController(ProvideAuthorizer())
		authorizationController = ctrl_mw.AuthorizationAuthorizing(logger.SecLog, ProvideAuthorizer(), ProvideConfig(), ProvideAuthenticator())(authorizationController)
	})
	return authorizationController
}

var authorizationStoreOnce sync.Once
var authorizationStore authorization_service.AuthorizationStore

func ProvideAuthorizationStore() authorization_service.AuthorizationStore {
	authorizationStoreOnce.Do(func() {
		db := ProvideMainDB(WithClient("authorization-store"))
		switch db.Type {
		case POSTGRES:
			authorizationStore = postgres.NewAuthorizationStorage(db.DB.GetSqlxDB())
		default:
			logger.TechLog.Fatal(context.Background(), "unsupported database type for authorization store", zap.String("db_type", string(db.Type)))
		}
		authorizationStore = store_mw.Logging(logger.TechLog)(authorizationStore)
	})
	return authorizationStore
}
