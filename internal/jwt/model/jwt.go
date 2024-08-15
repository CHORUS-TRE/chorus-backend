package jwt

import (
	"context"
	"errors"
	"fmt"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"

	"github.com/golang-jwt/jwt"
	"go.uber.org/zap"
)

type contextKey string

const JWTClaimsContextKey contextKey = "JWTClaims"
const JWTTokenContextKey contextKey = "JWTToken"

type ErrInvalidClaims struct {
	msg string
}

func (e *ErrInvalidClaims) Error() string {
	return fmt.Sprintf("invalid claims: %s", e.msg)
}

type ClaimsFactory func(context.Context) jwt.Claims

func NewJWTClaimsFactory(logger *logger.ContextLogger) ClaimsFactory {
	return func(ctx context.Context) jwt.Claims {
		return &JWTClaims{
			ctx:    ctx,
			logger: logger,
		}
	}
}

// JWTClaims is the JWT claims section. The jwt.StandardClaims and a list of authorised clients
// are validated. The supported claims are:
//
//	{
//			"client": string (client)
//			"aud": string (Audience)
//			"exp": int64 (ExpiresAt)
//			"jti": string (Id)
//			"iat": int64 (IssuedAt)
//			"iss": string (Issuer)
//			"nbf": int64 (NotBefore)
//			"sub": string (Subject)
//	}
type JWTClaims struct {
	ID        uint64   `json:"id"`
	TenantID  uint64   `json:"tenantID"`
	FirstName string   `json:"firstName"`
	LastName  string   `json:"lastName"`
	Roles     []string `json:"roles"`
	Username  string   `json:"username"`
	ctx       context.Context

	jwt.StandardClaims
	logger *logger.ContextLogger
}

func ExtractTenantID(ctx context.Context) (uint64, error) {
	claims, ok := ctx.Value(JWTClaimsContextKey).(*JWTClaims)
	if !ok {
		return 0, errors.New("malformed jwt-token")
	}
	if claims.TenantID == 0 {
		return 0, errors.New("invalid tenant in jwt-token")
	}
	return claims.TenantID, nil
}

func ExtractUserID(ctx context.Context) (uint64, error) {
	claims, ok := ctx.Value(JWTClaimsContextKey).(*JWTClaims)
	if !ok {
		return 0, errors.New("malformed jwt-token")
	}
	return claims.ID, nil
}

func (c *JWTClaims) Valid() error {
	if err := c.StandardClaims.Valid(); err != nil {
		c.logger.Error(c.ctx, "claims are not valid",
			zap.String("unit", "authentication"),
			zap.String("status", "failure"),
			zap.Error(err),

			zap.Uint64("id", c.ID),
			zap.Uint64("tenant_id", c.TenantID),
			zap.Strings("roles", c.Roles),
		)
		return &ErrInvalidClaims{msg: err.Error()}
	}
	c.logger.Info(c.ctx, "claims are valid",
		zap.String("unit", "authentication"),
		zap.Uint64("id", c.ID),
		zap.String("status", "success"),

		zap.Uint64("id", c.ID),
		zap.Uint64("tenant_id", c.TenantID),
		zap.Strings("roles", c.Roles),
	)
	return nil
}
