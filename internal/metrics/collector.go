package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/suselz/mihomo-exporter/internal/engine"
	"github.com/suselz/mihomo-exporter/internal/mihomo"
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

	memoryInuse    prometheus.Gauge
	memoryOSLimit  prometheus.Gauge
	trafficUp      prometheus.Gauge
	trafficDown    prometheus.Gauge
	trafficUpTot   prometheus.Gauge
	trafficDownTot prometheus.Gauge
	versionInfo    *prometheus.GaugeVec
	proxiesTotal   prometheus.Gauge
	proxiesAlive   prometheus.Gauge
	proxiesByType  *prometheus.GaugeVec
	rulesTotal     prometheus.Gauge
	rulesDisabled  prometheus.Gauge
	rulesHits      prometheus.Gauge
	rulesMisses    prometheus.Gauge
	rulesByType    *prometheus.GaugeVec

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

	m.memoryInuse = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mihomo_memory_inuse_bytes", Help: "Mihomo memory in-use bytes"})
	m.memoryOSLimit = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mihomo_memory_oslimit_bytes", Help: "Mihomo memory OS limit bytes"})
	m.trafficUp = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mihomo_traffic_up_bytes_per_second", Help: "Current upload speed from /traffic"})
	m.trafficDown = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mihomo_traffic_down_bytes_per_second", Help: "Current download speed from /traffic"})
	m.trafficUpTot = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mihomo_traffic_upload_total_bytes", Help: "Upload total from /traffic"})
	m.trafficDownTot = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mihomo_traffic_download_total_bytes", Help: "Download total from /traffic"})
	m.versionInfo = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mihomo_version_info", Help: "Mihomo version info (value=1)"}, []string{"version", "meta"})
	m.proxiesTotal = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mihomo_proxies_total", Help: "Total proxy entries from /proxies"})
	m.proxiesAlive = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mihomo_proxies_alive_total", Help: "Alive proxy entries from /proxies"})
	m.proxiesByType = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mihomo_proxies_by_type", Help: "Proxy count by proxy type"}, []string{"type"})
	m.rulesTotal = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mihomo_rules_total", Help: "Total rules from /rules"})
	m.rulesDisabled = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mihomo_rules_disabled_total", Help: "Disabled rules from /rules"})
	m.rulesHits = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mihomo_rules_hits", Help: "Sum of rule hitCount from /rules"})
	m.rulesMisses = prometheus.NewGauge(prometheus.GaugeOpts{Name: "mihomo_rules_misses", Help: "Sum of rule missCount from /rules"})
	m.rulesByType = prometheus.NewGaugeVec(prometheus.GaugeOpts{Name: "mihomo_rules_by_type", Help: "Rule count by type"}, []string{"type"})

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
		m.memoryInuse, m.memoryOSLimit, m.trafficUp, m.trafficDown, m.trafficUpTot, m.trafficDownTot,
		m.versionInfo, m.proxiesTotal, m.proxiesAlive, m.proxiesByType,
		m.rulesTotal, m.rulesDisabled, m.rulesHits, m.rulesMisses, m.rulesByType,
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

func (m *Metrics) SetMemory(mem mihomo.MemorySnapshot) {
	m.memoryInuse.Set(float64(mem.Inuse))
	m.memoryOSLimit.Set(float64(mem.OSLimit))
}

func (m *Metrics) SetTraffic(tr mihomo.TrafficSnapshot) {
	m.trafficUp.Set(float64(tr.Up))
	m.trafficDown.Set(float64(tr.Down))
	m.trafficUpTot.Set(float64(tr.UpTotal))
	m.trafficDownTot.Set(float64(tr.DownTotal))
}

func (m *Metrics) SetVersion(v mihomo.VersionSnapshot) {
	m.versionInfo.Reset()
	m.versionInfo.WithLabelValues(v.Version, strconv.FormatBool(v.Meta)).Set(1)
}

func (m *Metrics) SetProxies(p mihomo.ProxiesSnapshot) {
	var total int
	var alive int
	byType := make(map[string]int)
	for _, proxy := range p.Proxies {
		total++
		if proxy.Alive {
			alive++
		}
		t := proxy.Type
		if t == "" {
			t = "unknown"
		}
		byType[t]++
	}
	m.proxiesTotal.Set(float64(total))
	m.proxiesAlive.Set(float64(alive))
	m.proxiesByType.Reset()
	for t, c := range byType {
		m.proxiesByType.WithLabelValues(t).Set(float64(c))
	}
}

func (m *Metrics) SetRules(r mihomo.RulesSnapshot) {
	byType := make(map[string]int)
	var disabled int
	var hits uint64
	var misses uint64
	for _, rule := range r.Rules {
		t := rule.Type
		if t == "" {
			t = "unknown"
		}
		byType[t]++
		if rule.Extra.Disabled {
			disabled++
		}
		hits += rule.Extra.HitCount
		misses += rule.Extra.MissCount
	}
	m.rulesTotal.Set(float64(len(r.Rules)))
	m.rulesDisabled.Set(float64(disabled))
	m.rulesHits.Set(float64(hits))
	m.rulesMisses.Set(float64(misses))
	m.rulesByType.Reset()
	for t, c := range byType {
		m.rulesByType.WithLabelValues(t).Set(float64(c))
	}
}
