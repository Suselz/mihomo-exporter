package httpserver

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/suselz/mihomo-exporter/internal/health"
	"github.com/suselz/mihomo-exporter/internal/logx"
)

type Server struct {
	http.Server
}

func New(listenAddr, metricsPath string, reg *prometheus.Registry, hs *health.State, logger *logx.Logger) *Server {
	mux := http.NewServeMux()
	mux.Handle(metricsPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		logger.Debugf("http request path=%s remote=%s", r.URL.Path, r.RemoteAddr)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "timestamp": time.Now().UTC().Format(time.RFC3339)})
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, r *http.Request) {
		logger.Debugf("http request path=%s remote=%s", r.URL.Path, r.RemoteAddr)
		ready, lastErr, lastSuccess := hs.Snapshot()
		w.Header().Set("Content-Type", "application/json")
		if !ready {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ready":        ready,
			"last_error":   lastErr,
			"last_success": lastSuccess.UTC().Format(time.RFC3339),
		})
	})

	return &Server{Server: http.Server{Addr: listenAddr, Handler: mux}}
}
