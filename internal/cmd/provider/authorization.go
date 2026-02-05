package provider

import (
	"context"
	"sync"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	authorization_service "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"
	authorization_store_middleware "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/store/middleware"
	authorization_store "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/store/postgres"
	gatekeeper_model "github.com/CHORUS-TRE/chorus-gatekeeper/pkg/authorization/model"
	gatekeeper_service "github.com/CHORUS-TRE/chorus-gatekeeper/pkg/authorization/service"
	"go.uber.org/zap"
)

var gatekeeperOnce sync.Once
var gatekeeper gatekeeper_service.AuthorizationServiceInterface

func ProvideGatekeeper() gatekeeper_service.AuthorizationServiceInterface {
	gatekeeperOnce.Do(func() {
		schema, err := gatekeeper_model.GetDefaultSchema()
		if err != nil {
			logger.TechLog.Fatal(context.Background(), "failed to get default gatekeeper model schema", zap.Error(err))
		}
		gatekeeper, err = gatekeeper_service.NewAuthorizationService(&schema)
	})
	return gatekeeper
}

var userPermissionStoreOnce sync.Once
var userPermissionStore authorization_service.UserPermissionStore

func ProvideUserPermissionStore() authorization_service.UserPermissionStore {
	userPermissionStoreOnce.Do(func() {
		authStructures, err := authorization_service.ExtractAuthoizationStructures(ProvideGatekeeper())
		if err != nil {
			logger.TechLog.Fatal(context.Background(), "failed to extract authorization structures from gatekeeper", zap.Error(err))
		}
		store := authorization_store.NewUserPermissionStorage(ProvideMainDB().GetSqlxDB(), authStructures.RolesGrantingPermission)
		userPermissionStore = authorization_store_middleware.UserPermissionLogging(logger.TechLog)(store)
	})
	return userPermissionStore
}

var authorizerOnce sync.Once
var authorizer authorization_service.Authorizer

func ProvideAuthorizer() authorization_service.Authorizer {
	authorizerOnce.Do(func() {
		var err error
		authorizer, err = authorization_service.NewAuthorizer(ProvideGatekeeper(), ProvideUserPermissionStore())
		if err != nil {
			logger.TechLog.Fatal(context.Background(), "failed to create authorizer", zap.Error(err))
		}
	})
	return authorizer
}
