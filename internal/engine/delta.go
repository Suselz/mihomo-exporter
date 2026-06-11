package engine

import (
	"net"

	"github.com/suselz/mihomo-exporter/internal/mihomo"
)

type prevConn struct {
	upload   uint64
	download uint64
}

type Engine struct {
	initialized       bool
	prevConnections   map[string]prevConn
	prevUploadTotal   uint64
	prevDownloadTotal uint64
	maxClientSeries   int
	allowCIDRs        []*net.IPNet
	knownClients      map[string]struct{}
	knownOutbounds    map[string]struct{}
}

func New(maxClientSeries int, allowCIDRs []*net.IPNet) *Engine {
	return &Engine{
		prevConnections: make(map[string]prevConn),
		maxClientSeries: maxClientSeries,
		allowCIDRs:      allowCIDRs,
		knownClients:    make(map[string]struct{}),
		knownOutbounds:  make(map[string]struct{}),
	}
}

func (e *Engine) Process(s mihomo.Snapshot) Result {
	res := Result{
		ClientUploadDelta:     map[string]uint64{},
		ClientDownloadDelta:   map[string]uint64{},
		OutboundUploadDelta:   map[string]uint64{},
		OutboundDownloadDelta: map[string]uint64{},
		ClientConnections:     map[string]int{},
		OutboundConnections:   map[string]int{},
		ConnectionsTotal:      len(s.Connections),
	}

	if !e.initialized {
		e.prevUploadTotal = s.UploadTotal
		e.prevDownloadTotal = s.DownloadTotal
		for _, c := range s.Connections {
			e.prevConnections[c.ID] = prevConn{upload: c.Upload, download: c.Download}
		}
		e.initialized = true
		return res
	}

	if s.UploadTotal >= e.prevUploadTotal {
		res.TotalUploadDelta = s.UploadTotal - e.prevUploadTotal
	} else {
		res.ConnectionResetCount++
	}
	if s.DownloadTotal >= e.prevDownloadTotal {
		res.TotalDownloadDelta = s.DownloadTotal - e.prevDownloadTotal
	} else {
		res.ConnectionResetCount++
	}
	e.prevUploadTotal = s.UploadTotal
	e.prevDownloadTotal = s.DownloadTotal

	next := make(map[string]prevConn, len(s.Connections))
	for _, c := range s.Connections {
		client := e.normalizeClient(c.Metadata.SourceIP)
		outbound := normalizeOutbound(c.Chains)
		res.OutboundConnections[outbound]++
		if client != "" {
			res.ClientConnections[client]++
		}

		curr := prevConn{upload: c.Upload, download: c.Download}
		prev, ok := e.prevConnections[c.ID]
		next[c.ID] = curr
		if !ok {
			continue
		}

		uDelta, uNeg := diff(curr.upload, prev.upload)
		dDelta, dNeg := diff(curr.download, prev.download)
		if uNeg || dNeg {
			res.NegativeDeltaCount++
			res.ConnectionResetCount++
			continue
		}

		if client != "" {
			res.ClientUploadDelta[client] += uDelta
			res.ClientDownloadDelta[client] += dDelta
		}
		res.OutboundUploadDelta[outbound] += uDelta
		res.OutboundDownloadDelta[outbound] += dDelta
	}
	e.prevConnections = next

	for k := range res.ClientConnections {
		e.knownClients[k] = struct{}{}
	}
	for k := range res.OutboundConnections {
		e.knownOutbounds[k] = struct{}{}
	}
	res.TrackedClientSeries = len(e.knownClients)
	res.TrackedOutboundSeries = len(e.knownOutbounds)

	return res
}

func normalizeOutbound(chains []string) string {
	if len(chains) == 0 || chains[0] == "" {
		return "unknown"
	}
	return chains[0]
}

func diff(curr, prev uint64) (uint64, bool) {
	if curr < prev {
		return 0, true
	}
	return curr - prev, false
}

func (e *Engine) normalizeClient(ipRaw string) string {
	ip := net.ParseIP(ipRaw)
	if ip == nil {
		return ""
	}
	if len(e.allowCIDRs) > 0 {
		allowed := false
		for _, n := range e.allowCIDRs {
			if n.Contains(ip) {
				allowed = true
				break
			}
		}
		if !allowed {
			return ""
		}
	}
	label := ip.String()
	if _, ok := e.knownClients[label]; ok {
		return label
	}
	if len(e.knownClients) >= e.maxClientSeries {
		return "__other__"
	}
	return label
}
