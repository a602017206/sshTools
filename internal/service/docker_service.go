package service

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"AHaSSHTools/internal/config"
	"AHaSSHTools/internal/docker"
	"AHaSSHTools/internal/ssh"
)

const (
	dockerModeAPI = "API"
	dockerModeCLI = "CLI"

	dockerViaCLI = "cli"

	defaultDockerRemoteSocket = "/var/run/docker.sock"
	dockerOpTimeout           = 15 * time.Second
)

// dockerSSHSession is the SSH capability DockerService needs. It is backed by a
// dedicated ssh.Client that is never shared with terminal sessions.
type dockerSSHSession interface {
	Dial(network, addr string) (net.Conn, error)
	CombinedOutput(cmd string) ([]byte, error)
	Close() error
}

// sshFactory creates and connects a dedicated SSH session.
type sshFactory func(cfg *ssh.Config) (dockerSSHSession, error)

func defaultDockerSSHFactory(cfg *ssh.Config) (dockerSSHSession, error) {
	client, err := ssh.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	if err := client.Connect(); err != nil {
		return nil, err
	}
	return client, nil
}

// dockerAssetStore resolves saved SSH assets and their credentials.
// ConnectionService satisfies it.
type dockerAssetStore interface {
	GetConnection(id string) (config.ConnectionConfig, error)
	GetPassword(connectionID string) (string, error)
}

// DockerConnectOptions is the parsed form of a docker ConnectionConfig.
type DockerConnectOptions struct {
	Transport       string // local | ssh | tcp
	Host            string
	Port            int
	TLS             bool
	SocketPath      string
	DockerHost      string
	SSHConnectionID string
	RemoteSocket    string
	CLIFallback     bool
}

func parseDockerOptions(conn config.ConnectionConfig) DockerConnectOptions {
	md := conn.Metadata
	get := func(k string) string { return strings.TrimSpace(md[k]) }
	opts := DockerConnectOptions{
		Transport:       strings.ToLower(get("transport")),
		Host:            strings.TrimSpace(conn.Host),
		Port:            conn.Port,
		TLS:             strings.EqualFold(get("tls"), "true"),
		SocketPath:      get("socket_path"),
		DockerHost:      get("docker_host"),
		SSHConnectionID: get("ssh_connection_id"),
		RemoteSocket:    get("remote_socket"),
		CLIFallback:     get("cli_fallback") != "false",
	}
	if opts.Transport == "" {
		switch {
		case opts.SSHConnectionID != "":
			opts.Transport = docker.TransportSSH
		case opts.Host != "":
			opts.Transport = docker.TransportTCP
		default:
			opts.Transport = docker.TransportLocal
		}
	}
	if opts.RemoteSocket == "" {
		opts.RemoteSocket = defaultDockerRemoteSocket
	}
	return opts
}

type dockerSession struct {
	mode      string
	apiClient *docker.Client
	tunnel    *docker.UnixTunnel
	ssh       dockerSSHSession
}

func (s *dockerSession) close() error {
	var first error
	keep := func(err error) {
		if err != nil && first == nil {
			first = err
		}
	}
	if s.apiClient != nil {
		keep(s.apiClient.Close())
	}
	if s.tunnel != nil {
		keep(s.tunnel.Close())
	}
	if s.ssh != nil {
		keep(s.ssh.Close())
	}
	return first
}

// DockerService manages Docker connections (local socket, TCP, SSH tunnel).
type DockerService struct {
	assets   dockerAssetStore
	newSSH   sshFactory
	mu       sync.Mutex
	sessions map[string]*dockerSession
}

// NewDockerService creates a DockerService. assets may be nil when only local/tcp
// transports are used.
func NewDockerService(assets dockerAssetStore) *DockerService {
	return newDockerServiceWithFactory(assets, defaultDockerSSHFactory)
}

func newDockerServiceWithFactory(assets dockerAssetStore, factory sshFactory) *DockerService {
	return &DockerService{
		assets:   assets,
		newSSH:   factory,
		sessions: make(map[string]*dockerSession),
	}
}

// DetectLocalDocker pings the default local Docker endpoint. It does not need a
// saved asset. Unavailability is reported via ok=false and detail, not err.
func (s *DockerService) DetectLocalDocker() (bool, string, error) {
	client, err := docker.NewClient(docker.Config{Transport: docker.TransportLocal})
	if err != nil {
		return false, err.Error(), nil
	}
	defer client.Close()
	ctx, cancel := context.WithTimeout(context.Background(), dockerOpTimeout)
	defer cancel()
	if err := client.Ping(ctx); err != nil {
		return false, err.Error(), nil
	}
	return true, "本机 Docker 可用", nil
}

