package middleware

import (
	"context"
	"net/http"
	"regexp"
	"strconv"

	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/middleware"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"github.com/CHORUS-TRE/chorus-backend/pkg/user/model"
	jwt_go "github.com/golang-jwt/jwt"
	"go.uber.org/zap"
)

type ProxyWorkbenchHandler func(ctx context.Context, tenantID, workbenchID uint64, w http.ResponseWriter, r *http.Request) error

func AddProxyWorkbench(h http.Handler, pw ProxyWorkbenchHandler, keyFunc jwt_go.Keyfunc, claimsFactory jwt_model.ClaimsFactory) http.Handler {
	reg := regexp.MustCompile(`^/api/rest/v1/workbenchs/([0-9]+)/stream`)

	auth := middleware.NewAuthorization(logger.TechLog, []string{model.RoleAuthenticated.String()})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		m := reg.FindStringSubmatch(r.RequestURI)
		if m == nil {
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

		ctx = GetContextWithAuth(ctx, r, keyFunc, claimsFactory)

		err = auth.IsAuthenticatedAndAuthorized(ctx)
		if err != nil {
			logger.TechLog.Error(context.Background(), "invalid authentication token", zap.Error(err))
			h.ServeHTTP(w, r)
			return
		}

		err = pw(ctx, 1, workbenchID, w, r.WithContext(ctx))
		if err != nil {
			logger.TechLog.Error(context.Background(), "unable to proxy", zap.Error(err))
			h.ServeHTTP(w, r)
			return
		}
	})
}
