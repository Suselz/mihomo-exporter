package mihomo

type Snapshot struct {
	DownloadTotal uint64       `json:"downloadTotal"`
	UploadTotal   uint64       `json:"uploadTotal"`
	Memory        uint64       `json:"memory"`
	Connections   []Connection `json:"connections"`
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
