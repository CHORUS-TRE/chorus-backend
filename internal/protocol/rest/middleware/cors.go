package middleware

import (
	"net/http"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"go.uber.org/zap"
)

// AddCORS returns a new http.Handler that allows Cross Origin Resoruce Sharing.
func AddCORS(h http.Handler, cfg config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		logger.TechLog.Debug(r.Context(), "Add CORS",
			zap.String("origin", origin), zap.String("method", r.Method), zap.String("path", r.URL.Path),
		)

		headers := GetCORSHeaders(origin, r.Method, cfg)
		for k, v := range headers {
			w.Header().Set(k, v)
		}

		h.ServeHTTP(w, r)
	})
}

func GetCORSHeaders(origin string, method string, cfg config.Config) map[string]string {
	headers := make(map[string]string)

	if isOriginInAllowedList(origin, cfg.Daemon.HTTP.Headers.AccessControlAllowOrigins) {
		headers["Access-Control-Allow-Origin"] = origin
	} else if cfg.Daemon.HTTP.Headers.AccessControlAllowOriginWildcard {
		headers["Access-Control-Allow-Origin"] = "*"
	} else if len(cfg.Daemon.HTTP.Headers.AccessControlAllowOrigins) > 0 {
		headers["Access-Control-Allow-Origin"] = cfg.Daemon.HTTP.Headers.AccessControlAllowOrigins[0]
	}

	if len(cfg.Daemon.HTTP.Headers.AccessControlAllowOrigins) > 1 || cfg.Daemon.HTTP.Headers.AccessControlAllowOriginWildcard {
		headers["Vary"] = "Origin" // Avoid cache poisoning
	}

	headers["Access-Control-Max-Age"] = cfg.Daemon.HTTP.Headers.AccessControlMaxAge
	headers["Access-Control-Allow-Credentials"] = "true"
	if method == "OPTIONS" {
		hdrs := []string{"Content-Type", "Accept", "Authorization", "Access-Control-Allow-Credentials"}
		headers["Access-Control-Allow-Headers"] = strings.Join(hdrs, ",")
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
		headers["Access-Control-Allow-Methods"] = strings.Join(methods, ",")
	}

	return headers
}

func isOriginInAllowedList(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	for _, allowedOrigin := range allowedOrigins {
		if len(allowedOrigin) <= len(origin) && origin[:len(allowedOrigin)] == allowedOrigin {
			return true
		}
	}

	return false
}
