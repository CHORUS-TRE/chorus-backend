package middleware

import (
	"context"

	"github.com/CHORUS-TRE/chorus-backend/internal/authorization"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Authorization struct {
	logger     *logger.ContextLogger
	authorizer authorization.Authorizer
	// authorizedRoles []string
}

func NewAuthorization(logger *logger.ContextLogger, authorizer authorization.Authorizer) Authorization {
	return Authorization{
		logger:     logger,
		authorizer: authorizer,
		// authorizedRoles: []string{},
	}
}

func (c Authorization) IsAuthorized(ctx context.Context, permissionName authorization.PermissionName, opts ...authorization.NewContextOption) error {
	permission := authorization.NewPermission(permissionName, opts...)

	claims, ok := ctx.Value(jwt_model.JWTClaimsContextKey).(*jwt_model.JWTClaims)
	if !ok {
		c.logger.Warn(ctx, "malformed JWT token")
		return status.Error(codes.Unauthenticated, "malformed jwt-token")
	}

	aRoles := make([]authorization.Role, 0, len(claims.Roles))
	for _, r := range claims.Roles {
		ar, err := authorization.ToRoleName(r)
		if err != nil {
			c.logger.Warn(ctx, "invalid role", zap.String("role", r))
			continue
		}
		aRoles = append(aRoles, authorization.Role{
			Name: ar,
		})
	}

	isAuthorized, err := c.authorizer.IsUserAllowed(aRoles, permission)
	if err != nil {
		c.logger.Error(ctx, "authorization error", zap.Error(err))
		return status.Error(codes.Internal, "authorization error")
	}

	if !isAuthorized {
		return c.permissionDenied(ctx, claims, aRoles)
	}

	return nil
}

func (c Authorization) permissionDenied(ctx context.Context, claims *jwt_model.JWTClaims, authorizedRoles []authorization.Role) error {
	c.logger.Warn(ctx, "permission denied",
		zap.Uint64("id", claims.ID),
		zap.Uint64("tenant_id", claims.TenantID),
		zap.Strings("roles", claims.Roles))
	return status.Errorf(codes.PermissionDenied, "authorized roles: %v", authorizedRoles)
}
