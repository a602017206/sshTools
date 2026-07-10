package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"sync"
)

type AgentProcessConfig struct {
	JavaPath string
	AgentJar string
}

type AgentProcessHandle struct {
	Port    int
	Token   string
	Process AgentProcess
}

type AgentProcess interface {
	Stop() error
}

type AgentCommandRunner interface {
	Start(ctx context.Context, name string, args ...string) (AgentProcess, error)
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
	process, err := m.runner.Start(
		ctx,
		config.JavaPath,
		"-jar",
		config.AgentJar,
		"--port",
		strconv.Itoa(port),
		"--token",
		token,
	)
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
	if m.handle == nil {
		return fmt.Errorf("JDBC agent 未启动")
	}
	return nil
}

type execAgentCommandRunner struct{}

func (execAgentCommandRunner) Start(ctx context.Context, name string, args ...string) (AgentProcess, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return execAgentProcess{cmd: cmd}, nil
}

type execAgentProcess struct {
	cmd *exec.Cmd
}

func (p execAgentProcess) Stop() error {
	if p.cmd == nil || p.cmd.Process == nil {
		return nil
	}
	return p.cmd.Process.Kill()
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
