package httpserver

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/suselz/mihomo-exporter/internal/health"
)

type Server struct {
	http.Server
}

func New(listenAddr, metricsPath string, reg *prometheus.Registry, hs *health.State) *Server {
	mux := http.NewServeMux()
	mux.Handle(metricsPath, promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "timestamp": time.Now().UTC().Format(time.RFC3339)})
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
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
