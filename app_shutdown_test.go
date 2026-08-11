package main

import (
	"context"
	"testing"

	"AHaSSHTools/internal/config"
	"AHaSSHTools/internal/service"
	"AHaSSHTools/internal/service/jdbcproto"
	"AHaSSHTools/internal/ssh"
)

func TestAppShutdownClosesSessionsAndJDBCAgent(t *testing.T) {
	starter := &shutdownAgentStarter{}
	supervisor := service.NewJDBCAgentSupervisor(
		&shutdownRuntimeSelector{selection: &service.RuntimeSelection{
			Kind:     service.RuntimeKindManaged,
			JavaPath: "/managed/bin/java",
			Version:  "21",
		}},
		starter,
		&shutdownAgentDialer{},
		"/agent/jdbc-agent.jar",
	)
	if _, err := supervisor.Client(context.Background()); err != nil {
		t.Fatalf("start agent: %v", err)
	}

	sessionManager := ssh.NewSessionManager()
	sessionService := service.NewSessionService(sessionManager)
	if _, err := sessionManager.CreateLocalSession("local-1", "zsh"); err != nil {
		t.Logf("skip real local shell setup: %v", err)
	}

	nativeClient := &shutdownNativeClient{}
	nativeService := service.NewNativeDatabaseService(map[service.NativeDatabaseType]service.NativeDatabaseProvider{
		service.NativeDatabaseTypeRedis: &shutdownNativeProvider{client: nativeClient},
	})
	if err := nativeService.Connect(context.Background(), "redis-1", service.NativeDatabaseConfig{
		Type: service.NativeDatabaseTypeRedis,
		Host: "127.0.0.1",
		Port: 6379,
	}); err != nil {
		t.Fatalf("connect native: %v", err)
	}

	gateway := &shutdownDatabaseGateway{}
	databaseService := service.NewDatabaseServiceWithGateway(nil, gateway)
	if err := databaseService.ConnectDatabase("db-1", "127.0.0.1", 3306, "u", "p", "mysql", "demo"); err != nil {
		t.Fatalf("connect database: %v", err)
	}

	app := &App{
		sessionService:        sessionService,
		databaseService:       databaseService,
		nativeDatabaseService: nativeService,
		jdbcAgentSupervisor:   supervisor,
	}
	app.shutdown(context.Background())

	if got := sessionManager.ListSessions(); len(got) != 0 {
		t.Fatalf("expected all shell sessions closed, got %v", got)
	}
	if got := databaseService.ListSessions(); len(got) != 0 {
		t.Fatalf("expected all database sessions closed, got %v", got)
	}
	if got := nativeService.ListSessions(); len(got) != 0 {
		t.Fatalf("expected all native sessions closed, got %v", got)
	}
	if !nativeClient.closed {
		t.Fatal("expected native client closed")
	}
	if gateway.closeCalls == 0 {
		t.Fatal("expected database gateway sessions closed")
	}
	if starter.stopCalls == 0 {
		t.Fatal("expected JDBC agent process stopped")
	}
	if status := supervisor.Status(); status.State != service.JDBCAgentStateStopped {
		t.Fatalf("expected agent stopped, got %+v", status)
	}
}

type shutdownRuntimeSelector struct {
	selection *service.RuntimeSelection
}

func (s *shutdownRuntimeSelector) SelectRuntime() (*service.RuntimeSelection, error) {
	return s.selection, nil
}

type shutdownAgentStarter struct {
	stopCalls int
}

func (s *shutdownAgentStarter) StartAgent(context.Context, service.AgentProcessConfig) (*service.AgentProcessHandle, error) {
	return &service.AgentProcessHandle{
		Port:    47001,
		Token:   "token",
		Process: &shutdownAgentProcess{alive: true},
	}, nil
}

func (s *shutdownAgentStarter) Stop() error {
	s.stopCalls++
	return nil
}

type shutdownAgentProcess struct{ alive bool }