// TestDockerConnection opens and immediately closes a connection.
func (s *DockerService) TestDockerConnection(conn config.ConnectionConfig) error {
	sess, err := s.open(conn)
	if err != nil {
		return err
	}
	return sess.close()
}

// ConnectDocker opens a connection and registers it under sessionID, replacing
// (and closing) any previous session with the same ID.
func (s *DockerService) ConnectDocker(sessionID string, conn config.ConnectionConfig) error {
	if strings.TrimSpace(sessionID) == "" {
		return fmt.Errorf("docker: sessionID 不能为空")
	}
	sess, err := s.open(conn)
	if err != nil {
		return err
	}
	s.mu.Lock()
	old := s.sessions[sessionID]
	s.sessions[sessionID] = sess
	s.mu.Unlock()
	if old != nil {
		_ = old.close()
	}
	return nil
}

// CloseAll closes every registered Docker session (used on app shutdown).
func (s *DockerService) CloseAll() error {
	s.mu.Lock()
	sessions := s.sessions
	s.sessions = make(map[string]*dockerSession)
	s.mu.Unlock()
	var first error
	for _, sess := range sessions {
		if err := sess.close(); err != nil && first == nil {
			first = err
		}
	}
	return first
}

// CloseDocker closes the session. Unknown sessions are a no-op.
func (s *DockerService) CloseDocker(sessionID string) error {
	s.mu.Lock()
	sess := s.sessions[sessionID]
	delete(s.sessions, sessionID)
	s.mu.Unlock()
	if sess == nil {
		return nil
	}
	return sess.close()
}

// ListDockerContainers lists containers through the API, or the remote docker CLI
// when the session fell back to CLI mode.
func (s *DockerService) ListDockerContainers(sessionID string, all bool) ([]docker.ContainerSummary, error) {
	s.mu.Lock()
	sess := s.sessions[sessionID]
	s.mu.Unlock()
	if sess == nil {
		return nil, fmt.Errorf("docker: 会话 %q 不存在或已关闭", sessionID)
	}
	if sess.mode == dockerModeCLI {
		return listContainersCLI(sess.ssh, all)
	}
	ctx, cancel := context.WithTimeout(context.Background(), dockerOpTimeout)
	defer cancel()
	return sess.apiClient.ListContainers(ctx, all)
}

func (s *DockerService) sessionMode(sessionID string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sess := s.sessions[sessionID]; sess != nil {
		return sess.mode
	}
	return ""
}

func (s *DockerService) open(conn config.ConnectionConfig) (*dockerSession, error) {
	opts := parseDockerOptions(conn)
	switch opts.Transport {
	case docker.TransportLocal:
		return openAPIClient(docker.Config{
			Transport:  docker.TransportLocal,
			SocketPath: opts.SocketPath,
			DockerHost: opts.DockerHost,
		})
	case docker.TransportTCP:
		port := opts.Port
		if port <= 0 {
			port = 2375
			if opts.TLS {
				port = 2376
			}
		}
		return openAPIClient(docker.Config{
			Transport: docker.TransportTCP,
			Host:      opts.Host,
			Port:      port,
			TLS:       opts.TLS,
		})
	case docker.TransportSSH:
		return s.openSSH(opts)
	default:
		return nil, fmt.Errorf("docker: 未知传输方式 %q", opts.Transport)
	}
}

func openAPIClient(cfg docker.Config) (*dockerSession, error) {
	client, err := docker.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	if err := pingDocker(client); err != nil {
		_ = client.Close()
		return nil, err
	}
	return &dockerSession{mode: dockerModeAPI, apiClient: client}, nil
}

func pingDocker(client *docker.Client) error {
	ctx, cancel := context.WithTimeout(context.Background(), dockerOpTimeout)
	defer cancel()
	return client.Ping(ctx)
}

