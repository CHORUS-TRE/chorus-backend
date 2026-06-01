package middleware

import (
	"context"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	"github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"
	"go.uber.org/zap"
)

var _ service.Authorizer = (*authorizationServiceLogging)(nil)

type authorizationServiceLogging struct {
	logger *logger.ContextLogger
	next   service.Authorizer
}

func Logging(logger *logger.ContextLogger) func(service.Authorizer) service.Authorizer {
	return func(next service.Authorizer) service.Authorizer {
		return &authorizationServiceLogging{
			logger: logger,
			next:   next,
		}
	}
}

func (c authorizationServiceLogging) GetAuthorizationSchema() *model.AuthorizationSchema {
	now := time.Now()

	schema := c.next.GetAuthorizationSchema()

	c.logger.Info(context.Background(), "request completed",
		zap.Bool("schema_exists", schema != nil),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return schema
}

func (c authorizationServiceLogging) CreateDynamicRole(ctx context.Context, user []model.Role, role *model.RoleDefinition, validationContext model.Context) (*model.RoleDefinition, error) {
	now := time.Now()

	createdRole, err := c.next.CreateDynamicRole(ctx, user, role, validationContext)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Any("role", role),
			zap.Any("validation_context", validationContext),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Info(ctx, "request completed",
		zap.Any("created_role", createdRole),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return createdRole, nil
}

func (c authorizationServiceLogging) IsUserAllowed(user []model.Role, permission model.Permission) (bool, error) {
	now := time.Now()

	allowed, err := c.next.IsUserAllowed(user, permission)
	if err != nil {
		c.logger.Error(context.Background(), logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return false, err
	}

	c.logger.Info(context.Background(), "request completed",
		zap.Bool("allowed", allowed),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return allowed, nil
}

func (c authorizationServiceLogging) ExplainIsUserAllowed(user []model.Role, permission model.Permission) string {
	now := time.Now()

	explanation := c.next.ExplainIsUserAllowed(user, permission)

	c.logger.Info(context.Background(), "request completed",
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return explanation
}

func (c authorizationServiceLogging) GetUserPermissions(user []model.Role) ([]model.Permission, error) {
	now := time.Now()

	permissions, err := c.next.GetUserPermissions(user)
	if err != nil {
		c.logger.Error(context.Background(), logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Info(context.Background(), "request completed",
		zap.Int("num_permissions", len(permissions)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return permissions, nil
}

func (c authorizationServiceLogging) GetContextListForPermission(user []model.Role, permissionName model.PermissionName) ([]model.Context, error) {
	now := time.Now()

	contextList, err := c.next.GetContextListForPermission(user, permissionName)
	if err != nil {
		c.logger.Error(context.Background(), logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Info(context.Background(), "request completed",
		zap.Int("num_contexts", len(contextList)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return contextList, nil
}

func (c authorizationServiceLogging) FindUsersWithPermission(ctx context.Context, tenantID uint64, filter model.FindUsersWithPermissionFilter) ([]uint64, error) {
	now := time.Now()

	userIDs, err := c.next.FindUsersWithPermission(ctx, tenantID, filter)
	if err != nil {
		c.logger.Error(ctx, logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Uint64("tenant_id", tenantID),
			zap.Any("filter", filter),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return nil, err
	}

	c.logger.Info(ctx, "request completed",
		zap.Int("num_user_ids", len(userIDs)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return userIDs, nil
}

func (c authorizationServiceLogging) GetRolesGrantingPermission(permissionName model.PermissionName) []model.RoleName {
	now := time.Now()

	roleNames := c.next.GetRolesGrantingPermission(permissionName)

	c.logger.Info(context.Background(), "request completed",
		zap.Int("num_role_names", len(roleNames)),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return roleNames
}

func (c authorizationServiceLogging) GetRoleDefinition(roleName model.RoleName) (*model.RoleDefinition, bool) {
	now := time.Now()

	roleDef, found := c.next.GetRoleDefinition(roleName)

	c.logger.Info(context.Background(), "request completed",
		zap.Bool("found", found),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return roleDef, found
}

func (c authorizationServiceLogging) IsRoleInScope(roleName model.RoleName, scopes ...model.RoleScope) bool {
	now := time.Now()

	inScope := c.next.IsRoleInScope(roleName, scopes...)

	c.logger.Info(context.Background(), "request completed",
		zap.Bool("in_scope", inScope),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return inScope
}

func (c authorizationServiceLogging) CanAssignRole(user []model.Role, roleName model.RoleName, assignmentContext model.Context) (bool, error) {
	now := time.Now()

	allowed, err := c.next.CanAssignRole(user, roleName, assignmentContext)
	if err != nil {
		c.logger.Error(context.Background(), logger.LoggerMessageRequestFailed,
			zap.Error(err),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
		)
		return false, err
	}

	c.logger.Info(context.Background(), "request completed",
		zap.Bool("allowed", allowed),
		zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(now).Nanoseconds())/1000000.0),
	)
	return allowed, nil
}
