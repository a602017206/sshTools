//go:build windows

package service

import (
	"os/exec"
)

func configureAgentCmd(cmd *exec.Cmd) {
	// Windows does not need process-group configuration for the agent command.
}

func stopAgentCmd(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	return cmd.Process.Kill()
}

func forceKillAgentCmd(cmd *exec.Cmd) error {
	return stopAgentCmd(cmd)
}
