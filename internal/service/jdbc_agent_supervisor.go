package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type JDBCRuntimeSelector interface {
	SelectRuntime() (*RuntimeSelection, error)
}

type JDBCAgentStarter interface {
	StartAgent(ctx context.Context, config AgentProcessConfig) (*AgentProcessHandle, error)
	Stop() error
}

type JDBCAgentDialer interface {
	Dial(ctx context.Context, host string, port int) (JdbcAgentClient, func() error, error)
}

type JDBCAgentConnection struct {
	Client JdbcAgentClient
	Token  string
}

type JDBCAgentSupervisor struct {
	runtime  JDBCRuntimeSelector
	starter  JDBCAgentStarter
	dialer   JDBCAgentDialer
	agentJar string

	mu          sync.Mutex
	connection  *JDBCAgentConnection
	closeClient func() error
	statusMu    sync.RWMutex
	status      JDBCAgentStatus
}

func NewJDBCAgentSupervisor(runtime JDBCRuntimeSelector, starter JDBCAgentStarter, dialer JDBCAgentDialer, agentJar string) *JDBCAgentSupervisor {
	if dialer == nil {
		dialer = grpcJDBCAgentDialer{}
	}
	return &JDBCAgentSupervisor{
		runtime:  runtime,
		starter:  starter,
		dialer:   dialer,
		agentJar: agentJar,
		status:   JDBCAgentStatus{State: JDBCAgentStateStopped, RuntimeKind: RuntimeKindMissing},
	}
}

func (s *JDBCAgentSupervisor) Status() JDBCAgentStatus {
	s.statusMu.RLock()
	defer s.statusMu.RUnlock()
	return s.status
}

func (s *JDBCAgentSupervisor) Client(ctx context.Context) (*JDBCAgentConnection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.clientLocked(ctx)
}

func (s *JDBCAgentSupervisor) Restart(ctx context.Context) (*JDBCAgentConnection, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.closeLocked(); err != nil {
		return nil, fmt.Errorf("重启前关闭 JDBC agent 失败: %w", err)
	}
	return s.clientLocked(ctx)
}

func (s *JDBCAgentSupervisor) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closeLocked()
}

func (s *JDBCAgentSupervisor) clientLocked(ctx context.Context) (*JDBCAgentConnection, error) {
	if s.connection != nil {
		return s.connection, nil
	}
	s.setStatus(JDBCAgentStateStarting, "", "")
	if s.runtime == nil {
		err := &JDBCError{Code: JDBCErrorRuntimeMissing, Message: "JDBC runtime selector 未配置"}
		s.setStatus(JDBCAgentStateFailed, RuntimeKindMissing, err.Error())
		return nil, err
	}
	if s.starter == nil {
		err := &JDBCError{Code: JDBCErrorAgentUnavailable, Message: "JDBC agent starter 未配置"}
		s.setStatus(JDBCAgentStateFailed, "", err.Error())
		return nil, err
	}

	selected, err := s.runtime.SelectRuntime()
	if err != nil {
		mapped := newJDBCError(err.Error(), err)
		s.setStatus(JDBCAgentStateFailed, RuntimeKindMissing, mapped.Error())
		return nil, mapped
	}
	if selected == nil || selected.Kind == RuntimeKindMissing || selected.JavaPath == "" {
		err := &JDBCError{Code: JDBCErrorRuntimeMissing, Message: "未找到可用 Java 运行时"}
		s.setStatus(JDBCAgentStateFailed, RuntimeKindMissing, err.Error())
		return nil, err
	}
	s.setStatus(JDBCAgentStateStarting, selected.Kind, "")
	if s.agentJar == "" {
		err := &JDBCError{Code: JDBCErrorAgentUnavailable, Message: "JDBC agent jar 未配置"}
		s.setStatus(JDBCAgentStateFailed, selected.Kind, err.Error())
		return nil, err
	}

	handle, err := s.starter.StartAgent(context.Background(), AgentProcessConfig{
		JavaPath: selected.JavaPath,
		AgentJar: s.agentJar,
	})
	if err != nil {
		mapped := &JDBCError{Code: JDBCErrorAgentUnavailable, Message: err.Error(), Err: err}
		s.setStatus(JDBCAgentStateFailed, selected.Kind, mapped.Error())
		return nil, mapped
	}
	client, closeClient, err := s.dialer.Dial(ctx, "127.0.0.1", handle.Port)
	if err != nil {
		_ = s.starter.Stop()
		mapped := &JDBCError{Code: JDBCErrorAgentUnavailable, Message: err.Error(), Err: err}
		s.setStatus(JDBCAgentStateFailed, selected.Kind, mapped.Error())
		return nil, mapped
	}

	s.connection = &JDBCAgentConnection{Client: client, Token: handle.Token}
	s.closeClient = closeClient
	s.setStatus(JDBCAgentStateRunning, selected.Kind, "")
	return s.connection, nil
}

func (s *JDBCAgentSupervisor) closeLocked() error {
	var closeErr error
	if s.closeClient != nil {
		closeErr = s.closeClient()
	}
	stopErr := error(nil)
	if s.starter != nil {
		stopErr = s.starter.Stop()
	}
	s.connection = nil
	s.closeClient = nil
	combined := errors.Join(closeErr, stopErr)
	if combined != nil {
		s.setStatus(JDBCAgentStateFailed, "", combined.Error())
	} else {
		s.setStatus(JDBCAgentStateStopped, "", "")
	}
	return combined
}

func (s *JDBCAgentSupervisor) setStatus(state JDBCAgentState, runtimeKind RuntimeKind, lastError string) {
	s.statusMu.Lock()
	defer s.statusMu.Unlock()
	if runtimeKind == "" {
		runtimeKind = s.status.RuntimeKind
	}
	s.status = JDBCAgentStatus{State: state, RuntimeKind: runtimeKind, LastError: lastError}
}

type grpcJDBCAgentDialer struct{}

func (grpcJDBCAgentDialer) Dial(ctx context.Context, host string, port int) (JdbcAgentClient, func() error, error) {
	return NewGRPCJdbcAgentClient(ctx, host, port)
}
