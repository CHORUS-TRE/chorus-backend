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
	authorization_store "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/store/postgres"
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
var authorizationStore authorization_service.Store

func ProvideAuthorizationStore() authorization_service.Store {
	authorizationStoreOnce.Do(func() {
		authorizationStore = authorization_store.NewRoleStorage(ProvideMainDB().GetSqlxDB())
	})
	return authorizationStore
}
