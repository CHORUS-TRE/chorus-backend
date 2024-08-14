package middleware

import (
	"context"
	"net/http"
	"regexp"
	"strconv"

	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	jwt_go "github.com/golang-jwt/jwt"
	"go.uber.org/zap"
)

type ProxyWorkbenchHandler func(ctx context.Context, tenantID, workbenchID uint64, w http.ResponseWriter, r *http.Request) error

func AddProxyWorkbench(h http.Handler, pw ProxyWorkbenchHandler, keyFunc jwt_go.Keyfunc, claimsFactory jwt_model.ClaimsFactory) http.Handler {
	reg := regexp.MustCompile(`^/api/rest/v1/workbenchs/([0-9]+)/stream`)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		m := reg.FindStringSubmatch(r.RequestURI)
		if m == nil {
			// no match, continue to next mid
			h.ServeHTTP(w, r)
			return
		}

		remainingPath := reg.ReplaceAllString(r.RequestURI, "")
		if remainingPath == "" {
			handler := http.RedirectHandler(r.RequestURI+"/", http.StatusFound)
			handler.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		workbenchID, err := strconv.ParseUint(m[1], 10, 32)
		if err != nil {
			logger.TechLog.Error(context.Background(), "unable to parse workbenchID", zap.Error(err))
			h.ServeHTTP(w, r)
			return
		}

		// TODO get jwt via cookie

		// jwt := r.Header.Get("Authorization")
		// tokenString := strings.TrimPrefix(jwt, "Bearer ")

		// claims, err := jwt_helper.ParseToken(ctx, tokenString, keyFunc, claimsFactory)
		// if err != nil {
		// 	logger.TechLog.Error(context.Background(), "invalid authentication token", zap.Error(err))
		// 	h.ServeHTTP(w, r)
		// 	return
		// }
		// tenantID := jwt_helper.TenantIDFromClaims(claims)
		// ctx = context.WithValue(ctx, jwt_model.JWTClaimsContextKey, claims)
		// ctx = context.WithValue(ctx, logger.TenantIDContextKey{}, tenantID)

		err = pw(ctx, 1, workbenchID, w, r.WithContext(ctx))
		if err != nil {
			logger.TechLog.Error(context.Background(), "unable to proxy", zap.Error(err))
			h.ServeHTTP(w, r)
			return
		}
	})
}
