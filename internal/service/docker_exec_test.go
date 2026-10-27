package service

import (
	"bufio"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

// execCollector 线程安全地收集 emit 的数据。
type execCollector struct {
	mu   sync.Mutex
	data []byte
}

func (c *execCollector) emit(b []byte) {
	c.mu.Lock()
	c.data = append(c.data, b...)
	c.mu.Unlock()
}

func (c *execCollector) String() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return string(c.data)
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for %s", what)
}

func connectCLIDocker(t *testing.T) (*DockerService, *fakeDockerSSH) {
	t.Helper()
	sess := &fakeDockerSSH{dialErr: errors.New("no socket"), output: cliJSONLine}
	svc := newTestDockerService(sess, nil)
	if err := svc.ConnectDocker("s1", dockerSSHConn(nil)); err != nil {
		t.Fatalf("connect: %v", err)
	}
	if svc.sessionMode("s1") != dockerModeCLI {
		t.Fatal("expected CLI mode")
	}
	return svc, sess
}

func TestDockerExecCLIStartSendResizeClose(t *testing.T) {
	svc, sess := connectCLIDocker(t)
	var out execCollector
	closed := make(chan string, 4)
	svc.SetDockerExecClosedHandler(func(id string) { closed <- id })

	id, err := svc.StartDockerExec("s1", "abc123", 100, 30, out.emit)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	if !strings.HasPrefix(id, "dexec-") {
		t.Fatalf("id=%q", id)
	}
	last := sess.commands[len(sess.commands)-1]
	if !strings.Contains(last, "docker exec -i -t") || !strings.Contains(last, "abc123") || !strings.HasSuffix(last, "/bin/sh") {
		t.Fatalf("cmd=%q", last)
	}
	pty := sess.ptys[0]
	if pty.cols != 100 || pty.rows != 30 {
		t.Fatalf("初始尺寸 %dx%d", pty.cols, pty.rows)
	}

	// 输出
	go func() { _, _ = pty.fakeStream.pw.Write([]byte("hello\x1b[0m")) }()
	waitFor(t, "emit", func() bool { return out.String() == "hello\x1b[0m" })

	// 输入
	if err := svc.SendDockerExecData(id, []byte("ls\n")); err != nil {
		t.Fatalf("send: %v", err)
	}
	pty.mu.Lock()
	written := string(pty.written)
	pty.mu.Unlock()
	if written != "ls\n" {
		t.Fatalf("written=%q", written)
	}

	// resize
	if err := svc.ResizeDockerExec(id, 120, 40); err != nil {
		t.Fatalf("resize: %v", err)
	}
	pty.mu.Lock()
	cols, rows := pty.cols, pty.rows
	pty.mu.Unlock()
	if cols != 120 || rows != 40 {
		t.Fatalf("resize 后 %dx%d", cols, rows)
	}
	if err := svc.ResizeDockerExec(id, 0, 10); err == nil {
		t.Fatal("非法尺寸应报错")
	}

	// 关闭
	if err := svc.CloseDockerExec(id); err != nil {
		t.Fatalf("close: %v", err)
	}
	if !pty.isClosed() {
		t.Fatal("流应被关闭")
	}
	select {
	case got := <-closed:
		if got != id {
			t.Fatalf("closed id=%q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("应收到 exec 结束回调")
	}
	if err := svc.SendDockerExecData(id, []byte("x")); err == nil {
		t.Fatal("关闭后发送应失败")
	}
	// 重复关闭无副作用
	if err := svc.CloseDockerExec(id); err != nil {
		t.Fatal(err)
	}
}

func TestDockerExecSendFailsAfterCloseDocker(t *testing.T) {
	svc, sess := connectCLIDocker(t)
	var out execCollector
	id, err := svc.StartDockerExec("s1", "abc", 80, 24, out.emit)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.SendDockerExecData(id, []byte("a")); err != nil {
		t.Fatal(err)
	}
	if err := svc.CloseDocker("s1"); err != nil {
		t.Fatal(err)
	}
	if !sess.ptys[0].isClosed() {
		t.Fatal("CloseDocker 应关闭子 exec 流")
	}
	if err := svc.SendDockerExecData(id, []byte("b")); err == nil {
		t.Fatal("CloseDocker 后 Send 应失败")
	}
	if err := svc.ResizeDockerExec(id, 80, 24); err == nil {
		t.Fatal("CloseDocker 后 Resize 应失败")
	}
	if _, err := svc.StartDockerExec("s1", "abc", 80, 24, out.emit); err == nil {
		t.Fatal("会话关闭后不应再能启动 exec")
	}
}

func TestDockerExecCloseAllStopsExecs(t *testing.T) {
	svc, sess := connectCLIDocker(t)
	var out execCollector
	id1, err := svc.StartDockerExec("s1", "a", 80, 24, out.emit)
	if err != nil {
		t.Fatal(err)
	}
	id2, err := svc.StartDockerExec("s1", "b", 80, 24, out.emit)
	if err != nil {
		t.Fatal(err)
	}
	if id1 == id2 {
		t.Fatal("exec id 应唯一")
	}
	if err := svc.CloseAll(); err != nil {
		t.Fatal(err)
	}
	for i, p := range sess.ptys {
		if !p.isClosed() {
			t.Fatalf("pty %d 未关闭", i)
		}
	}
	if err := svc.SendDockerExecData(id1, []byte("x")); err == nil {
		t.Fatal("CloseAll 后 Send 应失败")
	}
}

func TestDockerExecOnlyClosesOwnSession(t *testing.T) {
	sess := &fakeDockerSSH{dialErr: errors.New("x"), output: cliJSONLine}
	svc := newTestDockerService(sess, nil)
	for _, sid := range []string{"s1", "s2"} {
		if err := svc.ConnectDocker(sid, dockerSSHConn(nil)); err != nil {
			t.Fatal(err)
		}
	}
	var out execCollector
	id1, _ := svc.StartDockerExec("s1", "a", 80, 24, out.emit)
	id2, err := svc.StartDockerExec("s2", "b", 80, 24, out.emit)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.CloseDocker("s1"); err != nil {
		t.Fatal(err)
	}
	if err := svc.SendDockerExecData(id1, []byte("x")); err == nil {
		t.Fatal("s1 的 exec 应失效")
	}
	if err := svc.SendDockerExecData(id2, []byte("x")); err != nil {
		t.Fatalf("s2 的 exec 不应受影响: %v", err)
	}
}

func TestDockerExecEOFRemovesExec(t *testing.T) {
	svc, sess := connectCLIDocker(t)
	var out execCollector
	closed := make(chan string, 1)
	svc.SetDockerExecClosedHandler(func(id string) { closed <- id })
	id, err := svc.StartDockerExec("s1", "a", 80, 24, out.emit)
	if err != nil {
		t.Fatal(err)
	}
	_ = sess.ptys[0].fakeStream.pw.Close() // 远端结束 → EOF
	select {
	case <-closed:
	case <-time.After(2 * time.Second):
		t.Fatal("EOF 后应触发结束回调")
	}
	waitFor(t, "exec removed", func() bool {
		_, err := svc.lookupExec(id)
		return err != nil
	})
	if !sess.ptys[0].isClosed() {
		t.Fatal("EOF 后应关闭流")
	}
}

func TestDockerExecValidation(t *testing.T) {
	svc, sess := connectCLIDocker(t)
	var out execCollector
	if _, err := svc.StartDockerExec("nope", "a", 80, 24, out.emit); err == nil {
		t.Fatal("未知会话应报错")
	}
	if _, err := svc.StartDockerExec("s1", "  ", 80, 24, out.emit); err == nil {
		t.Fatal("空容器 ID 应报错")
	}
	if _, err := svc.StartDockerExec("s1", "a", 80, 24, nil); err == nil {
		t.Fatal("emit 为空应报错")
	}
	sess.streamErr = errors.New("pty denied")
	if _, err := svc.StartDockerExec("s1", "a", 80, 24, out.emit); err == nil || !strings.Contains(err.Error(), "pty denied") {
		t.Fatalf("err=%v", err)
	}
	if err := svc.SendDockerExecData("dexec-none", []byte("x")); err == nil {
		t.Fatal("未知 exec 发送应报错")
	}
	if err := svc.CloseDockerExec("dexec-none"); err != nil {
		t.Fatalf("未知 exec 关闭应为 no-op: %v", err)
	}
}

// fakeExecAPIServer 模拟 Docker Engine exec 端点：create → 201，start → 101 Upgrade 后回显。
func fakeExecAPIServer(t *testing.T, resizeCh chan<- string, createBody chan<- string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/_ping":
			_, _ = w.Write([]byte("OK"))
		case r.Method == http.MethodPost && strings.HasSuffix(r.URL.Path, "/containers/cid1/exec"):
			b, _ := io.ReadAll(r.Body)
			createBody <- string(b)
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"Id":"exec-1"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/exec/exec-1/start":
			_, _ = io.Copy(io.Discard, r.Body) // 先消费请求体，避免混入回显
			hj, ok := w.(http.Hijacker)
			if !ok {
				http.Error(w, "no hijack", 500)
				return
			}
			conn, brw, err := hj.Hijack()
			if err != nil {
				return
			}
			defer conn.Close()
			_, _ = conn.Write([]byte("HTTP/1.1 101 UPGRADED\r\nContent-Type: application/vnd.docker.raw-stream\r\nConnection: Upgrade\r\nUpgrade: tcp\r\n\r\n"))
			_, _ = conn.Write([]byte("$ "))
			rd := bufio.NewReader(brw)
			buf := make([]byte, 256)
			for {
				n, err := rd.Read(buf)
				if n > 0 {
					_, _ = conn.Write(append([]byte("echo:"), buf[:n]...))
				}
				if err != nil {
					return
				}
			}
		case r.Method == http.MethodPost && r.URL.Path == "/exec/exec-1/resize":
			select {
			case resizeCh <- r.URL.RawQuery:
			default:
			}
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
}

func TestDockerExecAPIStartSendResizeClose(t *testing.T) {
	resizeCh := make(chan string, 8)
	createBody := make(chan string, 1)
	srv := fakeExecAPIServer(t, resizeCh, createBody)
	defer srv.Close()

	sess := &fakeDockerSSH{dialAddr: srv.Listener.Addr().String(), output: cliJSONLine}
	svc := newTestDockerService(sess, nil)
	if err := svc.ConnectDocker("s1", dockerSSHConn(nil)); err != nil {
		t.Fatal(err)
	}
	if svc.sessionMode("s1") != dockerModeAPI {
		t.Fatal("expected API mode")
	}

	var out execCollector
	id, err := svc.StartDockerExec("s1", "cid1", 90, 28, out.emit)
	if err != nil {
		t.Fatalf("start: %v", err)
	}
	body := <-createBody
	for _, want := range []string{`"AttachStdin":true`, `"AttachStdout":true`, `"AttachStderr":true`, `"Tty":true`, `"/bin/sh"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("create body 缺少 %s: %s", want, body)
		}
	}
	waitFor(t, "initial prompt", func() bool { return out.String() == "$ " })

	if err := svc.SendDockerExecData(id, []byte("ls\n")); err != nil {
		t.Fatalf("send: %v", err)
	}
	waitFor(t, "echo", func() bool { return strings.Contains(out.String(), "echo:ls\n") })

	// StartExec 会尽力 resize 一次（初始尺寸）
	select {
	case q := <-resizeCh:
		if !strings.Contains(q, "w=90") || !strings.Contains(q, "h=28") {
			t.Fatalf("初始 resize query=%q", q)
		}
	case <-time.After(time.Second):
		t.Fatal("缺少初始 resize")
	}
	if err := svc.ResizeDockerExec(id, 130, 50); err != nil {
		t.Fatalf("resize: %v", err)
	}
	select {
	case q := <-resizeCh:
		if !strings.Contains(q, "w=130") || !strings.Contains(q, "h=50") {
			t.Fatalf("resize query=%q", q)
		}
	case <-time.After(time.Second):
		t.Fatal("缺少 resize 请求")
	}

	if err := svc.CloseDockerExec(id); err != nil {
		t.Fatal(err)
	}
	if err := svc.SendDockerExecData(id, []byte("x")); err == nil {
		t.Fatal("关闭后发送应失败")
	}
}

func TestDockerExecAPICloseDockerThenSendFails(t *testing.T) {
	srv := fakeExecAPIServer(t, make(chan string, 8), make(chan string, 1))
	defer srv.Close()
	sess := &fakeDockerSSH{dialAddr: srv.Listener.Addr().String(), output: cliJSONLine}
	svc := newTestDockerService(sess, nil)
	if err := svc.ConnectDocker("s1", dockerSSHConn(nil)); err != nil {
		t.Fatal(err)
	}
	var out execCollector
	id, err := svc.StartDockerExec("s1", "cid1", 80, 24, out.emit)
	if err != nil {
		t.Fatal(err)
	}
	if err := svc.SendDockerExecData(id, []byte("a")); err != nil {
		t.Fatal(err)
	}
	if err := svc.CloseDocker("s1"); err != nil {
		t.Fatal(err)
	}
	if err := svc.SendDockerExecData(id, []byte("b")); err == nil {
		t.Fatal("CloseDocker 后 Send 应失败")
	}
}
