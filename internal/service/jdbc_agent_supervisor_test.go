package service

import (
	"context"
	"fmt"
	"testing"
)

func TestJDBCAgentSupervisorStartsOnceAndReturnsAuthenticatedClient(t *testing.T) {
	runtimeSelector := &supervisorRuntimeSelector{
		selection: &RuntimeSelection{Kind: RuntimeKindManaged, JavaPath: "/managed/bin/java", Version: "21"},
	}
	starter := &supervisorAgentStarter{}
	dialer := &supervisorAgentDialer{}
	supervisor := NewJDBCAgentSupervisor(runtimeSelector, starter, dialer, "/agent/jdbc-agent.jar", "/logs/jdbc-agent.log")

	first, err := supervisor.Client(context.Background())
	if err != nil {
		t.Fatalf("first client failed: %v", err)
	}
	second, err := supervisor.Client(context.Background())
	if err != nil {
		t.Fatalf("second client failed: %v", err)
	}

	if runtimeSelector.calls != 1 || starter.startCalls != 1 || dialer.calls != 1 {
		t.Fatalf("expected one lazy start, runtime=%d start=%d dial=%d", runtimeSelector.calls, starter.startCalls, dialer.calls)
	}
	if first.Client != second.Client || first.Token != second.Token {
		t.Fatalf("expected cached connection")
	}
	if first.Token != "token-1" {
		t.Fatalf("expected process token, got %q", first.Token)
	}
	if dialer.host != "127.0.0.1" || dialer.port != 47001 {
		t.Fatalf("unexpected dial target: %s:%d", dialer.host, dialer.port)
	}
	if starter.lastConfig.JavaPath != "/managed/bin/java" || starter.lastConfig.AgentJar != "/agent/jdbc-agent.jar" {
		t.Fatalf("unexpected process config: %+v", starter.lastConfig)
	}
	if starter.lastConfig.LogPath != "/logs/jdbc-agent.log" {
		t.Fatalf("unexpected agent log path: %q", starter.lastConfig.LogPath)
	}
}

func TestJDBCAgentSupervisorRestartClosesClientAndRotatesProcess(t *testing.T) {
	runtimeSelector := &supervisorRuntimeSelector{
		selection: &RuntimeSelection{Kind: RuntimeKindSystem, JavaPath: "/system/bin/java", Version: "21"},
	}
	starter := &supervisorAgentStarter{}
	dialer := &supervisorAgentDialer{}
	supervisor := NewJDBCAgentSupervisor(runtimeSelector, starter, dialer, "/agent/jdbc-agent.jar")

	before, err := supervisor.Client(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	after, err := supervisor.Restart(context.Background())
	if err != nil {
		t.Fatalf("restart failed: %v", err)
	}

	if dialer.closeCalls != 1 {
		t.Fatalf("expected old gRPC client to close, got %d", dialer.closeCalls)
	}
	if starter.stopCalls != 1 {
		t.Fatalf("expected old process to stop, got %d", starter.stopCalls)
	}
	if starter.startCalls != 2 || dialer.calls != 2 {
		t.Fatalf("expected a new process and client, start=%d dial=%d", starter.startCalls, dialer.calls)
	}
	if before.Token == after.Token || after.Token != "token-2" {
		t.Fatalf("expected rotated token, before=%q after=%q", before.Token, after.Token)
	}
	if before.Client == after.Client {
		t.Fatalf("expected a new gRPC client")
	}
}

func TestJDBCAgentSupervisorRefreshStatusDetectsExitedProcess(t *testing.T) {
	runtimeSelector := &supervisorRuntimeSelector{
		selection: &RuntimeSelection{Kind: RuntimeKindManaged, JavaPath: "/managed/bin/java", Version: "21"},
	}
	starter := &supervisorAgentStarter{}
	supervisor := NewJDBCAgentSupervisor(runtimeSelector, starter, &supervisorAgentDialer{}, "/agent/jdbc-agent.jar")
	if _, err := supervisor.Client(context.Background()); err != nil {
		t.Fatal(err)
	}

	starter.healthErr = fmt.Errorf("JDBC agent 进程已退出")
	status := supervisor.RefreshStatus(context.Background())
	if status.State != JDBCAgentStateFailed || status.LastError == "" {
		t.Fatalf("expected failed status after process exit, got %+v", status)
	}
	if supervisor.connection != nil {
		t.Fatal("dead agent connection was not cleared")
	}
}

type supervisorRuntimeSelector struct {
	selection *RuntimeSelection
	err       error
	calls     int
}

func (s *supervisorRuntimeSelector) SelectRuntime() (*RuntimeSelection, error) {
	s.calls++
	return s.selection, s.err
}

type supervisorAgentStarter struct {
	startCalls int
	stopCalls  int
	lastConfig AgentProcessConfig
	healthErr  error
}

func (s *supervisorAgentStarter) StartAgent(_ context.Context, config AgentProcessConfig) (*AgentProcessHandle, error) {
	s.startCalls++
	s.lastConfig = config
	return &AgentProcessHandle{
		Port:    47000 + s.startCalls,
		Token:   fmt.Sprintf("token-%d", s.startCalls),
		Process: &fakeAgentProcess{alive: true},
	}, nil
}

func (s *supervisorAgentStarter) Stop() error {
	s.stopCalls++
	return nil
}

func (s *supervisorAgentStarter) Health(context.Context) error {
	return s.healthErr
}

type supervisorAgentDialer struct {
	calls      int
	closeCalls int
	host       string
	port       int
}

func (d *supervisorAgentDialer) Dial(_ context.Context, host string, port int) (JdbcAgentClient, func() error, error) {
	d.calls++
	d.host = host
	d.port = port
	client := &fakeJdbcAgentClient{}
	return client, func() error {
		d.closeCalls++
		return nil
	}, nil
}
