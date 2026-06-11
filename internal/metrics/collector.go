package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/suselz/mihomo-exporter/internal/engine"
)

type Metrics struct {
	registry *prometheus.Registry

	clientUpload   *prometheus.CounterVec
	clientDownload *prometheus.CounterVec
	outUpload      *prometheus.CounterVec
	outDownload    *prometheus.CounterVec
	totalUpload    prometheus.Counter
	totalDownload  prometheus.Counter

	clientConn   *prometheus.GaugeVec
	outboundConn *prometheus.GaugeVec
	connTotal    prometheus.Gauge

	scrapeSuccess    prometheus.Gauge
	lastScrapeUnix   prometheus.Gauge
	scrapeDuration   prometheus.Gauge
	parseErrors      prometheus.Counter
	resets           prometheus.Counter
	negativeDeltas   prometheus.Counter
	trackedClients   prometheus.Gauge
	trackedOutbounds prometheus.Gauge
}

func New() *Metrics {
	m := &Metrics{registry: prometheus.NewRegistry()}

	m.clientUpload = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "mihomo_client_upload_bytes_total", Help: "Upload bytes by client IP"}, []string{"client_ip"})
	m.clientDownload = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "mihomo_client_download_bytes_total", Help: "Download bytes by client IP"}, []string{"client_ip"})
	m.outUpload = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "mihomo_outbound_upload_bytes_total", Help: "Upload bytes by outbound"}, []string{"outbound"})
	m.outDownload = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "mihomo_outbound_download_bytes_total", Help: "Download bytes by outbound"}, []string{"outbound"})
	m.totalUpload = prometheus.NewCounter(prometheus.CounterOpts{Name: "mihomo_total_upload_bytes_total", Help: "Total upload bytes"})
	m.totalDownload = prometheus.NewCounter(prometheus.CounterOpts{Name: "mihomo_total_download_bytes_total", Help: "Total download bytes"})

	m.clientConn = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mihomo_client_connections", Help: "Active connections by client IP"}, []string{"client_ip"})
	m.outboundConn = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mihomo_outbound_connections", Help: "Active connections by outbound"}, []string{"outbound"})
	m.connTotal = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mihomo_connections_total", Help: "Total active connections"})

	m.scrapeSuccess = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mihomo_exporter_scrape_success", Help: "Last scrape success (1/0)"})
	m.lastScrapeUnix = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mihomo_exporter_last_scrape_timestamp_seconds", Help: "Unix timestamp of last successful scrape"})
	m.scrapeDuration = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mihomo_exporter_scrape_duration_seconds", Help: "Last scrape duration"})
	m.parseErrors = prometheus.NewCounter(prometheus.CounterOpts{Name: "mihomo_exporter_parse_errors_total", Help: "Number of parse errors"})
	m.resets = prometheus.NewCounter(prometheus.CounterOpts{Name: "mihomo_exporter_connection_resets_total", Help: "Detected reset events"})
	m.negativeDeltas = prometheus.NewCounter(prometheus.CounterOpts{Name: "mihomo_exporter_negative_delta_total", Help: "Negative per-connection delta events"})
	m.trackedClients = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mihomo_exporter_tracked_clients", Help: "Tracked client series count"})
	m.trackedOutbounds = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mihomo_exporter_tracked_outbounds", Help: "Tracked outbound series count"})

	m.registry.MustRegister(
		m.clientUpload, m.clientDownload, m.outUpload, m.outDownload,
		m.totalUpload, m.totalDownload,
		m.clientConn, m.outboundConn, m.connTotal,
		m.scrapeSuccess, m.lastScrapeUnix, m.scrapeDuration, m.parseErrors, m.resets, m.negativeDeltas, m.trackedClients, m.trackedOutbounds,
	)
	return m
}

func (m *Metrics) Registry() *prometheus.Registry { return m.registry }

func (m *Metrics) IncParseErrors() { m.parseErrors.Inc() }

func (m *Metrics) SetScrape(success bool, d time.Duration) {
	if success {
		m.scrapeSuccess.Set(1)
		m.lastScrapeUnix.Set(float64(time.Now().Unix()))
	} else {
		m.scrapeSuccess.Set(0)
	}
	m.scrapeDuration.Set(d.Seconds())
}

func (m *Metrics) Apply(res engine.Result) {
	for k, v := range res.ClientUploadDelta {
		m.clientUpload.WithLabelValues(k).Add(float64(v))
	}
	for k, v := range res.ClientDownloadDelta {
		m.clientDownload.WithLabelValues(k).Add(float64(v))
	}
	for k, v := range res.OutboundUploadDelta {
		m.outUpload.WithLabelValues(k).Add(float64(v))
	}
	for k, v := range res.OutboundDownloadDelta {
		m.outDownload.WithLabelValues(k).Add(float64(v))
	}
	m.totalUpload.Add(float64(res.TotalUploadDelta))
	m.totalDownload.Add(float64(res.TotalDownloadDelta))

	m.clientConn.Reset()
	for k, v := range res.ClientConnections {
		m.clientConn.WithLabelValues(k).Set(float64(v))
	}
	m.outboundConn.Reset()
	for k, v := range res.OutboundConnections {
		m.outboundConn.WithLabelValues(k).Set(float64(v))
	}
	m.connTotal.Set(float64(res.ConnectionsTotal))

	m.negativeDeltas.Add(float64(res.NegativeDeltaCount))
	m.resets.Add(float64(res.ConnectionResetCount))
	m.trackedClients.Set(float64(res.TrackedClientSeries))
	m.trackedOutbounds.Set(float64(res.TrackedOutboundSeries))
}
