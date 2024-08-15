package middleware

import (
	"context"

	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type authorization struct {
	logger          *logger.ContextLogger
	authorizedRoles []string
}

func NewAuthorization(logger *logger.ContextLogger, authorizedRoles []string) authorization {
	return authorization{
		logger,
		authorizedRoles,
	}
}

func (c authorization) IsAuthenticatedAndAuthorized(ctx context.Context) error {
	claims, ok := ctx.Value(jwt_model.JWTClaimsContextKey).(*jwt_model.JWTClaims)
	if !ok {
		c.logger.Warn(ctx, "malformed JWT token")
		return status.Error(codes.Unauthenticated, "malformed jwt-token")
	}
	if !c.isAuthorized(claims.Roles) {
		return c.permissionDenied(ctx, claims)
	}
	return nil
}

func (c authorization) IsAuthenticatedAndAuthorizedWithRoles(ctx context.Context, roles []string) error {
	claims, ok := ctx.Value(jwt_model.JWTClaimsContextKey).(*jwt_model.JWTClaims)
	if !ok {
		c.logger.Warn(ctx, "malformed JWT token")
		return status.Error(codes.Unauthenticated, "malformed jwt-token")
	}
	if !c.isAuthorized(claims.Roles) && !hasAnyOfRoles(claims.Roles, roles) {
		return c.permissionDenied(ctx, claims)
	}
	return nil
}

func (c authorization) permissionDenied(ctx context.Context, claims *jwt_model.JWTClaims) error {
	c.logger.Warn(ctx, "permission denied",
		zap.Uint64("id", claims.ID),
		zap.Uint64("tenant_id", claims.TenantID),
		zap.Strings("roles", claims.Roles))
	return status.Errorf(codes.PermissionDenied, "authorized roles: %v", c.authorizedRoles)
}

func (c authorization) isAuthorized(roles []string) bool {
	for _, r := range roles {
		for _, authorizedRole := range c.authorizedRoles {
			if r == authorizedRole {
				return true
			}
		}
	}
	return false
}

func hasRole(role string, roles []string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

func hasAnyOfRoles(roles []string, wantedRoles []string) bool {
	for _, r := range roles {
		if hasRole(r, wantedRoles) {
			return true
		}
	}
	return false
}
