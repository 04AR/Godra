package api

import (
	"log/slog"
	"net/http"
	"time"

	"godra/internal/metrics"

	"github.com/go-chi/chi/v5/middleware"
)

func RequestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

		metrics.TotalRequests.Add(1)

		next.ServeHTTP(ww, r)

		slog.Info("Request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.Status(),
			"duration", time.Since(start),
			"remote", r.RemoteAddr,
		)
	})
}
