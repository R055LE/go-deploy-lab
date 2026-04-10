package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/R055LE/go-deploy-lab/internal/metrics"
)

type metricsRecorder struct {
	http.ResponseWriter
	status int
}

func (r *metricsRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// Metrics records Prometheus request count and latency.
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &metricsRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		metrics.HTTPRequestsTotal.WithLabelValues(
			r.Method, r.URL.Path, strconv.Itoa(rec.status),
		).Inc()
		metrics.HTTPRequestDuration.WithLabelValues(
			r.Method, r.URL.Path,
		).Observe(time.Since(start).Seconds())
	})
}
