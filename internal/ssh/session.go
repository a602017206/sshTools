package ssh

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"golang.org/x/crypto/ssh"
)

// Session represents an SSH session
type Session struct {
	client  *Client
	session *ssh.Session
	stdin   io.WriteCloser
	stdout  io.Reader
	stderr  io.Reader
}

// NewSession creates a new SSH session
func (c *Client) NewSession() (*Session, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("client not connected")
	}

	session, err := c.client.NewSession()
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		session.Close()
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	return &Session{
		client:  c,
		session: session,
		stdin:   stdin,
		stdout:  stdout,
		stderr:  stderr,
	}, nil
}

// RequestPTY requests a pseudo-terminal
func (s *Session) RequestPTY(term string, height, width int) error {
	modes := ssh.TerminalModes{
		ssh.ECHO:          1, // 启用本地回显，让用户看到输入
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := s.session.RequestPty(term, height, width, modes); err != nil {
		return fmt.Errorf("failed to request PTY: %w", err)
	}

	return nil
}

// Shell starts a login shell
func (s *Session) Shell() error {
	if err := s.session.Shell(); err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}
	return nil
}

// Write writes data to session stdin
func (s *Session) Write(data []byte) (int, error) {
	return s.stdin.Write(data)
}

// Read reads data from session stdout
func (s *Session) Read(buf []byte) (int, error) {
	return s.stdout.Read(buf)
}

// Close closes the session
func (s *Session) Close() error {
	if s.session != nil {
		return s.session.Close()
	}
	return nil
}

// Wait waits for the session to finish
func (s *Session) Wait() error {
	return s.session.Wait()
}

// Resize changes the terminal size
func (s *Session) Resize(height, width int) error {
	return s.session.WindowChange(height, width)
}

// ExecuteCommand runs a single command and returns output
// This creates a separate session for non-PTY command execution
func (s *Session) ExecuteCommand(cmd string, timeout time.Duration) (stdout string, stderr string, err error) {
	// Create a new session from the client
	newSession, err := s.client.client.NewSession()
	if err != nil {
		return "", "", fmt.Errorf("failed to create command session: %w", err)
	}
	defer newSession.Close()

	// Set timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Capture stdout and stderr
	var stdoutBuf, stderrBuf bytes.Buffer
	newSession.Stdout = &stdoutBuf
	newSession.Stderr = &stderrBuf

	// Run command with timeout
	done := make(chan error, 1)
	go func() {
		done <- newSession.Run(cmd)
	}()

	select {
	case err := <-done:
		return stdoutBuf.String(), stderrBuf.String(), err
	case <-ctx.Done():
		newSession.Signal(ssh.SIGKILL)
		return stdoutBuf.String(), stderrBuf.String(), fmt.Errorf("command timeout")
	}
}
