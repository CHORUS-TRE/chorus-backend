package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type codeResponseWriter struct {
	http.ResponseWriter
	StatusCode int
}

// AddInstrumenting add prometheus metrics
func AddInstrumenting(h http.Handler) http.Handler {

	latency := promauto.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "http_server_handling_seconds",
		Help:       "Total duration of requests in seconds.",
		Objectives: map[float64]float64{0.5: 0.05, 0.99: 0.001, 0.999: 0.0001},
	}, []string{"http_endpoint", "http_status"})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		t1 := time.Now()

		cw := NewCodeResponseWriter(w)
		h.ServeHTTP(cw, r)

		// Save only these http status code to not overload metrics
		switch cw.StatusCode {
		case http.StatusOK, http.StatusInternalServerError, http.StatusServiceUnavailable:
			latency.WithLabelValues(r.RequestURI, strconv.Itoa(cw.StatusCode)).Observe(time.Since(t1).Seconds())
		}

	})
}

func NewCodeResponseWriter(w http.ResponseWriter) *codeResponseWriter {
	// WriteHeader(int) is not called if our response implicitly returns 200 OK, so
	// we default to that status code.
	return &codeResponseWriter{w, http.StatusOK}
}

func (w *codeResponseWriter) WriteHeader(code int) {
	w.StatusCode = code
	w.ResponseWriter.WriteHeader(code)
}
