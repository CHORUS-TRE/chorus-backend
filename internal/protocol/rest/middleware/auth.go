package middleware

import (
	"context"
	"net/http"
	"strings"

	jwt_helper "github.com/CHORUS-TRE/chorus-backend/internal/jwt/helper"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	jwt_go "github.com/golang-jwt/jwt"
)

func AddJWTFromCookie(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwt := r.Header.Get("Authorization")
		if jwt == "" {
			token, err := r.Cookie("jwttoken")
			if err == nil {
				r.Header.Set("Authorization", "Bearer "+token.Value)
			}
		}

		h.ServeHTTP(w, r)
	})
}

func GetContextWithAuth(ctx context.Context, r *http.Request, keyFunc jwt_go.Keyfunc, claimsFactory jwt_model.ClaimsFactory) context.Context {
	jwt := r.Header.Get("Authorization")
	tokenString := strings.TrimPrefix(jwt, "Bearer ")

	claims, err := jwt_helper.ParseToken(ctx, tokenString, keyFunc, claimsFactory)
	if err != nil {
		return ctx
	}
	tenantID := jwt_helper.TenantIDFromClaims(claims)
	ctx = context.WithValue(ctx, jwt_model.JWTClaimsContextKey, claims)
	ctx = context.WithValue(ctx, logger.TenantIDContextKey{}, tenantID)

	return ctx
}
