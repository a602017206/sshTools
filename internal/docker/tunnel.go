package docker

import (
	"io"
	"net"
	"sync"
)

// UnixDialer dials a network address on the remote side (e.g. via SSH).
type UnixDialer interface {
	Dial(network, addr string) (net.Conn, error)
}

// UnixTunnel forwards local TCP connections to a remote Unix socket via UnixDialer.
type UnixTunnel struct {
	listener     net.Listener
	dialer       UnixDialer
	remoteSocket string

	mu     sync.Mutex
	closed bool
	conns  map[net.Conn]struct{}
}

// StartUnixTunnel listens on 127.0.0.1:0 and forwards each accepted connection
// to dialer.Dial("unix", remoteSocket).
func StartUnixTunnel(dialer UnixDialer, remoteSocket string) (localAddr string, tun *UnixTunnel, err error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", nil, err
	}
	tun = &UnixTunnel{
		listener:     ln,
		dialer:       dialer,
		remoteSocket: remoteSocket,
		conns:        make(map[net.Conn]struct{}),
	}
	go tun.serve()
	return ln.Addr().String(), tun, nil
}

func (t *UnixTunnel) serve() {
	for {
		tcpConn, err := t.listener.Accept()
		if err != nil {
			return
		}
		t.track(tcpConn)
		go t.handle(tcpConn)
	}
}

func (t *UnixTunnel) handle(tcpConn net.Conn) {
	defer t.untrack(tcpConn)
	defer tcpConn.Close()

	remote, err := t.dialer.Dial("unix", t.remoteSocket)
	if err != nil {
		return
	}
	defer remote.Close()

	pipeBidirectional(tcpConn, remote)
}

// pipeBidirectional 双向转发 a 与 b；任一方向结束即关闭两端，
// 否则另一方向的 Copy 会一直阻塞（如 exec 结束后远端已关、本地仍在等读），泄漏连接与 goroutine。
func pipeBidirectional(a, b io.ReadWriteCloser) {
	var once sync.Once
	closeBoth := func() {
		once.Do(func() {
			_ = a.Close()
			_ = b.Close()
		})
	}
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		defer closeBoth()
		_, _ = io.Copy(b, a)
	}()
	go func() {
		defer wg.Done()
		defer closeBoth()
		_, _ = io.Copy(a, b)
	}()
	wg.Wait()
}

func (t *UnixTunnel) track(c net.Conn) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.closed {
		return
	}
	t.conns[c] = struct{}{}
}

func (t *UnixTunnel) untrack(c net.Conn) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.conns, c)
}

// Close shuts down the listener and all active forwarded connections.
func (t *UnixTunnel) Close() error {
	t.mu.Lock()
	if t.closed {
		t.mu.Unlock()
		return nil
	}
	t.closed = true
	active := make([]net.Conn, 0, len(t.conns))
	for c := range t.conns {
		active = append(active, c)
	}
	t.mu.Unlock()

	err := t.listener.Close()
	for _, c := range active {
		_ = c.Close()
	}
	return err
}
