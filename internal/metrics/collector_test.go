package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/suselz/mihomo-exporter/internal/engine"
	"github.com/suselz/mihomo-exporter/internal/mihomo"
)

func TestApplyAndSetters(t *testing.T) {
	m := New()

	m.Apply(engine.Result{
		ClientUploadDelta:     map[string]uint64{"10.0.0.1": 11},
		ClientDownloadDelta:   map[string]uint64{"10.0.0.1": 22},
		OutboundUploadDelta:   map[string]uint64{"out-a": 33},
		OutboundDownloadDelta: map[string]uint64{"out-a": 44},
		ClientConnections:     map[string]int{"10.0.0.1": 2},
		OutboundConnections:   map[string]int{"out-a": 3},
		ConnectionsTotal:      5,
		TotalUploadDelta:      100,
		TotalDownloadDelta:    200,
		NegativeDeltaCount:    1,
		ConnectionResetCount:  2,
		TrackedClientSeries:   7,
		TrackedOutboundSeries: 8,
	})

	m.SetMemory(mihomo.MemorySnapshot{Inuse: 1024, OSLimit: 2048})
	m.SetTraffic(mihomo.TrafficSnapshot{Up: 1, Down: 2, UpTotal: 3, DownTotal: 4})
	m.SetVersion(mihomo.VersionSnapshot{Meta: true, Version: "v1.2.3"})
	m.SetProxies(mihomo.ProxiesSnapshot{Proxies: map[string]mihomo.Proxy{
		"a": {Type: "Selector", Alive: true},
		"b": {Type: "Direct", Alive: false},
	}})
	m.SetRules(mihomo.RulesSnapshot{Rules: []mihomo.Rule{
		{Type: "InName", Extra: mihomo.RuleExtra{Disabled: false, HitCount: 10, MissCount: 20}},
		{Type: "InName", Extra: mihomo.RuleExtra{Disabled: true, HitCount: 1, MissCount: 2}},
	}})

	if got := testutil.ToFloat64(m.totalUpload); got != 100 {
		t.Fatalf("total upload=%v want=100", got)
	}
	if got := testutil.ToFloat64(m.clientUpload.WithLabelValues("10.0.0.1")); got != 11 {
		t.Fatalf("client upload=%v want=11", got)
	}
	if got := testutil.ToFloat64(m.memoryInuse); got != 1024 {
		t.Fatalf("memory inuse=%v want=1024", got)
	}
	if got := testutil.ToFloat64(m.trafficDownTot); got != 4 {
		t.Fatalf("traffic down total=%v want=4", got)
	}
	if got := testutil.ToFloat64(m.versionInfo.WithLabelValues("v1.2.3", "true")); got != 1 {
		t.Fatalf("version info=%v want=1", got)
	}
	if got := testutil.ToFloat64(m.proxiesTotal); got != 2 {
		t.Fatalf("proxies total=%v want=2", got)
	}
	if got := testutil.ToFloat64(m.rulesDisabled); got != 1 {
		t.Fatalf("rules disabled=%v want=1", got)
	}
	if got := testutil.ToFloat64(m.rulesHits); got != 11 {
		t.Fatalf("rules hits=%v want=11", got)
	}
}
