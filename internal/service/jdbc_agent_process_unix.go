//go:build !windows

package service

import (
	"os/exec"
	"syscall"
)

func configureAgentCmd(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func stopAgentCmd(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	pid := cmd.Process.Pid
	// 负 PID：向整个进程组发送信号，避免 java 子进程残留。
	if err := syscall.Kill(-pid, syscall.SIGTERM); err != nil {
		if err == syscall.ESRCH {
			return nil
		}
		// 回退到单进程终止
		if killErr := cmd.Process.Signal(syscall.SIGTERM); killErr != nil && killErr != syscall.ESRCH {
			_ = cmd.Process.Kill()
			return killErr
		}
	}
	return nil
}

func forceKillAgentCmd(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	pid := cmd.Process.Pid
	if err := syscall.Kill(-pid, syscall.SIGKILL); err != nil {
		if err == syscall.ESRCH {
			return nil
		}
		killErr := cmd.Process.Kill()
		if killErr == nil || killErr == syscall.ESRCH {
			return nil
		}
		return killErr
	}
	return nil
}
