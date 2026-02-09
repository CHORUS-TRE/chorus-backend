package provider

import (
	"context"
	"fmt"
	"sync"

	v1 "github.com/CHORUS-TRE/chorus-backend/internal/api/v1"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/chorus"
	ctrl_mw "github.com/CHORUS-TRE/chorus-backend/internal/api/v1/middleware"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
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

		cfg := ProvideConfig()
		if cfg.Services.AuthorizationService.WorkspaceAdminCanAssignDataManager {
			var permissionManagerUserDataRole *gatekeeper_model.Permission
			for i, perm := range schema.Permissions {
				if perm.Name == model.PermissionManageUsersDataRoleInWorkspace.String() {
					permissionManagerUserDataRole = &schema.Permissions[i]
					break
				}
			}
			if permissionManagerUserDataRole == nil {
				logger.TechLog.Fatal(context.Background(), fmt.Sprintf("permission %s not found in schema", model.PermissionManageUsersDataRoleInWorkspace.String()))
			}
			for i, role := range schema.Roles {
				if role.Name == model.RoleWorkspaceAdmin.String() {
					schema.Roles[i].Permissions = append(schema.Roles[i].Permissions, *permissionManagerUserDataRole)
					break
				}
			}
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

var authorizationControllerOnce sync.Once
var authorizationController chorus.AuthorizationServiceServer

func ProvideAuthorizationController() chorus.AuthorizationServiceServer {
	authorizationControllerOnce.Do(func() {
		authorizationController = v1.NewAuthorizationController(ProvideGatekeeper())
		authorizationController = ctrl_mw.AuthorizationAuthorizing(logger.SecLog, ProvideAuthorizer(), ProvideConfig(), ProvideAuthenticator())(authorizationController)
	})
	return authorizationController
}
