package ssh

import (
	"fmt"
	"io"
	"sync"

	cryptossh "golang.org/x/crypto/ssh"
)

// InteractiveStream 是带 PTY 的远程命令双向流：Read 读取输出（stdout/stderr 经 PTY 合并），
// Write 写入 stdin，Resize 调整窗口大小，Close 终止远程命令。
type InteractiveStream interface {
	io.ReadWriteCloser
	Resize(cols, rows int) error
}

type ptyStream struct {
	session   *cryptossh.Session
	stdin     io.WriteCloser
	stdout    io.Reader
	closeOnce sync.Once
	closeErr  error
}

func (p *ptyStream) Read(b []byte) (int, error)  { return p.stdout.Read(b) }
func (p *ptyStream) Write(b []byte) (int, error) { return p.stdin.Write(b) }

func (p *ptyStream) Resize(cols, rows int) error {
	if cols <= 0 || rows <= 0 {
		return fmt.Errorf("invalid terminal size %dx%d", cols, rows)
	}
	return p.session.WindowChange(rows, cols)
}

func (p *ptyStream) Close() error {
	p.closeOnce.Do(func() {
		_ = p.stdin.Close()
		p.closeErr = p.session.Close()
		if p.closeErr == io.EOF {
			p.closeErr = nil
		}
	})
	return p.closeErr
}

// StartPTYCommand 在新会话中申请 xterm-256color PTY 并启动 cmd，返回可交互流。
// 关闭流会结束会话并终止远程命令。
func (c *Client) StartPTYCommand(cmd string, cols, rows int) (InteractiveStream, error) {
	if c.client == nil {
		return nil, fmt.Errorf("ssh client not connected")
	}
	if cols <= 0 {
		cols = 80
	}
	if rows <= 0 {
		rows = 24
	}
	session, err := c.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create ssh session: %w", err)
	}
	modes := cryptossh.TerminalModes{
		cryptossh.ECHO:          1,
		cryptossh.TTY_OP_ISPEED: 14400,
		cryptossh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm-256color", rows, cols, modes); err != nil {
		_ = session.Close()
		return nil, fmt.Errorf("failed to request pty: %w", err)
	}
	stdin, err := session.StdinPipe()
	if err != nil {
		_ = session.Close()
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}
	stdout, err := session.StdoutPipe()
	if err != nil {
		_ = session.Close()
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}
	if err := session.Start(cmd); err != nil {
		_ = session.Close()
		return nil, fmt.Errorf("failed to start command: %w", err)
	}
	return &ptyStream{session: session, stdin: stdin, stdout: stdout}, nil
}
