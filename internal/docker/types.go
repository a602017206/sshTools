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

type ContainerDetail struct {
	ID          string
	Name        string
	Image       string
	State       string
	Status      string
	Ports       string
	NetworkMode string
	Created     int64
	Mounts      []string
	Env         []string
	Cmd         []string
	Entrypoint  []string
}

type ImageSummary struct {
	ID      string
	Tags    []string
	Size    int64
	Created int64
	// Via 标识数据来源；经远程 docker CLI 回退列出时为 "cli"，API 模式为空。
	Via string `json:"via,omitempty"`
}

type ImageDetail struct {
	ImageSummary
	Architecture string
	Os           string
	Author       string
	Comment      string
}
