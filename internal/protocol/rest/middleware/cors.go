package middleware

import (
	"net/http"
	"strings"

	"github.com/CHORUS-TRE/chorus-backend/internal/config"
)

// AddCORS returns a new http.Handler that allows Cross Origin Resoruce Sharing.
func AddCORS(h http.Handler, cfg config.Config) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// XXX: What origins do we allow? Currently all can pass.
		w.Header().Set("Access-Control-Allow-Origin", cfg.Daemon.HTTP.Headers.AccessControlAllowOrigin)
		w.Header().Set("Access-Control-Max-Age", cfg.Daemon.HTTP.Headers.AccessControlMaxAge)
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
			headers := []string{"Content-Type", "Accept", "Authorization", "Access-Control-Allow-Credentials"}
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(headers, ","))
			methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(methods, ","))
			return
		}
		h.ServeHTTP(w, r)
	})
}
