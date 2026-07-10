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
	}
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
	if s.runtime == nil {
		return nil, &JDBCError{Code: JDBCErrorRuntimeMissing, Message: "JDBC runtime selector 未配置"}
	}
	if s.starter == nil {
		return nil, &JDBCError{Code: JDBCErrorAgentUnavailable, Message: "JDBC agent starter 未配置"}
	}

	selected, err := s.runtime.SelectRuntime()
	if err != nil {
		return nil, newJDBCError(err.Error(), err)
	}
	if selected == nil || selected.Kind == RuntimeKindMissing || selected.JavaPath == "" {
		return nil, &JDBCError{Code: JDBCErrorRuntimeMissing, Message: "未找到可用 Java 运行时"}
	}
	if s.agentJar == "" {
		return nil, &JDBCError{Code: JDBCErrorAgentUnavailable, Message: "JDBC agent jar 未配置"}
	}

	handle, err := s.starter.StartAgent(context.Background(), AgentProcessConfig{
		JavaPath: selected.JavaPath,
		AgentJar: s.agentJar,
	})
	if err != nil {
		return nil, &JDBCError{Code: JDBCErrorAgentUnavailable, Message: err.Error(), Err: err}
	}
	client, closeClient, err := s.dialer.Dial(ctx, "127.0.0.1", handle.Port)
	if err != nil {
		_ = s.starter.Stop()
		return nil, &JDBCError{Code: JDBCErrorAgentUnavailable, Message: err.Error(), Err: err}
	}

	s.connection = &JDBCAgentConnection{Client: client, Token: handle.Token}
	s.closeClient = closeClient
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
	return errors.Join(closeErr, stopErr)
}

type grpcJDBCAgentDialer struct{}

func (grpcJDBCAgentDialer) Dial(ctx context.Context, host string, port int) (JdbcAgentClient, func() error, error) {
	return NewGRPCJdbcAgentClient(ctx, host, port)
}
