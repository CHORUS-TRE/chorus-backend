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

		if isOriginInAllowedList(origin, cfg.Daemon.HTTP.Headers.AccessControlAllowOrigins) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if cfg.Daemon.HTTP.Headers.AccessControlAllowOriginWildcard {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if len(cfg.Daemon.HTTP.Headers.AccessControlAllowOrigins) > 0 {
			w.Header().Set("Access-Control-Allow-Origin", cfg.Daemon.HTTP.Headers.AccessControlAllowOrigins[0])
		}

		if len(cfg.Daemon.HTTP.Headers.AccessControlAllowOrigins) > 1 || cfg.Daemon.HTTP.Headers.AccessControlAllowOriginWildcard {
			w.Header().Set("Vary", "Origin") // Avoid cache poisoning
		}

		w.Header().Set("Access-Control-Max-Age", cfg.Daemon.HTTP.Headers.AccessControlMaxAge)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
			headers := []string{"Content-Type", "Accept", "Authorization", "Access-Control-Allow-Credentials"}
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
			methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
			return
		}
		h.ServeHTTP(w, r)
	})
}

func isOriginInAllowedList(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	for _, allowedOrigin := range allowedOrigins {
		if origin[:len(allowedOrigin)] == allowedOrigin {
			return true
		}
	}

	return false
}
