package middleware

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	chorus_errors "github.com/CHORUS-TRE/chorus-backend/internal/errors"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	authz_model "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/model"
	authorization_service "github.com/CHORUS-TRE/chorus-backend/pkg/authorization/service"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
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

func (c Authorization) getRolesAndClaims(ctx context.Context) ([]authz_model.Role, *jwt_model.JWTClaims, error) {

	claims, ok := ctx.Value(jwt_model.JWTClaimsContextKey).(*jwt_model.JWTClaims)
	if !ok {
		c.logger.Warn(ctx, "malformed JWT token", zap.Any("content", ctx.Value(jwt_model.JWTClaimsContextKey)))
		return nil, nil, chorus_errors.ErrUnauthenticated.WithMessage("malformed jwt-token")
	}

	aRoles, err := claimRolesToAuthRoles(claims)
	if err != nil {
		c.logger.Error(ctx, "failed to convert claim roles to auth roles", zap.Error(err))
		return nil, nil, chorus_errors.ErrInternal.Wrap(err, "failed to convert claim roles to auth roles")
	}

	return aRoles, claims, nil
}

func (c Authorization) IsAuthorized(ctx context.Context, permission authz_model.Permission) error {
	aRoles, claims, err := c.getRolesAndClaims(ctx)
	if err != nil {
		c.logger.Error(ctx, "failed to get roles and claims", zap.Error(err))
		return err
	}

	isAuthorized, err := c.authorizer.IsUserAllowed(aRoles, permission)
	if err != nil {
		c.logger.Error(ctx, "failed to evaluate permission", zap.String("permission", permission.Name.String()), zap.Error(err))
		return chorus_errors.ErrInternal.Wrap(err, fmt.Sprintf("failed to evaluate permission %s: %v", permission.Name, err))
	}

	if !isAuthorized {
		return c.permissionDenied(ctx, claims, aRoles, permission)
	}

	return nil
}

func (c Authorization) GetContextListForPermission(ctx context.Context, permissionName authz_model.PermissionName) ([]authz_model.Context, error) {
	aRoles, _, err := c.getRolesAndClaims(ctx)
	if err != nil {
		c.logger.Error(ctx, "failed to get roles and claims", zap.Error(err))
		return nil, err
	}

	contextList, err := c.authorizer.GetContextListForPermission(aRoles, permissionName)
	if err != nil {
		c.logger.Error(ctx, "failed to get context list for permission", zap.String("permission", string(permissionName)), zap.Error(err))
		return nil, chorus_errors.ErrInternal.Wrap(err, fmt.Sprintf("failed to get context list for permission %s: %v", permissionName, err))
	}

	return contextList, nil
}

func (c Authorization) IsRoleInScope(roleName authz_model.RoleName, scopes ...authz_model.RoleScope) bool {
	return c.authorizer.IsRoleInScope(roleName, scopes...)
}

func (c Authorization) CanAssignRole(ctx context.Context, roleName authz_model.RoleName, assignmentContext authz_model.Context) error {
	aRoles, _, err := c.getRolesAndClaims(ctx)
	if err != nil {
		c.logger.Error(ctx, "failed to get roles and claims", zap.Error(err))
		return chorus_errors.ErrInvalidRequest.Wrap(err, "failed to get roles and claims")
	}

	allowed, err := c.authorizer.CanAssignRole(aRoles, roleName, assignmentContext)
	if err != nil {
		c.logger.Error(ctx, "role assignment authorization error", zap.Error(err))
		return chorus_errors.ErrPermissionDenied.Wrap(err, "role assignment authorization error")
	}
	if !allowed {
		return chorus_errors.ErrPermissionDenied.WithMessage(fmt.Sprintf("caller cannot assign role %s", roleName))
	}
	return nil
}

func (c Authorization) TriggerRefreshToken(ctx context.Context) error {
	res, t, err := c.refresher.RefreshToken(ctx)
	if err != nil {
		return chorus_errors.ErrUnauthenticated.Wrap(err, "unable to refresh token")
	}

	expiresDate := time.Now().Add(t)
	expires := expiresDate.Format(time.RFC1123)

	header := c.getSetCookieHeader(res, expires)
	if err := grpc.SetHeader(ctx, header); err != nil {
		return chorus_errors.ErrInternal.Wrap(err, "unable to set grpc set-cookie header")
	}

	return nil
}

func (c Authorization) getSetCookieHeader(token string, expires string) metadata.MD {
	return metadata.Pairs("Set-Cookie", "jwttoken="+token+"; Path=/; Domain="+c.cfg.Daemon.HTTP.Headers.CookieDomain+"; SameSite=None; Secure; HttpOnly; Expires="+expires)
}

func (c Authorization) permissionDenied(ctx context.Context, claims *jwt_model.JWTClaims, aRoles []authz_model.Role, p authz_model.Permission) error {
	permissions, _ := c.authorizer.GetUserPermissions(aRoles)
	explanation := c.authorizer.ExplainIsUserAllowed(aRoles, p)

	c.logger.Warn(ctx, "permission denied",
		zap.Uint64("id", claims.ID),
		zap.Uint64("tenant_id", claims.TenantID),
		zap.String("required_permission", string(p.Name)),
		zap.String("explanation", explanation),
		zap.Strings("user_permissions", authz_model.UniquePermissionNames(permissions)),
		zap.Strings("user_roles", uniqueRoleNames(claims.Roles)))
	return chorus_errors.ErrPermissionDenied.WithMessage(fmt.Sprintf("required permission: %v", p))
}

func claimRolesToAuthRoles(claims *jwt_model.JWTClaims) ([]authz_model.Role, error) {
	roles := make([]authz_model.Role, 0, len(claims.Roles))
	for _, r := range claims.Roles {
		role, err := authz_model.ToRole(r.Name, r.Context)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, nil
}

// userFromCtx resolves the caller's own user id from the JWT claims, for a
// self-scoped permission check (e.g. "get my own user").
func userFromCtx(ctx context.Context) authz_model.UserID {
	if claims, ok := ctx.Value(jwt_model.JWTClaimsContextKey).(*jwt_model.JWTClaims); ok {
		return authz_model.UserID(claims.ID)
	}
	return 0
}

// uniqueRoleNames returns a sorted, deduplicated list of role names from the
// caller's JWT roles.
func uniqueRoleNames(roles []jwt_model.Role) []string {
	seen := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		seen[r.Name] = struct{}{}
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}
