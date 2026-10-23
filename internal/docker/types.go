package docker

const (
	TransportLocal = "local"
	TransportSSH   = "ssh"
	TransportTCP   = "tcp"
)

type Config struct {
	Transport  string
	Host       string
	Port       int
	SocketPath string
	DockerHost string
	TLS        bool
}

type ContainerSummary struct {
	ID      string
	Name    string
	Image   string
	Status  string
	Ports   string
	Created int64
	// Via 标识数据来源；经远程 docker CLI 回退列出时为 "cli"，API 模式为空。
	Via string `json:"via,omitempty"`
}
