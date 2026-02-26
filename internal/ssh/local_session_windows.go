//go:build windows
// +build windows

package ssh

import (
	"fmt"
	"io"
	"os/exec"
	"sync"
)

// LocalSession represents a local shell session on Windows
// Uses stdin/stdout pipes instead of PTY since creack/pty doesn't support Windows well
type LocalSession struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	mu     sync.Mutex
	closed bool
}

// NewLocalSession creates a new local shell session on Windows
// Only PowerShell is supported as CMD doesn't work properly in pipe mode
func NewLocalSession(shellType string) (*LocalSession, error) {
	// Use PowerShell in interactive mode
	// -NoExit: Keep the shell open after executing commands
	// -NoLogo: Don't display the copyright banner
	// -NoProfile: Don't load the user profile (faster startup)
	cmd := exec.Command("powershell.exe", "-NoExit", "-NoLogo", "-NoProfile")

	// Create pipes for stdin and stdout
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		stdin.Close()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdin.Close()
		stdout.Close()
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Start the process
	if err := cmd.Start(); err != nil {
		stdin.Close()
		stdout.Close()
		stderr.Close()
		return nil, fmt.Errorf("failed to start shell: %w", err)
	}

	session := &LocalSession{
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
	}

	return session, nil
}

// Write writes data to session stdin
func (ls *LocalSession) Write(data []byte) (int, error) {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	if ls.closed {
		return 0, fmt.Errorf("session is closed")
	}

	return ls.stdin.Write(data)
}

// Read reads data from session stdout
func (ls *LocalSession) Read(buf []byte) (int, error) {
	ls.mu.Lock()
	if ls.closed {
		ls.mu.Unlock()
		return 0, io.EOF
	}
	ls.mu.Unlock()

	// Try stdout first (non-blocking)
	n, err := ls.stdout.Read(buf)
	if n > 0 {
		return n, nil
	}
	if err != nil && err != io.EOF {
		return n, err
	}

	// Try stderr (non-blocking)
	n, err = ls.stderr.Read(buf)
	if n > 0 {
		return n, nil
	}

	return 0, err
}

// Close closes the session
func (ls *LocalSession) Close() error {
	ls.mu.Lock()
	defer ls.mu.Unlock()

	if ls.closed {
		return nil
	}
	ls.closed = true

	var err error

	// Close stdin first to signal the process to exit
	if ls.stdin != nil {
		if cerr := ls.stdin.Close(); cerr != nil {
			err = cerr
		}
	}

	// Close stdout and stderr
	if ls.stdout != nil {
		ls.stdout.Close()
	}
	if ls.stderr != nil {
		ls.stderr.Close()
	}

	// Kill the process if it's still running
	if ls.cmd != nil && ls.cmd.Process != nil {
		if killErr := ls.cmd.Process.Kill(); killErr != nil {
			if err == nil {
				err = killErr
			}
		}
	}

	return err
}

// Resize is a no-op on Windows since we don't have a real PTY
func (ls *LocalSession) Resize(cols, rows int) error {
	// On Windows without a real PTY, we can't resize the terminal
	return nil
}

// Wait waits for the session to finish
func (ls *LocalSession) Wait() error {
	if ls.cmd == nil {
		return nil
	}
	return ls.cmd.Wait()
}
