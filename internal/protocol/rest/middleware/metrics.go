package middleware

import (
	"net/http"
	"strings"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func AddMetrics(h http.Handler) http.Handler {
	prom := promhttp.Handler()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// Serving prometheus metrics under the prefix "/metrics"
		if strings.HasPrefix(r.RequestURI, "/metrics") {
			prom.ServeHTTP(w, r)
			return
		}

		// If not "/metrics", passing to the next middleware
		h.ServeHTTP(w, r.WithContext(ctx))
	})

}