func (p *shutdownAgentProcess) Stop() error { p.alive = false; return nil }
func (p *shutdownAgentProcess) Alive() bool { return p.alive }

type shutdownAgentDialer struct{}

func (shutdownAgentDialer) Dial(context.Context, string, int) (service.JdbcAgentClient, func() error, error) {
	return stubShutdownJdbcClient{}, func() error { return nil }, nil
}

type stubShutdownJdbcClient struct{}

func (stubShutdownJdbcClient) OpenSession(context.Context, *jdbcproto.OpenSessionRequest) (*jdbcproto.OpenSessionResponse, error) {
	return &jdbcproto.OpenSessionResponse{}, nil
}
func (stubShutdownJdbcClient) ExecuteQuery(context.Context, *jdbcproto.ExecuteQueryRequest) (*jdbcproto.QueryResult, error) {
	return &jdbcproto.QueryResult{}, nil
}
func (stubShutdownJdbcClient) ListSchemas(context.Context, *jdbcproto.ListSchemasRequest) (*jdbcproto.ListSchemasResponse, error) {
	return &jdbcproto.ListSchemasResponse{}, nil
}
func (stubShutdownJdbcClient) ListRoutines(context.Context, *jdbcproto.ListRoutinesRequest) (*jdbcproto.ListRoutinesResponse, error) {
	return &jdbcproto.ListRoutinesResponse{}, nil
}
func (stubShutdownJdbcClient) ListTables(context.Context, *jdbcproto.ListTablesRequest) (*jdbcproto.ListTablesResponse, error) {
	return &jdbcproto.ListTablesResponse{}, nil
}
func (stubShutdownJdbcClient) ListColumns(context.Context, *jdbcproto.ListColumnsRequest) (*jdbcproto.ListColumnsResponse, error) {
	return &jdbcproto.ListColumnsResponse{}, nil
}
func (stubShutdownJdbcClient) CloseSession(context.Context, *jdbcproto.CloseSessionRequest) (*jdbcproto.CloseSessionResponse, error) {
	return &jdbcproto.CloseSessionResponse{Closed: true}, nil
}

type shutdownNativeProvider struct {
	client *shutdownNativeClient
}

func (p *shutdownNativeProvider) Test(context.Context, service.NativeDatabaseConfig) error {
	return nil
}
func (p *shutdownNativeProvider) Connect(context.Context, service.NativeDatabaseConfig) (service.NativeDatabaseClient, error) {
	return p.client, nil
}

type shutdownNativeClient struct{ closed bool }

func (*shutdownNativeClient) ListPrimaryResources(context.Context) ([]service.NativeResource, error) {
	return nil, nil
}
func (*shutdownNativeClient) ListSecondaryResources(context.Context, string) ([]service.NativeResource, error) {
	return nil, nil
}
func (c *shutdownNativeClient) Close() error {
	c.closed = true
	return nil
}

type shutdownDatabaseGateway struct {
	closeCalls int
}

func (g *shutdownDatabaseGateway) ConnectDatabase(context.Context, string, config.DatabaseConfig) error {
	return nil
}
func (*shutdownDatabaseGateway) ExecuteQuery(context.Context, string, string) (*service.QueryResult, error) {
	return &service.QueryResult{}, nil
}
func (*shutdownDatabaseGateway) ListTables(context.Context, string, string) ([]string, error) {
	return nil, nil
}
func (*shutdownDatabaseGateway) ListDatabases(context.Context, string) ([]string, error) {
	return nil, nil
}
func (*shutdownDatabaseGateway) GetTableSchema(context.Context, string, string) (*config.TableSchema, error) {
	return &config.TableSchema{}, nil
}
func (*shutdownDatabaseGateway) GetTableSchemaInSchema(context.Context, string, string, string) (*config.TableSchema, error) {
	return &config.TableSchema{}, nil
}
func (g *shutdownDatabaseGateway) CloseDatabase(context.Context, string) error {
	g.closeCalls++
	return nil
}
