package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/CHORUS-TRE/chorus-backend/internal/logger"

	"go.uber.org/zap"
)

func AddLogger(log *logger.ContextLogger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Do not log Kubernetes health check.
		if r.Header.Get("X-Liveness-Probe") == "Healthz" {
			h.ServeHTTP(w, r)
			return
		}

		var scheme string
		if r.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
		proto := r.Proto
		method := r.Method
		remoteAddr := r.RemoteAddr
		userAgent := r.UserAgent()
		uri := strings.Join([]string{scheme, "://", r.Host, r.RequestURI}, "")

		log.Debug(ctx, "request started",
			zap.String("http-scheme", scheme),
			zap.String("http-proto", proto),
			zap.String("http-method", method),
			zap.String("remote-addr", remoteAddr),
			zap.String("user-agent", userAgent),
			zap.String("uri", uri),
		)

		t1 := time.Now()

		h.ServeHTTP(w, r)

		log.Debug(ctx, "request completed",
			zap.String("http-scheme", scheme),
			zap.String("http-proto", proto),
			zap.String("http-method", method),
			zap.String("remote-addr", remoteAddr),
			zap.String("user-agent", userAgent),
			zap.String("uri", uri),
			zap.Float64(logger.LoggerKeyElapsedMs, float64(time.Since(t1).Nanoseconds())/1000000.0),
		)
	})
}
