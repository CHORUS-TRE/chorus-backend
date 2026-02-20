package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	authorization "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	authorization_service "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Refresher interface {
	RefreshToken(ctx context.Context) (string, time.Duration, error)
}

type Authorization struct {
	logger     *logger.ContextLogger
	authorizer authorization_service.Authorizer
	refresher  Refresher
	cfg        config.Config
	// authorizedRoles []string
}

func NewAuthorization(logger *logger.ContextLogger, cfg config.Config, authorizer authorization_service.Authorizer, refresher Refresher) Authorization {
	return Authorization{
		logger:     logger,
		authorizer: authorizer,
		refresher:  refresher,
		cfg:        cfg,
		// authorizedRoles: []string{},
	}
}

func (c Authorization) getRolesAndClaims(ctx context.Context) ([]authorization.Role, *jwt_model.JWTClaims, error) {

	claims, ok := ctx.Value(jwt_model.JWTClaimsContextKey).(*jwt_model.JWTClaims)
	if !ok {
		c.logger.Warn(ctx, "malformed JWT token", zap.Any("content", ctx.Value(jwt_model.JWTClaimsContextKey)))
		return nil, nil, status.Error(codes.Unauthenticated, "malformed jwt-token")
	}

	aRoles, err := claimRolesToAuthRoles(claims)
	if err != nil {
		c.logger.Error(ctx, "failed to convert claim roles to auth roles", zap.Error(err))
		return nil, nil, status.Error(codes.Internal, "failed to convert claim roles to auth roles")
	}

	return aRoles, claims, nil
}

func (c Authorization) getUserID(ctx context.Context) (uint64, error) {
	claims, ok := ctx.Value(jwt_model.JWTClaimsContextKey).(*jwt_model.JWTClaims)
	if !ok {
		c.logger.Warn(ctx, "malformed JWT token")
		return 0, status.Error(codes.Unauthenticated, "malformed jwt-token")
	}

	return claims.ID, nil
}

func (c Authorization) IsAuthorized(ctx context.Context, permissionName authorization.PermissionName, opts ...authorization.NewContextOption) error {
	permission := authorization.NewPermission(permissionName, opts...)

	aRoles, claims, err := c.getRolesAndClaims(ctx)
	if err != nil {
		c.logger.Error(ctx, "failed to get roles and claims", zap.Error(err))
		return status.Error(codes.Internal, "failed to get roles and claims")
	}

	isAuthorized, err := c.authorizer.IsUserAllowed(aRoles, permission)
	if err != nil {
		c.logger.Error(ctx, "authorization error", zap.Error(err))
		return status.Error(codes.Internal, "authorization error")
	}

	if !isAuthorized {
		return c.permissionDenied(ctx, claims, permission)
	}

	return nil
}

func (c Authorization) ExplainIsAuthorized(ctx context.Context, permissionName authorization.PermissionName, opts ...authorization.NewContextOption) string {
	permission := authorization.NewPermission(permissionName, opts...)

	aRoles, _, err := c.getRolesAndClaims(ctx)
	if err != nil {
		return fmt.Sprintf("error getting roles and claims: %v", err)
	}

	return c.authorizer.ExplainIsUserAllowed(aRoles, permission)

}

func (c Authorization) GetContextListForPermission(ctx context.Context, permissionName authorization.PermissionName) ([]authorization.Context, error) {
	aRoles, _, err := c.getRolesAndClaims(ctx)
	if err != nil {
		c.logger.Error(ctx, "failed to get roles and claims", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get roles and claims")
	}

	contextList, err := c.authorizer.GetContextListForPermission(aRoles, permissionName)
	if err != nil {
		c.logger.Error(ctx, "authorization error", zap.Error(err))
		return nil, status.Error(codes.Internal, "authorization error")
	}

	return contextList, nil
}

func (c Authorization) TriggerRefreshToken(ctx context.Context) error {
	res, t, err := c.refresher.RefreshToken(ctx)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "%v", err)
	}

	expiresDate := time.Now().Add(t)
	expires := expiresDate.Format(time.RFC1123)

	header := c.getSetCookieHeader(res, expires)
	if err := grpc.SetHeader(ctx, header); err != nil {
		return status.Errorf(codes.Internal, "%v", err)
	}

	return nil
}

func (c Authorization) getSetCookieHeader(token string, expires string) metadata.MD {
	return metadata.Pairs("Set-Cookie", "jwttoken="+token+"; Path=/; Domain="+c.cfg.Daemon.HTTP.Headers.CookieDomain+"; SameSite=None; Secure; HttpOnly; Expires="+expires)
}

func (c Authorization) permissionDenied(ctx context.Context, claims *jwt_model.JWTClaims, p authorization.Permission) error {
	aRoles, err := claimRolesToAuthRoles(claims)
	var permissions []authorization.Permission
	if err == nil {
		permissions, _ = c.authorizer.GetUserPermissions(aRoles)
	}

	c.logger.Warn(ctx, "permission denied",
		zap.Uint64("id", claims.ID),
		zap.Uint64("tenant_id", claims.TenantID),
		zap.String("required_permission", string(p.Name)),
		zap.Strings("user_permissions", authorization.UniquePermissionNames(permissions)),
		zap.Strings("user_roles", authorization.UniqueRoleNames(claims.Roles)))
	return status.Errorf(codes.PermissionDenied, "required permission: %v", p)
}

func claimRolesToAuthRoles(claims *jwt_model.JWTClaims) ([]authorization.Role, error) {
	roles := make([]authorization.Role, 0, len(claims.Roles))
	for _, r := range claims.Roles {
		role, err := authorization.ToRole(r.Name, r.Context)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}
