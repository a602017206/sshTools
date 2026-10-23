package docker

import (
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
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
