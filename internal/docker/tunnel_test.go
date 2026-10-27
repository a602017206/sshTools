package docker

import (
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type unixDialerFunc func(network, addr string) (net.Conn, error)

func (f unixDialerFunc) Dial(network, addr string) (net.Conn, error) {
	return f(network, addr)
}

func TestUnixTunnelForwardsToDialer(t *testing.T) {
	// Short temp dir: t.TempDir() embeds the test name and can exceed macOS sun_path (~103).
	dir, err := os.MkdirTemp("", "ut")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	sock := filepath.Join(dir, "docker.sock")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		c, _ := ln.Accept()
		if c != nil {
			defer c.Close()
			buf := make([]byte, 4)
			_, _ = io.ReadFull(c, buf)
			_, _ = c.Write(buf)
		}
	}()

	dialer := unixDialerFunc(func(network, addr string) (net.Conn, error) {
		return net.Dial(network, addr)
	})
	addr, tun, err := StartUnixTunnel(dialer, sock)
	if err != nil {
		t.Fatal(err)
	}
	defer tun.Close()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_, _ = conn.Write([]byte("ping"))
	out := make([]byte, 4)
	if _, err := io.ReadFull(conn, out); err != nil {
		t.Fatal(err)
	}
	if string(out) != "ping" {
		t.Fatalf("got %q", out)
	}
}

// startUnixTunnelToServer 启动本地 unix socket 服务端（handler 处理首个连接）与指向它的隧道。
func startUnixTunnelToServer(t *testing.T, handler func(net.Conn)) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "ut")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	sock := filepath.Join(dir, "d.sock")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	go func() {
		c, err := ln.Accept()
		if err == nil {
			handler(c)
		}
	}()
	dialer := unixDialerFunc(func(network, addr string) (net.Conn, error) {
		return net.Dial(network, addr)
	})
	addr, tun, err := StartUnixTunnel(dialer, sock)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tun.Close() })
	return addr
}

func TestUnixTunnelClosesClientWhenRemoteEnds(t *testing.T) {
	addr := startUnixTunnelToServer(t, func(c net.Conn) {
		_, _ = c.Write([]byte("bye"))
		_ = c.Close() // 远端结束（如 exec 进程退出）
	})
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	got, err := io.ReadAll(conn)
	if err != nil {
		t.Fatalf("远端关闭后本地应读到 EOF，而不是超时: %v", err)
	}
	if string(got) != "bye" {
		t.Fatalf("got %q", got)
	}
}

func TestUnixTunnelClosesRemoteWhenClientEnds(t *testing.T) {
	remoteDone := make(chan error, 1)
	addr := startUnixTunnelToServer(t, func(c net.Conn) {
		defer c.Close()
		_ = c.SetReadDeadline(time.Now().Add(2 * time.Second))
		_, err := io.ReadAll(c)
		remoteDone <- err
	})
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = conn.Write([]byte("x"))
	_ = conn.Close() // 本地关闭（如 CloseDockerExec）
	select {
	case err := <-remoteDone:
		if err != nil {
			t.Fatalf("本地关闭后远端应读到 EOF，而不是超时: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("远端连接未被关闭")
	}
}

func TestPipeBidirectionalReturnsWhenEitherSideCloses(t *testing.T) {
	for _, name := range []string{"left", "right"} {
		t.Run(name, func(t *testing.T) {
			aOuter, aInner := net.Pipe()
			bInner, bOuter := net.Pipe()
			done := make(chan struct{})
			go func() {
				pipeBidirectional(aInner, bInner)
				close(done)
			}()
			survivor := bOuter
			if name == "left" {
				_ = aOuter.Close()
			} else {
				_ = bOuter.Close()
				survivor = aOuter
			}
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Fatal("一端关闭后 pipeBidirectional 应返回")
			}
			_ = survivor.SetReadDeadline(time.Now().Add(time.Second))
			if _, err := survivor.Read(make([]byte, 1)); err != io.EOF && err != io.ErrClosedPipe {
				t.Fatalf("另一端应被关闭，err=%v", err)
			}
			_ = survivor.Close()
		})
	}
}
