package middleware

import (
	"net/http"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"
	"go.uber.org/zap"
)

func AddRoot(h http.Handler, version string, started <-chan struct{}) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		select {
		case <-started:
			if r.RequestURI == "" || r.RequestURI == "/" {
				w.Header().Set("Content-Type", "application/json; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte(version))
				if err != nil {
					logger.TechLog.Error(r.Context(), "unable to write response", zap.Error(err))
				}

				return
			}
		default:
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, err := w.Write([]byte("{}"))
			if err != nil {
				logger.TechLog.Error(r.Context(), "unable to write response", zap.Error(err))
			}

			return
		}

		// Move to the next handler.
		h.ServeHTTP(w, r)
	})
}
