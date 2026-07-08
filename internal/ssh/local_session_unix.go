//go:build !windows
// +build !windows

package ssh

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"syscall"

	"github.com/creack/pty"
)

// LocalSession represents a local shell session
type LocalSession struct {
	cmd     *exec.Cmd
	ptyFile *os.File
	stdin   io.Writer
	stdout  io.Reader
}

// NewLocalSession creates a new local shell session
func NewLocalSession(shellType string) (*LocalSession, error) {
	var cmd *exec.Cmd

	// Determine shell command based on platform and requested shell type
	switch runtime.GOOS {
	case "darwin":
		// macOS: default to zsh
		if shellType == "bash" {
			cmd = exec.Command("/bin/bash", "-l")
		} else {
			cmd = exec.Command("/bin/zsh", "-l")
		}
	case "linux":
		// Linux: default to bash
		if shellType == "sh" {
			cmd = exec.Command("/bin/sh", "-l")
		} else {
			cmd = exec.Command("/bin/bash", "-l")
		}
	default:
		return nil, fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	// Start the process with a pseudo-terminal
	ptyFile, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to start PTY: %w", err)
	}

	return &LocalSession{
		cmd:     cmd,
		ptyFile: ptyFile,
		stdin:   ptyFile,
		stdout:  ptyFile,
	}, nil
}

// Write writes data to session stdin
func (ls *LocalSession) Write(data []byte) (int, error) {
	return ls.stdin.Write(data)
}

// Read reads data from session stdout
func (ls *LocalSession) Read(buf []byte) (int, error) {
	return ls.stdout.Read(buf)
}

// Close closes the session
func (ls *LocalSession) Close() error {
	var err error

	// Try to close the PTY file first
	if ls.ptyFile != nil {
		if cerr := ls.ptyFile.Close(); cerr != nil {
			err = cerr
		}
	}

	// Kill the process if it's still running
	if ls.cmd != nil && ls.cmd.Process != nil {
		// Send SIGTERM first for graceful shutdown
		if killErr := ls.cmd.Process.Signal(syscall.SIGTERM); killErr != nil {
			// If SIGTERM fails, force kill
			if killErr2 := ls.cmd.Process.Kill(); killErr2 != nil {
				if err == nil {
					err = killErr2
				}
			}
		}
	}

	return err
}

// Resize changes the terminal size
func (ls *LocalSession) Resize(cols, rows int) error {
	if ls.ptyFile == nil {
		return fmt.Errorf("pty file not available")
	}

	winsize := &pty.Winsize{
		Cols: uint16(cols),
		Rows: uint16(rows),
	}

	if err := pty.Setsize(ls.ptyFile, winsize); err != nil {
		return fmt.Errorf("failed to set PTY size: %w", err)
	}

	return nil
}

// Wait waits for the session to finish
func (ls *LocalSession) Wait() error {
	if ls.cmd == nil {
		return nil
	}
	return ls.cmd.Wait()
}
