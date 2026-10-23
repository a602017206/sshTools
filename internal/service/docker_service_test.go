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
	"time"

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
	if list[0].Via != "cli" {
		t.Fatalf("CLI 模式应标记 via=cli: %+v", list[0])
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

func TestDockerCloseAllClosesEverySession(t *testing.T) {
	sess := &fakeDockerSSH{dialErr: errors.New("x"), output: cliJSONLine}
	svc := newTestDockerService(sess, nil)
	if err := svc.ConnectDocker("a", dockerSSHConn(nil)); err != nil {
		t.Fatal(err)
	}
	if err := svc.ConnectDocker("b", dockerSSHConn(nil)); err != nil {
		t.Fatal(err)
	}
	if err := svc.CloseAll(); err != nil {
		t.Fatal(err)
	}
	if sess.closed != 2 {
		t.Fatalf("ssh closed=%d, want 2", sess.closed)
	}
	if _, err := svc.ListDockerContainers("a", true); err == nil {
		t.Fatal("CloseAll 后会话应不存在")
	}
}

type blockingDockerSSH struct{ fakeDockerSSH }

func (b *blockingDockerSSH) CombinedOutput(string) ([]byte, error) {
	select {}
}

func TestCombinedOutputWithTimeout(t *testing.T) {
	start := time.Now()
	_, err := combinedOutputWithTimeout(&blockingDockerSSH{}, "docker ps", 50*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "超时") {
		t.Fatalf("expected timeout error, got %v", err)
	}
	if time.Since(start) > 2*time.Second {
		t.Fatal("timeout took too long")
	}
}

// ---- Phase 2 Task 4: 生命周期 / Inspect / 一次性日志 ----

func newCLIDockerService(t *testing.T, sess *fakeDockerSSH) *DockerService {
	t.Helper()
	svc := newTestDockerService(sess, nil)
	if err := svc.ConnectDocker("s1", dockerSSHConn(nil)); err != nil {
		t.Fatalf("connect: %v", err)
	}
	if svc.sessionMode("s1") != dockerModeCLI {
		t.Fatalf("mode=%q", svc.sessionMode("s1"))
	}
	sess.commands = nil
	return svc
}

func TestDockerCLILifecycleCommands(t *testing.T) {
	sess := &fakeDockerSSH{dialErr: errors.New("x"), output: cliJSONLine}
	svc := newCLIDockerService(t, sess)
	sess.output = ""

	if err := svc.StartDockerContainer("s1", "abc"); err != nil {
		t.Fatal(err)
	}
	if err := svc.StopDockerContainer("s1", "abc"); err != nil {
		t.Fatal(err)
	}
	if err := svc.RestartDockerContainer("s1", "abc"); err != nil {
		t.Fatal(err)
	}
	if err := svc.RemoveDockerContainer("s1", "abc", false); err != nil {
		t.Fatal(err)
	}
	if err := svc.RemoveDockerContainer("s1", "abc", true); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"docker start 'abc'",
		"docker stop 'abc'",
		"docker restart 'abc'",
		"docker rm 'abc'",
		"docker rm -f 'abc'",
	}
	if strings.Join(sess.commands, "|") != strings.Join(want, "|") {
		t.Fatalf("commands=%v", sess.commands)
	}
}

func TestDockerCLILifecycleErrorSurfacesOutput(t *testing.T) {
	sess := &fakeDockerSSH{dialErr: errors.New("x"), output: cliJSONLine}
	svc := newCLIDockerService(t, sess)
	sess.output = "Error: No such container: nope"
	sess.outputErr = errors.New("exit status 1")
	err := svc.StartDockerContainer("s1", "nope")
	if err == nil || !strings.Contains(err.Error(), "No such container") {
		t.Fatalf("err=%v", err)
	}
}

