package helper

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	jwt_go "github.com/golang-jwt/jwt"
)

var (
	// ErrTokenContextMissing denotes a token was not passed into the parsing middleware's context.
	ErrTokenContextMissing = errors.New("token up for parsing was not passed through the context")

	// ErrTokenInvalid denotes a token was not able to be validated.
	ErrTokenInvalid = errors.New("jwt-token was invalid")

	// ErrTokenExpired denotes a token's expire header (exp) has since passed.
	ErrTokenExpired = errors.New("jwt-token is expired")

	// ErrTokenMalformed denotes a token was not formatted as a JWT-token.
	ErrTokenMalformed = errors.New("jwt-token is malformed")

	// ErrTokenNotActive denotes a token's not before header (nbf) is in the future.
	ErrTokenNotActive = errors.New("jwt-token is not valid yet")

	// ErrUnexpectedSigningMethod denotes a token was signed with an unexpected signing method.
	ErrUnexpectedSigningMethod = errors.New("unexpected signing method")

	// ErrSignatureValidationFailed denotes a token whose signature could not be verified.
	ErrSignatureValidationFailed = errors.New("invalid signature: hmac is different, please check jwt-secrets")
)

// ParseToken reads and verifies a provided JWT-token using a key function and a claims factory.
// It returns a 'jwt_go.Claims' object upon successful validation.
func ParseToken(ctx context.Context, tokenString string, keyFunc jwt_go.Keyfunc, claimsFactory jwt_model.ClaimsFactory) (jwt_go.Claims, error) {
	token, err := jwt_go.ParseWithClaims(tokenString, claimsFactory(ctx), func(token *jwt_go.Token) (interface{}, error) {
		// Check that signature method is correct.
		if token.Method != jwt_go.SigningMethodHS256 {
			return nil, ErrUnexpectedSigningMethod
		}
		return keyFunc(token)
	})
	if err != nil {
		if e, ok := err.(*jwt_go.ValidationError); ok {
			switch {
			case e.Errors&jwt_go.ValidationErrorMalformed != 0:
				return nil, ErrTokenMalformed
			case e.Errors&jwt_go.ValidationErrorExpired != 0:
				return nil, ErrTokenExpired
			case e.Errors&jwt_go.ValidationErrorNotValidYet != 0:
				return nil, ErrTokenNotActive
			case e.Errors&jwt_go.ValidationErrorSignatureInvalid != 0:
				return nil, ErrSignatureValidationFailed
			case e.Inner != nil:
				return nil, e.Inner // Miscellaneous error.
			}
			// We have a ValidationError but have no specific JWT-error for it.
			// Fall through to return original error.
		}
		return nil, err
	}

	if !token.Valid {
		return nil, ErrTokenInvalid
	}

	c, ok := token.Claims.(*jwt_model.JWTClaims)
	if !ok {
		return nil, ErrTokenInvalid
	}

	err = DecompressRoles(c)
	if err != nil {
		return nil, fmt.Errorf("failed to decode roles from token: %w", err)
	}

	return token.Claims, nil
}

func DecompressRoles(claims *jwt_model.JWTClaims) error {
	if len(claims.Roles) > 0 || len(claims.R) == 0 {
		return nil
	}

	gzBytes, err := base64.StdEncoding.DecodeString(claims.R)
	if err != nil {
		return fmt.Errorf("failed to base64-decode roles: %w", err)
	}

	gr, err := gzip.NewReader(bytes.NewReader(gzBytes))
	if err != nil {
		return fmt.Errorf("failed to create gzip reader: %w", err)
	}
	out, err := io.ReadAll(gr)
	if err != nil {
		return fmt.Errorf("failed to read gzip data: %w", err)
	}
	if err := gr.Close(); err != nil {
		return fmt.Errorf("failed to close gzip reader: %w", err)
	}

	if err := json.Unmarshal(out, &claims.Roles); err != nil {
		return fmt.Errorf("failed to unmarshal roles json: %w", err)
	}

	return nil
}

// userFromClaims retrieves the first and last name from
// JWT-claim as a human-readable string.
func UserFromClaims(claims jwt_go.Claims) string {
	c, ok := claims.(*jwt_model.JWTClaims)
	if !ok {
		return "bad claims format"
	}
	return fmt.Sprintf("UserID: %v, TenantID: %v", c.ID, c.TenantID)
}

func TenantIDFromClaims(claims jwt_go.Claims) uint64 {
	c, ok := claims.(*jwt_model.JWTClaims)
	if !ok {
		return 0
	}
	return c.TenantID
}

func RolesFromClaims(claims jwt_go.Claims) []jwt_model.Role {
	c, ok := claims.(*jwt_model.JWTClaims)
	if !ok {
		return nil
	}
	return c.Roles
}
