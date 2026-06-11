package engine

type Result struct {
	ClientUploadDelta     map[string]uint64
	ClientDownloadDelta   map[string]uint64
	OutboundUploadDelta   map[string]uint64
	OutboundDownloadDelta map[string]uint64
	ClientConnections     map[string]int
	OutboundConnections   map[string]int
	ConnectionsTotal      int
	TotalUploadDelta      uint64
	TotalDownloadDelta    uint64
	NegativeDeltaCount    int
	ConnectionResetCount  int
	TrackedClientSeries   int
	TrackedOutboundSeries int
}
