package engine

import (
	"testing"

	"github.com/suselz/mihomo-exporter/internal/mihomo"
)

func TestProcessComputesDeltasAndAggregates(t *testing.T) {
	e := New(100, nil)

	first := mihomo.Snapshot{
		UploadTotal:   100,
		DownloadTotal: 200,
		Connections: []mihomo.Connection{
			{ID: "c1", Upload: 10, Download: 20, Chains: []string{"out-a"}, Metadata: mihomo.Metadata{SourceIP: "10.0.0.1"}},
			{ID: "c2", Upload: 30, Download: 40, Chains: []string{"out-b"}, Metadata: mihomo.Metadata{SourceIP: "10.0.0.2"}},
		},
	}
	res := e.Process(first)
	if res.TotalUploadDelta != 0 || res.TotalDownloadDelta != 0 {
		t.Fatalf("first scrape must not produce deltas, got up=%d down=%d", res.TotalUploadDelta, res.TotalDownloadDelta)
	}

	second := mihomo.Snapshot{
		UploadTotal:   160,
		DownloadTotal: 290,
		Connections: []mihomo.Connection{
			{ID: "c1", Upload: 25, Download: 55, Chains: []string{"out-a"}, Metadata: mihomo.Metadata{SourceIP: "10.0.0.1"}},
			{ID: "c2", Upload: 35, Download: 50, Chains: []string{"out-b"}, Metadata: mihomo.Metadata{SourceIP: "10.0.0.2"}},
		},
	}
	res = e.Process(second)

	if got, want := res.TotalUploadDelta, uint64(60); got != want {
		t.Fatalf("TotalUploadDelta=%d want=%d", got, want)
	}
	if got, want := res.TotalDownloadDelta, uint64(90); got != want {
		t.Fatalf("TotalDownloadDelta=%d want=%d", got, want)
	}
	if got := res.ClientUploadDelta["10.0.0.1"]; got != 15 {
		t.Fatalf("client 10.0.0.1 upload delta=%d want=15", got)
	}
	if got := res.OutboundDownloadDelta["out-b"]; got != 10 {
		t.Fatalf("out-b download delta=%d want=10", got)
	}
	if got := res.ConnectionsTotal; got != 2 {
		t.Fatalf("ConnectionsTotal=%d want=2", got)
	}
}

func TestProcessDetectsNegativeDelta(t *testing.T) {
	e := New(100, nil)

	_ = e.Process(mihomo.Snapshot{
		UploadTotal:   100,
		DownloadTotal: 100,
		Connections: []mihomo.Connection{
			{ID: "c1", Upload: 50, Download: 50, Chains: []string{"out-a"}, Metadata: mihomo.Metadata{SourceIP: "10.0.0.1"}},
		},
	})

	res := e.Process(mihomo.Snapshot{
		UploadTotal:   120,
		DownloadTotal: 130,
		Connections: []mihomo.Connection{
			{ID: "c1", Upload: 49, Download: 60, Chains: []string{"out-a"}, Metadata: mihomo.Metadata{SourceIP: "10.0.0.1"}},
		},
	})

	if res.NegativeDeltaCount != 1 {
		t.Fatalf("NegativeDeltaCount=%d want=1", res.NegativeDeltaCount)
	}
	if res.ConnectionResetCount != 1 {
		t.Fatalf("ConnectionResetCount=%d want=1", res.ConnectionResetCount)
	}
	if got := res.ClientUploadDelta["10.0.0.1"]; got != 0 {
		t.Fatalf("unexpected upload delta for negative sample: %d", got)
	}
}
