package mihomo

type Snapshot struct {
	DownloadTotal uint64       `json:"downloadTotal"`
	UploadTotal   uint64       `json:"uploadTotal"`
	Memory        uint64       `json:"memory"`
	Connections   []Connection `json:"connections"`
}

type MemorySnapshot struct {
	Inuse   uint64 `json:"inuse"`
	OSLimit uint64 `json:"oslimit"`
}

type TrafficSnapshot struct {
	Up        uint64 `json:"up"`
	Down      uint64 `json:"down"`
	UpTotal   uint64 `json:"upTotal"`
	DownTotal uint64 `json:"downTotal"`
}

type VersionSnapshot struct {
	Meta    bool   `json:"meta"`
	Version string `json:"version"`
}

type ProxiesSnapshot struct {
	Proxies map[string]Proxy `json:"proxies"`
}

type Proxy struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	Alive bool   `json:"alive"`
}

type RulesSnapshot struct {
	Rules []Rule `json:"rules"`
}

type Rule struct {
	Type  string    `json:"type"`
	Extra RuleExtra `json:"extra"`
}

type RuleExtra struct {
	Disabled  bool   `json:"disabled"`
	HitCount  uint64 `json:"hitCount"`
	MissCount uint64 `json:"missCount"`
}

type Connection struct {
	ID       string   `json:"id"`
	Upload   uint64   `json:"upload"`
	Download uint64   `json:"download"`
	Chains   []string `json:"chains"`
	Metadata Metadata `json:"metadata"`
}

type Metadata struct {
	SourceIP string `json:"sourceIP"`
}
