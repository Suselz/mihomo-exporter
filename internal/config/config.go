package config

import (
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	BaseURL         string
	ConnectionsPath string
	ListenAddr      string
	MetricsPath     string
	ScrapeInterval  time.Duration
	MaxClientSeries int
	ClientAllowed   []*net.IPNet
}

func Load() (Config, error) {
	cfg := Config{
		BaseURL:         getenv("MIHOMO_URL", "http://10.66.0.2:9090"),
		ConnectionsPath: getenv("MIHOMO_CONNECTIONS_PATH", "/connections"),
		ListenAddr:      getenv("EXPORTER_LISTEN_ADDR", ":9109"),
		MetricsPath:     getenv("EXPORTER_METRICS_PATH", "/metrics"),
		MaxClientSeries: getenvInt("EXPORTER_MAX_CLIENT_SERIES", 2000),
	}

	interval, err := time.ParseDuration(getenv("EXPORTER_SCRAPE_INTERVAL", "2s"))
	if err != nil {
		return Config{}, fmt.Errorf("EXPORTER_SCRAPE_INTERVAL: %w", err)
	}
	if interval <= 0 {
		return Config{}, fmt.Errorf("EXPORTER_SCRAPE_INTERVAL must be > 0")
	}
	cfg.ScrapeInterval = interval

	allowed, err := parseCIDRs(getenv("EXPORTER_CLIENT_ALLOW_CIDRS", ""))
	if err != nil {
		return Config{}, fmt.Errorf("EXPORTER_CLIENT_ALLOW_CIDRS: %w", err)
	}
	cfg.ClientAllowed = allowed

	if cfg.MaxClientSeries <= 0 {
		return Config{}, fmt.Errorf("EXPORTER_MAX_CLIENT_SERIES must be > 0")
	}

	return cfg, nil
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func parseCIDRs(raw string) ([]*net.IPNet, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	out := make([]*net.IPNet, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		_, network, err := net.ParseCIDR(p)
		if err != nil {
			return nil, fmt.Errorf("bad cidr %q: %w", p, err)
		}
		out = append(out, network)
	}
	return out, nil
}