func TestDockerCLIInspectAndLogs(t *testing.T) {
	sess := &fakeDockerSSH{dialErr: errors.New("x"), output: cliJSONLine}
	svc := newCLIDockerService(t, sess)

	sess.output = `[{"Id":"abcdef1234567890","Name":"/web","Config":{"Image":"nginx"},"State":{"Status":"running"}}]`
	detail, err := svc.InspectDockerContainer("s1", "abc")
	if err != nil {
		t.Fatal(err)
	}
	if detail.Name != "web" || detail.Image != "nginx" || detail.State != "running" || detail.ID != "abcdef123456" {
		t.Fatalf("detail=%+v", detail)
	}
	if got := sess.commands[len(sess.commands)-1]; got != "docker inspect 'abc'" {
		t.Fatalf("cmd=%q", got)
	}

	sess.output = "line1\nline2\n"
	logs, err := svc.GetDockerContainerLogs("s1", "abc", 200)
	if err != nil || logs != "line1\nline2\n" {
		t.Fatalf("logs=%q err=%v", logs, err)
	}
	if got := sess.commands[len(sess.commands)-1]; got != "docker logs --tail 200 'abc'" {
		t.Fatalf("cmd=%q", got)
	}
}

func TestDockerContainerOpsUnknownSession(t *testing.T) {
	svc := newTestDockerService(&fakeDockerSSH{}, nil)
	if err := svc.StartDockerContainer("nope", "a"); err == nil {
		t.Error("start: expected error")
	}
	if err := svc.StopDockerContainer("nope", "a"); err == nil {
		t.Error("stop: expected error")
	}
	if err := svc.RestartDockerContainer("nope", "a"); err == nil {
		t.Error("restart: expected error")
	}
	if err := svc.RemoveDockerContainer("nope", "a", true); err == nil {
		t.Error("remove: expected error")
	}
	if _, err := svc.InspectDockerContainer("nope", "a"); err == nil {
		t.Error("inspect: expected error")
	}
	if _, err := svc.GetDockerContainerLogs("nope", "a", 10); err == nil {
		t.Error("logs: expected error")
	}
}

func TestDockerContainerOpsRejectEmptyID(t *testing.T) {
	sess := &fakeDockerSSH{dialErr: errors.New("x"), output: cliJSONLine}
	svc := newCLIDockerService(t, sess)
	if err := svc.StartDockerContainer("s1", "  "); err == nil {
		t.Fatal("expected error for empty id")
	}
	if len(sess.commands) != 0 {
		t.Fatalf("no command expected: %v", sess.commands)
	}
}

func TestDockerAPILifecycleInspectLogs(t *testing.T) {
	var mu sync.Mutex
	var calls []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		calls = append(calls, r.Method+" "+r.URL.RequestURI())
		mu.Unlock()
		switch {
		case r.URL.Path == "/_ping":
			_, _ = w.Write([]byte("OK"))
		case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/containers/"):
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case r.URL.Path == "/containers/abc/json":
			_, _ = w.Write([]byte(`{"Id":"abcdef1234567890","Name":"/api","Config":{"Image":"img"},"State":{"Status":"running"}}`))
		case r.URL.Path == "/containers/abc/logs":
			// Docker 多路复用帧：stdout, 长度 3
			_, _ = w.Write([]byte{1, 0, 0, 0, 0, 0, 0, 3, 'h', 'i', '\n'})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	host, portStr, _ := net.SplitHostPort(srv.Listener.Addr().String())
	var port int
	fmt.Sscanf(portStr, "%d", &port)
	svc := NewDockerService(nil)
	if err := svc.ConnectDocker("t", config.ConnectionConfig{Host: host, Port: port, Type: "docker"}); err != nil {
		t.Fatal(err)
	}
	if err := svc.StartDockerContainer("t", "abc"); err != nil {
		t.Fatal(err)
	}
	if err := svc.StopDockerContainer("t", "abc"); err != nil {
		t.Fatal(err)
	}
	if err := svc.RestartDockerContainer("t", "abc"); err != nil {
		t.Fatal(err)
	}
	if err := svc.RemoveDockerContainer("t", "abc", true); err != nil {
		t.Fatal(err)
	}
	detail, err := svc.InspectDockerContainer("t", "abc")
	if err != nil || detail.Name != "api" || detail.Image != "img" {
		t.Fatalf("detail=%+v err=%v", detail, err)
	}
	logs, err := svc.GetDockerContainerLogs("t", "abc", 50)
	if err != nil || logs != "hi\n" {
		t.Fatalf("logs=%q err=%v", logs, err)
	}
	mu.Lock()
	defer mu.Unlock()
	joined := strings.Join(calls, "\n")
	for _, w := range []string{"POST /containers/abc/start", "POST /containers/abc/stop", "POST /containers/abc/restart", "DELETE /containers/abc?force=1", "tail=50"} {
		if !strings.Contains(joined, w) {
			t.Errorf("missing call %q in:\n%s", w, joined)
		}
	}
}
