package docker

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ExecCreateOptions 对应 POST /containers/{id}/exec 请求体。
type ExecCreateOptions struct {
	Cmd          []string // 默认 ["/bin/sh"]
	AttachStdin  bool
	AttachStdout bool
	AttachStderr bool
	Tty          bool
}

type execCreateRequest struct {
	AttachStdin  bool     `json:"AttachStdin"`
	AttachStdout bool     `json:"AttachStdout"`
	AttachStderr bool     `json:"AttachStderr"`
	Tty          bool     `json:"Tty"`
	Cmd          []string `json:"Cmd"`
}

type execCreateResponse struct {
	ID string `json:"Id"`
}

// CreateExec 在容器内创建 exec 实例，返回 execID。
func (c *Client) CreateExec(ctx context.Context, containerID string, opts ExecCreateOptions) (string, error) {
	cmd := opts.Cmd
	if len(cmd) == 0 {
		cmd = []string{"/bin/sh"}
	}
	payload, err := json.Marshal(execCreateRequest{
		AttachStdin:  opts.AttachStdin,
		AttachStdout: opts.AttachStdout,
		AttachStderr: opts.AttachStderr,
		Tty:          opts.Tty,
		Cmd:          cmd,
	})
	if err != nil {
		return "", err
	}
	reqURL := fmt.Sprintf("%s/containers/%s/exec", c.baseURL, url.PathEscape(containerID))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("docker: create exec failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	var out execCreateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.ID == "" {
		return "", fmt.Errorf("docker: create exec returned empty id")
	}
	return out.ID, nil
}

// ResizeExec 调整 exec TTY 尺寸。
func (c *Client) ResizeExec(ctx context.Context, execID string, cols, rows int) error {
	v := url.Values{}
	v.Set("h", strconv.Itoa(rows))
	v.Set("w", strconv.Itoa(cols))
	reqURL := fmt.Sprintf("%s/exec/%s/resize?%s", c.baseURL, url.PathEscape(execID), v.Encode())
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("docker: resize exec failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	return nil
}

// StartExec 启动 exec 并通过 HTTP Upgrade 劫持连接，返回 TTY 原始双向流（无 8 字节多路复用头）。
// 启动成功后会尽力按 cols/rows 调整一次尺寸（失败不影响返回）。
func (c *Client) StartExec(ctx context.Context, execID string, cols, rows int) (io.ReadWriteCloser, error) {
	payload, err := json.Marshal(map[string]bool{"Detach": false, "Tty": true})
	if err != nil {
		return nil, err
	}
	reqURL := fmt.Sprintf("%s/exec/%s/start", c.baseURL, url.PathEscape(execID))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "tcp")

	conn, err := c.dialRaw(ctx, req.URL)
	if err != nil {
		return nil, err
	}
	rwc, err := doExecStartHandshake(ctx, conn, req)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	if cols > 0 && rows > 0 {
		rctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		_ = c.ResizeExec(rctx, execID, cols, rows)
		cancel()
	}
	return rwc, nil
}

// doExecStartHandshake 写出请求并读取响应头；成功时返回带缓冲读取的连接。
func doExecStartHandshake(ctx context.Context, conn net.Conn, req *http.Request) (io.ReadWriteCloser, error) {
	if dl, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(dl)
	}
	stop := make(chan struct{})
	defer close(stop)
	go func() {
		select {
		case <-ctx.Done():
			_ = conn.SetDeadline(time.Unix(1, 0))
		case <-stop:
		}
	}()

	if err := req.Write(conn); err != nil {
		return nil, ctxErrOr(ctx, err)
	}
	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, req)
	if err != nil {
		return nil, ctxErrOr(ctx, err)
	}
	if err := parseExecStartResponse(resp); err != nil {
		return nil, err
	}
	// 握手完成，清除 deadline
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}
	_ = conn.SetDeadline(time.Time{})
	return &hijackedConn{Conn: conn, br: br}, nil
}

func ctxErrOr(ctx context.Context, err error) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return err
}

// parseExecStartResponse 校验 exec start 响应：101 (Upgrade) 或 200 视为成功，其余返回含响应体的错误。
func parseExecStartResponse(resp *http.Response) error {
	if resp.StatusCode == http.StatusSwitchingProtocols || resp.StatusCode == http.StatusOK {
		return nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	return fmt.Errorf("docker: start exec failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
}

// hijackedConn 先消费 bufio 中已缓冲的数据，再读底层连接。
type hijackedConn struct {
	net.Conn
	br *bufio.Reader
}

func (h *hijackedConn) Read(p []byte) (int, error) {
	return h.br.Read(p)
}

// dialRaw 复用 Client 传输层（unix socket / tcp / tls）建立原始连接。
func (c *Client) dialRaw(ctx context.Context, u *url.URL) (net.Conn, error) {
	addr := u.Host
	useTLS := u.Scheme == "https"
	if _, _, err := net.SplitHostPort(addr); err != nil {
		if useTLS {
			addr = net.JoinHostPort(u.Hostname(), "443")
		} else {
			addr = net.JoinHostPort(u.Hostname(), "80")
		}
	}
	var conn net.Conn
	var err error
	if c.transport != nil && c.transport.DialContext != nil {
		conn, err = c.transport.DialContext(ctx, "tcp", addr)
	} else {
		var d net.Dialer
		conn, err = d.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return nil, err
	}
	if !useTLS {
		return conn, nil
	}
	var cfg *tls.Config
	if c.transport != nil && c.transport.TLSClientConfig != nil {
		cfg = c.transport.TLSClientConfig.Clone()
	} else {
		cfg = &tls.Config{}
	}
	if cfg.ServerName == "" {
		cfg.ServerName = u.Hostname()
	}
	tlsConn := tls.Client(conn, cfg)
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return tlsConn, nil
}
