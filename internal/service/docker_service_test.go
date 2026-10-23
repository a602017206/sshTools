package service

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"AHaSSHTools/internal/config"
	"AHaSSHTools/internal/ssh"
)

type fakeDockerAssets struct {
	conns     map[string]config.ConnectionConfig
	passwords map[string]string
}

func (f *fakeDockerAssets) GetConnection(id string) (config.ConnectionConfig, error) {
	c, ok := f.conns[id]
	if !ok {
		return config.ConnectionConfig{}, fmt.Errorf("connection %s not found", id)
	}
	return c, nil
}

func (f *fakeDockerAssets) GetPassword(id string) (string, error) {
	return f.passwords[id], nil
}

type fakeDockerSSH struct {
	mu        sync.Mutex
	dialErr   error
	dialAddr  string // when set, Dial connects here (ignores network/addr)
	output    string
	outputErr error
	commands  []string
	closed    int
}

func (f *fakeDockerSSH) Dial(network, addr string) (net.Conn, error) {
	if f.dialErr != nil {
		return nil, f.dialErr
	}
	return net.Dial("tcp", f.dialAddr)
}

func (f *fakeDockerSSH) CombinedOutput(cmd string) ([]byte, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.commands = append(f.commands, cmd)
	return []byte(f.output), f.outputErr
}

func (f *fakeDockerSSH) Close() error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed++
	return nil
}

func newTestDockerService(sess *fakeDockerSSH, gotCfg **ssh.Config) *DockerService {
	assets := &fakeDockerAssets{
		conns: map[string]config.ConnectionConfig{
			"ssh-1":   {ID: "ssh-1", Host: "10.0.0.1", Port: 2222, User: "root", AuthType: "password"},
			"ssh-key": {ID: "ssh-key", Host: "10.0.0.2", Port: 22, User: "dev", AuthType: "key", KeyPath: "/k/id"},
		},
		passwords: map[string]string{"ssh-1": "secret"},
	}
	return newDockerServiceWithFactory(assets, func(cfg *ssh.Config) (dockerSSHSession, error) {
		if gotCfg != nil {
			*gotCfg = cfg
		}
		return sess, nil
	})
}

func dockerSSHConn(extra map[string]string) config.ConnectionConfig {
	md := map[string]string{"transport": "ssh", "ssh_connection_id": "ssh-1"}
	for k, v := range extra {
		md[k] = v
	}
	return config.ConnectionConfig{ID: "d1", Name: "d", Type: "docker", Metadata: md}
}

const cliJSONLine = `{"ID":"abcdef1234567890","Names":"/web","Image":"nginx","Status":"Up 1 minute","Ports":"80/tcp","CreatedAt":"2024-01-02 03:04:05 +0000 UTC"}` + "\n"

func TestDockerSSHDialFailFallsBackToCLI(t *testing.T) {
	sess := &fakeDockerSSH{dialErr: errors.New("administratively prohibited"), output: cliJSONLine}
	var cfg *ssh.Config
	svc := newTestDockerService(sess, &cfg)

	if err := svc.ConnectDocker("s1", dockerSSHConn(nil)); err != nil {
		t.Fatalf("connect: %v", err)
	}
	if cfg == nil || cfg.Host != "10.0.0.1" || cfg.Port != 2222 || cfg.User != "root" || cfg.Password != "secret" {
		t.Fatalf("ssh cfg: %+v", cfg)
	}
	list, err := svc.ListDockerContainers("s1", true)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 || list[0].ID != "abcdef123456" || list[0].Name != "web" {
		t.Fatalf("list: %+v", list)
	}
	last := sess.commands[len(sess.commands)-1]
	if !strings.Contains(last, "docker ps -a") || !strings.Contains(last, "{{json .}}") {
		t.Fatalf("cmd: %q", last)
	}
	if mode := svc.sessionMode("s1"); mode != dockerModeCLI {
		t.Fatalf("mode=%q", mode)
	}

	if _, err := svc.ListDockerContainers("s1", false); err != nil {
		t.Fatal(err)
	}
	last = sess.commands[len(sess.commands)-1]
	if strings.Contains(last, " -a") {
		t.Fatalf("running-only cmd should not use -a: %q", last)
	}

	if err := svc.CloseDocker("s1"); err != nil {
		t.Fatal(err)
	}
	if sess.closed != 1 {
		t.Fatalf("ssh closed=%d", sess.closed)
	}
	if _, err := svc.ListDockerContainers("s1", true); err == nil {
		t.Fatal("expected error after close")
	}
}

