package middleware

import (
	"net/http"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func AddMetrics(h http.Handler, cfg config.Config) http.Handler {
	prom := promhttp.Handler()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// Serving prometheus metrics under the prefix "/metrics"
		if cfg.Daemon.Metrics.Enabled && strings.HasPrefix(r.RequestURI, "/metrics") {
			if cfg.Daemon.Metrics.Authentication.Enabled {
				username, password, ok := r.BasicAuth()
				if !ok || username != cfg.Daemon.Metrics.Authentication.Username || password != cfg.Daemon.Metrics.Authentication.Password.PlainText() {
					http.Error(w, "Unauthorized", http.StatusUnauthorized)
					return
				}
			}

			prom.ServeHTTP(w, r)
			return
		}

		// If not "/metrics", passing to the next middleware
		h.ServeHTTP(w, r.WithContext(ctx))
	})

}
