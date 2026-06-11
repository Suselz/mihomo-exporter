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
	"github.com/suselz/mihomo-exporter/internal/logx"
	"github.com/suselz/mihomo-exporter/internal/metrics"
	"github.com/suselz/mihomo-exporter/internal/mihomo"
)

var version = "dev"

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	logger, err := logx.New(cfg.LogLevel)
	if err != nil {
		log.Fatalf("logger init error: %v", err)
	}
	logger.Infof("starting mihomo-exporter version=%s", version)
	logger.Infof("config mihomo_url=%s connections_path=%s listen_addr=%s metrics_path=%s scrape_interval=%s log_level=%s", cfg.BaseURL, cfg.ConnectionsPath, cfg.ListenAddr, cfg.MetricsPath, cfg.ScrapeInterval, logger.LevelString())
	client := mihomo.NewClient(cfg.BaseURL, cfg.ConnectionsPath)
	eng := engine.New(cfg.MaxClientSeries, cfg.ClientAllowed)
	m := metrics.New()
	h := health.New()
	srv := httpserver.New(cfg.ListenAddr, cfg.MetricsPath, m.Registry(), h, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go runPoller(ctx, logger, cfg, client, eng, m, h)

	go func() {
		logger.Infof("http server starting listen_addr=%s", cfg.ListenAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("http server failed: %v", err)
			os.Exit(1)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	logger.Infof("shutdown signal received, stopping")
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("shutdown error: %v", err)
		return
	}
	logger.Infof("shutdown complete")
}

func runPoller(
	ctx context.Context,
	logger *logx.Logger,
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
			logger.Errorf("scrape failed: %v", err)
			return
		}

		res := eng.Process(snap)
		m.Apply(res)

		if mem, err := client.FetchMemory(ctx); err == nil {
			m.SetMemory(mem)
		} else {
			if errors.Is(err, mihomo.ErrDecode) {
				m.IncParseErrors()
			}
			logger.Warnf("memory fetch failed: %v", err)
		}

		if tr, err := client.FetchTraffic(ctx); err == nil {
			m.SetTraffic(tr)
		} else {
			if errors.Is(err, mihomo.ErrDecode) {
				m.IncParseErrors()
			}
			logger.Warnf("traffic fetch failed: %v", err)
		}

		if ver, err := client.FetchVersion(ctx); err == nil {
			m.SetVersion(ver)
		} else {
			if errors.Is(err, mihomo.ErrDecode) {
				m.IncParseErrors()
			}
			logger.Warnf("version fetch failed: %v", err)
		}

		if proxies, err := client.FetchProxies(ctx); err == nil {
			m.SetProxies(proxies)
		} else {
			if errors.Is(err, mihomo.ErrDecode) {
				m.IncParseErrors()
			}
			logger.Warnf("proxies fetch failed: %v", err)
		}

		if rules, err := client.FetchRules(ctx); err == nil {
			m.SetRules(rules)
		} else {
			if errors.Is(err, mihomo.ErrDecode) {
				m.IncParseErrors()
			}
			logger.Warnf("rules fetch failed: %v", err)
		}

		m.SetScrape(true, dur)
		h.MarkSuccess()
		if res.NegativeDeltaCount > 0 || res.ConnectionResetCount > 0 {
			logger.Warnf("scrape completed with anomalies duration=%s connections=%d negative_deltas=%d resets=%d", dur, res.ConnectionsTotal, res.NegativeDeltaCount, res.ConnectionResetCount)
			return
		}
		logger.Debugf("scrape completed duration=%s connections=%d total_upload_delta=%d total_download_delta=%d", dur, res.ConnectionsTotal, res.TotalUploadDelta, res.TotalDownloadDelta)
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