func TestDockerCLIFallbackDisabledReturnsError(t *testing.T) {
	sess := &fakeDockerSSH{dialErr: errors.New("refused"), output: cliJSONLine}
	svc := newTestDockerService(sess, nil)
	conn := dockerSSHConn(map[string]string{"cli_fallback": "false"})

	if err := svc.TestDockerConnection(conn); err == nil {
		t.Fatal("expected test error")
	}
	if sess.closed == 0 {
		t.Fatal("ssh session must be closed on failure")
	}
	if len(sess.commands) != 0 {
		t.Fatalf("CLI must not run: %v", sess.commands)
	}
	if err := svc.ConnectDocker("s1", conn); err == nil {
		t.Fatal("expected connect error")
	}
	if svc.sessionMode("s1") != "" {
		t.Fatal("failed connect must not register session")
	}
}

func TestDockerCLIFallbackCommandFailure(t *testing.T) {
	sess := &fakeDockerSSH{dialErr: errors.New("refused"), output: "bash: docker: command not found", outputErr: errors.New("exit status 127")}
	svc := newTestDockerService(sess, nil)
	err := svc.TestDockerConnection(dockerSSHConn(nil))
	if err == nil || !strings.Contains(err.Error(), "command not found") {
		t.Fatalf("err=%v", err)
	}
}

