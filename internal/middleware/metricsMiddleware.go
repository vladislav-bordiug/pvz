package middleware

import (
	"net/http"
	"strconv"
	"time"

	"pvz/internal/metrics"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (m *Middleware) MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := newResponseWriter(w)
		next.ServeHTTP(rw, r)
		duration := time.Since(start).Seconds()
		codeStr := strconv.Itoa(rw.statusCode)
		handlerName := r.URL.Path
		metrics.HTTPRequestTotal.WithLabelValues(handlerName, r.Method, codeStr).Inc()
		metrics.HTTPResponseDuration.WithLabelValues(handlerName, r.Method, codeStr).Observe(duration)
	})
}
