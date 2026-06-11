package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/suselz/mihomo-exporter/internal/config"
	"github.com/suselz/mihomo-exporter/internal/engine"
	"github.com/suselz/mihomo-exporter/internal/health"
	"github.com/suselz/mihomo-exporter/internal/httpserver"
	"github.com/suselz/mihomo-exporter/internal/metrics"
	"github.com/suselz/mihomo-exporter/internal/mihomo"
)

var version = "dev"

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	logger.Printf("starting mihomo-exporter version=%s", version)
	client := mihomo.NewClient(cfg.BaseURL, cfg.ConnectionsPath)
	eng := engine.New(cfg.MaxClientSeries, cfg.ClientAllowed)
	m := metrics.New()
	h := health.New()
	srv := httpserver.New(cfg.ListenAddr, cfg.MetricsPath, m.Registry(), h)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go runPoller(ctx, logger, cfg, client, eng, m, h)

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatalf("http server failed: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Printf("shutdown error: %v", err)
	}
}

func runPoller(
	ctx context.Context,
	logger *log.Logger,
	cfg config.Config,
	client *mihomo.Client,
	eng *engine.Engine,
	m *metrics.Metrics,
	h *health.State,
) {
	ticker := time.NewTicker(cfg.ScrapeInterval)
	defer ticker.Stop()

	scrape := func() {
		started := time.Now()
		snap, err := client.Fetch(ctx)
		dur := time.Since(started)
		if err != nil {
			if errors.Is(err, mihomo.ErrDecode) {
				m.IncParseErrors()
			}
			m.SetScrape(false, dur)
			h.MarkFailure(err)
			logger.Printf("scrape failed: %v", err)
			return
		}

		res := eng.Process(snap)
		m.Apply(res)
		m.SetScrape(true, dur)
		h.MarkSuccess()
	}

	scrape()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			scrape()
		}
	}
}
