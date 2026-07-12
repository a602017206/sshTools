package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type AgentProcessConfig struct {
	JavaPath string
	AgentJar string
	LogPath  string
}

type AgentProcessHandle struct {
	Port    int
	Token   string
	Process AgentProcess
}

type AgentProcess interface {
	Stop() error
	Alive() bool
}

type AgentCommandRunner interface {
	Start(ctx context.Context, name string, args ...string) (AgentProcess, error)
}

type AgentCommandOutputRunner interface {
	StartWithOutput(ctx context.Context, name, logPath string, args ...string) (AgentProcess, error)
}

type AgentProcessManager struct {
	runner AgentCommandRunner
	config AgentProcessConfig
	handle *AgentProcessHandle
	mu     sync.Mutex
}

func NewAgentProcessManager(runner AgentCommandRunner, config AgentProcessConfig) *AgentProcessManager {
	if runner == nil {
		runner = execAgentCommandRunner{}
	}
	return &AgentProcessManager{runner: runner, config: config}
}

func (m *AgentProcessManager) Start(ctx context.Context) (*AgentProcessHandle, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.startLocked(ctx, m.config)
}

func (m *AgentProcessManager) StartAgent(ctx context.Context, config AgentProcessConfig) (*AgentProcessHandle, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = config
	return m.startLocked(ctx, config)
}

func (m *AgentProcessManager) startLocked(ctx context.Context, config AgentProcessConfig) (*AgentProcessHandle, error) {
	if m.handle != nil {
		return m.handle, nil
	}
	port, err := chooseLocalPort()
	if err != nil {
		return nil, err
	}
	token, err := generateAgentToken()
	if err != nil {
		return nil, err
	}
	args := []string{"-jar", config.AgentJar, "--port", strconv.Itoa(port), "--token", token}
	var process AgentProcess
	if outputRunner, ok := m.runner.(AgentCommandOutputRunner); ok && config.LogPath != "" {
		process, err = outputRunner.StartWithOutput(ctx, config.JavaPath, config.LogPath, args...)
	} else {
		process, err = m.runner.Start(ctx, config.JavaPath, args...)
	}
	if err != nil {
		return nil, fmt.Errorf("启动 JDBC agent 失败: %w", err)
	}
	handle := &AgentProcessHandle{Port: port, Token: token, Process: process}
	m.handle = handle
	return handle, nil
}

func (m *AgentProcessManager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.handle == nil || m.handle.Process == nil {
		return nil
	}
	err := m.handle.Process.Stop()
	m.handle = nil
	return err
}

func (m *AgentProcessManager) Health(context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.handle == nil || m.handle.Process == nil {
		return fmt.Errorf("JDBC agent 未启动")
	}
	if !m.handle.Process.Alive() {
		m.handle = nil
		return fmt.Errorf("JDBC agent 进程已退出")
	}
	return nil
}

type execAgentCommandRunner struct{}

func (execAgentCommandRunner) Start(ctx context.Context, name string, args ...string) (AgentProcess, error) {
	return startExecAgentProcess(ctx, name, nil, args...)
}

func (execAgentCommandRunner) StartWithOutput(ctx context.Context, name, logPath string, args ...string) (AgentProcess, error) {
	if err := os.MkdirAll(filepath.Dir(logPath), 0o700); err != nil {
		return nil, fmt.Errorf("创建 JDBC agent 日志目录失败: %w", err)
	}
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("打开 JDBC agent 日志失败: %w", err)
	}
	return startExecAgentProcess(ctx, name, logFile, args...)
}

func startExecAgentProcess(ctx context.Context, name string, logFile *os.File, args ...string) (AgentProcess, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if logFile != nil {
		if _, err := fmt.Fprintf(logFile, "%s JDBC agent 启动: executable=%s\n", time.Now().Format(time.RFC3339), name); err != nil {
			_ = logFile.Close()
			return nil, fmt.Errorf("写入 JDBC agent 启动日志失败: %w", err)
		}
		cmd.Stdout = logFile
		cmd.Stderr = logFile
	}
	if err := cmd.Start(); err != nil {
		if logFile != nil {
			_ = logFile.Close()
		}
		return nil, err
	}
	process := &execAgentProcess{cmd: cmd, done: make(chan struct{}), logFile: logFile}
	go func() {
		waitErr := cmd.Wait()
		if process.logFile != nil {
			_, _ = fmt.Fprintf(process.logFile, "%s JDBC agent 退出: %v\n", time.Now().Format(time.RFC3339), waitErr)
			_ = process.logFile.Close()
		}
		close(process.done)
	}()
	return process, nil
}

type execAgentProcess struct {
	cmd     *exec.Cmd
	done    chan struct{}
	logFile *os.File
}

func (p *execAgentProcess) Stop() error {
	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}
	if !p.Alive() {
		return nil
	}
	err := p.cmd.Process.Kill()
	if errors.Is(err, os.ErrProcessDone) {
		return nil
	}
	return err
}

func (p *execAgentProcess) Alive() bool {
	if p == nil || p.done == nil {
		return false
	}
	select {
	case <-p.done:
		return false
	default:
		return true
	}
}

func chooseLocalPort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("选择 JDBC agent 本地端口失败: %w", err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}

func generateAgentToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("生成 JDBC agent token 失败: %w", err)
	}
	return hex.EncodeToString(buf), nil
}
