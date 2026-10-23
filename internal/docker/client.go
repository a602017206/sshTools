package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	transport  *http.Transport
}

func NewClient(cfg Config) (*Client, error) {
	switch cfg.Transport {
	case TransportTCP:
		return newTCPClient(cfg)
	case TransportLocal:
		return newLocalClient(cfg)
	case TransportSSH:
		return nil, fmt.Errorf("docker: ssh transport is not implemented")
	default:
		return nil, fmt.Errorf("docker: unknown transport %q", cfg.Transport)
	}
}

func newTCPClient(cfg Config) (*Client, error) {
	if cfg.Host == "" {
		return nil, fmt.Errorf("docker: tcp transport requires Host")
	}
	if cfg.Port <= 0 {
		return nil, fmt.Errorf("docker: tcp transport requires Port")
	}
	scheme := "http"
	if cfg.TLS {
		scheme = "https"
	}
	hostPort := net.JoinHostPort(cfg.Host, fmt.Sprintf("%d", cfg.Port))
	baseURL := fmt.Sprintf("%s://%s", scheme, hostPort)
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Transport: transport},
		transport:  transport,
	}, nil
}

func newLocalClient(cfg Config) (*Client, error) {
	socketPath, err := resolveSocketPath(cfg)
	if err != nil {
		return nil, err
	}
	dialer := &net.Dialer{Timeout: 30 * time.Second}
	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.DialContext(ctx, "unix", socketPath)
		},
	}
	return &Client{
		baseURL:    "http://localhost",
		httpClient: &http.Client{Transport: transport},
		transport:  transport,
	}, nil
}

func resolveSocketPath(cfg Config) (string, error) {
	if cfg.SocketPath != "" {
		return cfg.SocketPath, nil
	}
	for _, raw := range []string{cfg.DockerHost, os.Getenv("DOCKER_HOST")} {
		if raw == "" {
			continue
		}
		path, err := unixPathFromDockerHost(raw)
		if err != nil {
			return "", err
		}
		if path != "" {
			return path, nil
		}
	}
	return defaultSocketPath(), nil
}

func unixPathFromDockerHost(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("docker: invalid DOCKER_HOST %q: %w", raw, err)
	}
	switch u.Scheme {
	case "unix":
		return u.Path, nil
	case "":
		return raw, nil
	default:
		return "", nil
	}
}

func defaultSocketPath() string {
	if runtime.GOOS == "windows" {
		return `\\.\pipe\docker_engine`
	}
	return "/var/run/docker.sock"
}

func (c *Client) Close() error {
	if c.transport != nil {
		c.transport.CloseIdleConnections()
	}
	return nil
}

func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/_ping", nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("docker: ping failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

type containerJSON struct {
	ID      string              `json:"Id"`
	Names   []string            `json:"Names"`
	Image   string              `json:"Image"`
	Status  string              `json:"Status"`
	Created int64               `json:"Created"`
	Ports   []containerPortJSON `json:"Ports"`
}

type containerPortJSON struct {
	IP          string `json:"IP,omitempty"`
	PrivatePort int    `json:"PrivatePort"`
	PublicPort  int    `json:"PublicPort,omitempty"`
	Type        string `json:"Type"`
}

func (c *Client) ListContainers(ctx context.Context, all bool) ([]ContainerSummary, error) {
	allParam := "0"
	if all {
		allParam = "1"
	}
	reqURL := fmt.Sprintf("%s/containers/json?all=%s", c.baseURL, allParam)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return nil, fmt.Errorf("docker: list containers failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	var raw []containerJSON
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	out := make([]ContainerSummary, 0, len(raw))
	for _, item := range raw {
		out = append(out, mapContainerSummary(item))
	}
	return out, nil
}

func mapContainerSummary(item containerJSON) ContainerSummary {
	return ContainerSummary{
		ID:      shortContainerID(item.ID),
		Name:    containerName(item.Names),
		Image:   item.Image,
		Status:  item.Status,
		Ports:   formatContainerPorts(item.Ports),
		Created: item.Created,
	}
}

func shortContainerID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

func containerName(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return strings.TrimPrefix(names[0], "/")
}

func formatContainerPorts(ports []containerPortJSON) string {
	if len(ports) == 0 {
		return ""
	}
	parts := make([]string, 0, len(ports))
	for _, p := range ports {
		parts = append(parts, formatContainerPort(p))
	}
	return strings.Join(parts, ", ")
}

func formatContainerPort(p containerPortJSON) string {
	typ := p.Type
	if typ == "" {
		typ = "tcp"
	}
	if p.PublicPort > 0 {
		return fmt.Sprintf("%d:%d/%s", p.PublicPort, p.PrivatePort, typ)
	}
	return fmt.Sprintf("%d/%s", p.PrivatePort, typ)
}
