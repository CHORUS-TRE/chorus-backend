package middleware

import (
	"context"
	"io/fs"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	embed "github.com/CHORUS-TRE/chorus-backend/api"
	"github.com/CHORUS-TRE/chorus-backend/internal/api/v1/middleware"
	"github.com/CHORUS-TRE/chorus-backend/internal/authorization"
	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	jwt_model "github.com/CHORUS-TRE/chorus-backend/internal/jwt/model"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	jwt_go "github.com/golang-jwt/jwt"

	"go.uber.org/zap"
)

type ProxyWorkbenchHandler func(ctx context.Context, tenantID, workbenchID uint64, w http.ResponseWriter, r *http.Request) error

func AddProxyWorkbench(h http.Handler, pw ProxyWorkbenchHandler, cfg config.Config, authorizer authorization.Authorizer, keyFunc jwt_go.Keyfunc, claimsFactory jwt_model.ClaimsFactory) http.Handler {
	reg := regexp.MustCompile(`^/api/rest/v1/workbenchs/([0-9]+)/stream`)

	auth := middleware.NewAuthorization(logger.TechLog, cfg, authorizer, nil)

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

		err = auth.IsAuthorized(ctx, authorization.PermissionStreamWorkbench, authorization.WithWorkbench(workbenchID))
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

func AddDevAuth(h http.Handler) http.Handler {
	devAuthFS, _ := fs.Sub(embed.DevAuthEmbed, "dev-auth")
	fs := http.FS(devAuthFS)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		u, err := url.Parse(r.RequestURI)
		if err != nil {
			h.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		if u.Path == "/dev-auth" {
			u.Path = "/dev-auth/"
			handler := http.RedirectHandler(u.String(), http.StatusFound)
			handler.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		if strings.HasPrefix(r.RequestURI, "/dev-auth") {
			handler := http.StripPrefix("/dev-auth", http.FileServer(fs))
			handler.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