func TestDockerSSHAPIPathViaTunnel(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/_ping":
			_, _ = w.Write([]byte("OK"))
		case "/containers/json":
			_, _ = w.Write([]byte(`[{"Id":"0123456789abcdef","Names":["/api"],"Image":"img","Status":"Up","Created":1}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	sess := &fakeDockerSSH{dialAddr: srv.Listener.Addr().String(), output: cliJSONLine}
	svc := newTestDockerService(sess, nil)

	if err := svc.ConnectDocker("s1", dockerSSHConn(nil)); err != nil {
		t.Fatalf("connect: %v", err)
	}
	if mode := svc.sessionMode("s1"); mode != dockerModeAPI {
		t.Fatalf("mode=%q", mode)
	}
	list, err := svc.ListDockerContainers("s1", true)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Name != "api" || list[0].ID != "0123456789ab" {
		t.Fatalf("list: %+v", list)
	}
	if len(sess.commands) != 0 {
		t.Fatalf("CLI should not be used: %v", sess.commands)
	}
	if err := svc.CloseDocker("s1"); err != nil {
		t.Fatal(err)
	}
	if sess.closed != 1 {
		t.Fatalf("closed=%d", sess.closed)
	}
}

func TestDockerSSHMissingAssetErrors(t *testing.T) {
	svc := newTestDockerService(&fakeDockerSSH{}, nil)
	conn := dockerSSHConn(map[string]string{"ssh_connection_id": "nope"})
	if err := svc.TestDockerConnection(conn); err == nil {
		t.Fatal("expected error")
	}
	conn = dockerSSHConn(map[string]string{"ssh_connection_id": ""})
	if err := svc.TestDockerConnection(conn); err == nil {
		t.Fatal("expected error for empty ssh_connection_id")
	}
}

func TestDockerSSHKeyAssetUsesKeyPath(t *testing.T) {
	sess := &fakeDockerSSH{dialErr: errors.New("x"), output: cliJSONLine}
	var cfg *ssh.Config
	svc := newTestDockerService(sess, &cfg)
	conn := dockerSSHConn(map[string]string{"ssh_connection_id": "ssh-key"})
	if err := svc.TestDockerConnection(conn); err != nil {
		t.Fatal(err)
	}
	if cfg.KeyPath != "/k/id" || cfg.Password != "" {
		t.Fatalf("%+v", cfg)
	}
}

func TestDockerReconnectReplacesSession(t *testing.T) {
	first := &fakeDockerSSH{dialErr: errors.New("x"), output: cliJSONLine}
	svc := newTestDockerService(first, nil)
	if err := svc.ConnectDocker("s1", dockerSSHConn(nil)); err != nil {
		t.Fatal(err)
	}
	if err := svc.ConnectDocker("s1", dockerSSHConn(nil)); err != nil {
		t.Fatal(err)
	}
	if first.closed != 1 {
		t.Fatalf("old session should be closed once, got %d", first.closed)
	}
}

func TestDockerCloseUnknownSessionIsNoop(t *testing.T) {
	svc := newTestDockerService(&fakeDockerSSH{}, nil)
	if err := svc.CloseDocker("missing"); err != nil {
		t.Fatal(err)
	}
}

func TestParseDockerOptionsTransportDefaults(t *testing.T) {
	cases := []struct {
		name string
		conn config.ConnectionConfig
		want string
	}{
		{"ssh by id", config.ConnectionConfig{Metadata: map[string]string{"ssh_connection_id": "a"}}, "ssh"},
		{"tcp by host", config.ConnectionConfig{Host: "h", Port: 2375}, "tcp"},
		{"local", config.ConnectionConfig{}, "local"},
		{"explicit", config.ConnectionConfig{Host: "h", Metadata: map[string]string{"transport": "local"}}, "local"},
	}
	for _, c := range cases {
		if got := parseDockerOptions(c.conn).Transport; got != c.want {
			t.Errorf("%s: got %q want %q", c.name, got, c.want)
		}
	}
	if !parseDockerOptions(config.ConnectionConfig{}).CLIFallback {
		t.Error("cli_fallback default must be true")
	}
	off := parseDockerOptions(config.ConnectionConfig{Metadata: map[string]string{"cli_fallback": "false"}})
	if off.CLIFallback {
		t.Error("cli_fallback=false must disable")
	}
	if got := parseDockerOptions(config.ConnectionConfig{}).RemoteSocket; got != "/var/run/docker.sock" {
		t.Errorf("remote socket default: %q", got)
	}
}

func startUnixDockerServer(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("/tmp", "dk")
	if err != nil {
		t.Skip("cannot create /tmp dir")
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	sock := filepath.Join(dir, "d.sock")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Skipf("unix sockets unavailable: %v", err)
	}
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/_ping":
			_, _ = w.Write([]byte("OK"))
		case "/containers/json":
			_, _ = w.Write([]byte(`[{"Id":"fedcba9876543210","Names":["/loc"],"Image":"i","Status":"Up"}]`))
		default:
			http.NotFound(w, r)
		}
	}))
	srv.Listener.Close()
	srv.Listener = ln
	srv.Start()
	t.Cleanup(srv.Close)
	return sock
}

func TestDetectLocalDockerOK(t *testing.T) {
	sock := startUnixDockerServer(t)
	t.Setenv("DOCKER_HOST", "unix://"+sock)
	svc := NewDockerService(nil)
	ok, detail, err := svc.DetectLocalDocker()
	if err != nil || !ok || detail == "" {
		t.Fatalf("ok=%v detail=%q err=%v", ok, detail, err)
	}
}

func TestDetectLocalDockerUnavailable(t *testing.T) {
	t.Setenv("DOCKER_HOST", "unix:///nonexistent/dir/docker.sock")
	svc := NewDockerService(nil)
	ok, detail, err := svc.DetectLocalDocker()
	if err != nil {
		t.Fatal(err)
	}
	if ok || detail == "" {
		t.Fatalf("ok=%v detail=%q", ok, detail)
	}
}

func TestDockerLocalConnectAndList(t *testing.T) {
	sock := startUnixDockerServer(t)
	svc := NewDockerService(nil)
	conn := config.ConnectionConfig{ID: "l", Type: "docker", Metadata: map[string]string{"transport": "local", "socket_path": sock}}
	if err := svc.ConnectDocker("L", conn); err != nil {
		t.Fatal(err)
	}
	list, err := svc.ListDockerContainers("L", false)
	if err != nil || len(list) != 1 || list[0].Name != "loc" {
		t.Fatalf("list=%+v err=%v", list, err)
	}
	if err := svc.CloseDocker("L"); err != nil {
		t.Fatal(err)
	}
}

func TestDockerTCPConnect(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/_ping" {
			_, _ = w.Write([]byte("OK"))
			return
		}
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()
	host, portStr, _ := net.SplitHostPort(srv.Listener.Addr().String())
	var port int
	fmt.Sscanf(portStr, "%d", &port)
	svc := NewDockerService(nil)
	conn := config.ConnectionConfig{Host: host, Port: port, Type: "docker"}
	if err := svc.TestDockerConnection(conn); err != nil {
		t.Fatal(err)
	}
	conn.Port = 1
	if err := svc.TestDockerConnection(conn); err == nil {
		t.Fatal("expected tcp failure")
	}
}