func (s *DockerService) sshConfigFor(opts DockerConnectOptions) (*ssh.Config, error) {
	if opts.SSHConnectionID == "" {
		return nil, fmt.Errorf("docker: ssh 传输需要 ssh_connection_id")
	}
	if s.assets == nil {
		return nil, fmt.Errorf("docker: 未初始化 SSH 资产存储")
	}
	asset, err := s.assets.GetConnection(opts.SSHConnectionID)
	if err != nil {
		return nil, fmt.Errorf("docker: 读取 SSH 资产失败: %w", err)
	}
	port := asset.Port
	if port <= 0 {
		port = 22
	}
	cfg := &ssh.Config{Host: asset.Host, Port: port, User: asset.User}
	if asset.AuthType == "key" {
		cfg.KeyPath = asset.KeyPath
	} else {
		password, err := s.assets.GetPassword(asset.ID)
		if err != nil {
			return nil, fmt.Errorf("docker: 读取 SSH 密码失败: %w", err)
		}
		cfg.Password = password
	}
	return cfg, nil
}

func (s *DockerService) openSSH(opts DockerConnectOptions) (*dockerSession, error) {
	cfg, err := s.sshConfigFor(opts)
	if err != nil {
		return nil, err
	}
	sshSess, err := s.newSSH(cfg)
	if err != nil {
		return nil, fmt.Errorf("docker: SSH 连接失败: %w", err)
	}

	apiErr := error(nil)
	addr, tunnel, err := docker.StartUnixTunnel(sshSess, opts.RemoteSocket)
	if err != nil {
		apiErr = fmt.Errorf("启动隧道失败: %w", err)
	} else {
		client, perr := newTunnelClient(addr)
		if perr == nil {
			perr = pingDocker(client)
			if perr == nil {
				return &dockerSession{mode: dockerModeAPI, apiClient: client, tunnel: tunnel, ssh: sshSess}, nil
			}
			_ = client.Close()
		}
		apiErr = perr
		_ = tunnel.Close()
	}

	if !opts.CLIFallback {
		_ = sshSess.Close()
		return nil, fmt.Errorf("docker: 通过 SSH 隧道访问 %s 失败: %w", opts.RemoteSocket, apiErr)
	}
	// 仅验证 CLI 可用，不保留结果
	if _, cliErr := listContainersCLI(sshSess, false); cliErr != nil {
		_ = sshSess.Close()
		return nil, fmt.Errorf("docker: 隧道访问失败 (%v)，CLI 回退也失败: %w", apiErr, cliErr)
	}
	return &dockerSession{mode: dockerModeCLI, ssh: sshSess}, nil
}

func newTunnelClient(addr string) (*docker.Client, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}
	return docker.NewClient(docker.Config{Transport: docker.TransportTCP, Host: host, Port: port})
}

func listContainersCLI(sess dockerSSHSession, all bool) ([]docker.ContainerSummary, error) {
	if sess == nil {
		return nil, fmt.Errorf("docker: CLI 模式缺少 SSH 会话")
	}
	cmd := "docker ps --format '{{json .}}'"
	if all {
		cmd = "docker ps -a --format '{{json .}}'"
	}
	out, err := combinedOutputWithTimeout(sess, cmd, dockerOpTimeout)
	text := string(out)
	if err != nil {
		msg := strings.TrimSpace(text)
		if msg == "" {
			msg = err.Error()
		}
		return nil, fmt.Errorf("docker ps 执行失败: %s", msg)
	}
	// stderr 已合并进输出，仅保留 JSON 行，忽略警告等杂项。
	var jsonLines []string
	for _, line := range strings.Split(text, "\n") {
		if t := strings.TrimSpace(line); strings.HasPrefix(t, "{") {
			jsonLines = append(jsonLines, t)
		}
	}
	list, err := docker.ParseDockerPSJSONLines(strings.Join(jsonLines, "\n"))
	if err != nil {
		return nil, err
	}
	for i := range list {
		list[i].Via = dockerViaCLI
	}
	return list, nil
}

// combinedOutputWithTimeout 为远程命令设置超时，避免 SSH 通道无响应时永久阻塞。
// 超时后命令 goroutine 会在会话关闭（Close）时随连接断开而退出。
func combinedOutputWithTimeout(sess dockerSSHSession, cmd string, timeout time.Duration) ([]byte, error) {
	type result struct {
		out []byte
		err error
	}
	ch := make(chan result, 1)
	go func() {
		out, err := sess.CombinedOutput(cmd)
		ch <- result{out, err}
	}()
	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case r := <-ch:
		return r.out, r.err
	case <-timer.C:
		return nil, fmt.Errorf("docker: 远程命令执行超时 (%s)", timeout)
	}
}
