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
	"go.uber.org/zap"
)

var authorizationPolicyOnce sync.Once
var authorizationPolicy authorization_service.AuthorizationServiceInterface

func ProvideAuthorizationPolicy() authorization_service.AuthorizationServiceInterface {
	authorizationPolicyOnce.Do(func() {
		schema := model.GetDefaultSchema()

		cfg := ProvideConfig()
		if cfg.Services.AuthorizationService.WorkspaceAdminCanAssignDataManager {
			var permissionManagerUserDataRole *model.PermissionName
			for i, perm := range schema.Permissions {
				if perm.Name == model.PermissionManageUsersDataRoleInWorkspace {
					permissionManagerUserDataRole = &schema.Permissions[i].Name
					break
				}
			}
			if permissionManagerUserDataRole == nil {
				logger.TechLog.Fatal(context.Background(), fmt.Sprintf("permission %s not found in schema", model.PermissionManageUsersDataRoleInWorkspace.String()))
			}
			for i, role := range schema.Roles {
				if role.Name == model.RoleWorkspaceAdmin {
					schema.Roles[i].Permissions = append(schema.Roles[i].Permissions, *permissionManagerUserDataRole)
					break
				}
			}
		}

		dynamicRoleStore := authorization_store.NewDynamicRoleStorage(ProvideMainDB().GetSqlxDB())
		dynamicRoles, loadErr := dynamicRoleStore.ListDynamicRoles(context.Background())
		if loadErr != nil {
			logger.TechLog.Fatal(context.Background(), "failed to load dynamic authorization roles", zap.Error(loadErr))
		}
		schema.Roles = append(schema.Roles, dynamicRoles...)

		var err error
		authorizationPolicy, err = authorization_service.NewAuthorizationService(&schema)
		if err != nil {
			logger.TechLog.Fatal(context.Background(), "failed to create authorization policy", zap.Error(err))
		}
		authorizationPolicy.SetDynamicRoleStore(dynamicRoleStore)
	})
	return authorizationPolicy
}

var userPermissionStoreOnce sync.Once
var userPermissionStore authorization_service.UserPermissionStore

func ProvideUserPermissionStore() authorization_service.UserPermissionStore {
	userPermissionStoreOnce.Do(func() {
		authStructures, err := authorization_service.ExtractAuthorizationStructures(ProvideAuthorizationPolicy())
		if err != nil {
			logger.TechLog.Fatal(context.Background(), "failed to extract authorization structures", zap.Error(err))
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
		authorizer, err = authorization_service.NewAuthorizer(ProvideAuthorizationPolicy(), ProvideUserPermissionStore())
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
		authorizationController = v1.NewAuthorizationController(ProvideAuthorizationPolicy())
		authorizationController = ctrl_mw.AuthorizationAuthorizing(logger.SecLog, ProvideAuthorizer(), ProvideConfig(), ProvideAuthenticator())(authorizationController)
	})
	return authorizationController
}
