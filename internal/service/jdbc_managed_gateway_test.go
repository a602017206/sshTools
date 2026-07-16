package service

import (
	"context"
	"testing"

	"AHaSSHTools/internal/config"
	"AHaSSHTools/internal/service/jdbcproto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestManagedJDBCGatewayReconnectsSessionAfterAgentUnavailable(t *testing.T) {
	failedClient := &managedGatewayClient{queryErr: status.Error(codes.Unavailable, "agent stopped")}
	recoveredClient := &managedGatewayClient{queryResult: &jdbcproto.QueryResult{
		Columns: []string{"VALUE"},
		Rows:    []*jdbcproto.Row{{Values: []string{"recovered"}}},
	}}
	supervisor := &managedGatewaySupervisor{
		current:   &JDBCAgentConnection{Client: failedClient, Token: "token-1"},
		restarted: &JDBCAgentConnection{Client: recoveredClient, Token: "token-2"},
	}
	gateway := NewManagedJDBCGateway(supervisor)
	gateway.SetProfileResolver(func(context.Context, config.DatabaseConfig) (config.JDBCDriverProfile, error) {
		return config.JDBCDriverProfile{
			ID:          "h2-2.2.224",
			DriverClass: "org.h2.Driver",
			URLTemplate: "jdbc:h2:file:{database}",
			InstallPath: "/drivers/h2/2.2.224",
			Jars:        []config.JDBCJar{{Name: "h2.jar"}},
		}, nil
	})
	cfg := config.DatabaseConfig{DBType: "h2", Host: "localhost", Database: "/tmp/recovery"}
	if err := gateway.ConnectDatabase(context.Background(), "recover-session", cfg); err != nil {
		t.Fatalf("connect failed: %v", err)
	}

	result, err := gateway.ExecuteQuery(context.Background(), "recover-session", "select 'recovered'")
	if err != nil {
		t.Fatalf("query recovery failed: %v", err)
	}
	if len(result.Rows) != 1 || result.Rows[0][0] != "recovered" {
		t.Fatalf("unexpected recovered result: %+v", result)
	}
	if supervisor.restartCalls != 1 {
		t.Fatalf("expected one supervisor restart, got %d", supervisor.restartCalls)
	}
	if len(recoveredClient.openRequests) != 1 {
		t.Fatalf("expected session to reopen once, got %d", len(recoveredClient.openRequests))
	}
	reopened := recoveredClient.openRequests[0]
	if reopened.GetSessionId() != "recover-session" || reopened.GetDatabase() != "/tmp/recovery" {
		t.Fatalf("session config was not restored: %+v", reopened)
	}
	if recoveredClient.queryCalls != 1 {
		t.Fatalf("expected original query to retry once, got %d", recoveredClient.queryCalls)
	}
}

func TestManagedJDBCGatewayListsMySQLDatabasesThroughQuery(t *testing.T) {
	client := &managedGatewayClient{queryResult: &jdbcproto.QueryResult{
		Columns: []string{"Database"},
		Rows: []*jdbcproto.Row{
			{Values: []string{"information_schema"}},
			{Values: []string{"app"}},
		},
	}}
	supervisor := &managedGatewaySupervisor{
		current: &JDBCAgentConnection{Client: client, Token: "token"},
	}
	gateway := NewManagedJDBCGateway(supervisor)
	gateway.SetProfileResolver(func(context.Context, config.DatabaseConfig) (config.JDBCDriverProfile, error) {
		return config.JDBCDriverProfile{ID: "mysql", DriverClass: "com.mysql.cj.jdbc.Driver"}, nil
	})
	if err := gateway.ConnectDatabase(context.Background(), "mysql-session", config.DatabaseConfig{DBType: "mysql"}); err != nil {
		t.Fatalf("connect failed: %v", err)
	}

	databases, err := gateway.ListDatabases(context.Background(), "mysql-session")
	if err != nil {
		t.Fatalf("list databases failed: %v", err)
	}
	if len(databases) != 2 || databases[0] != "information_schema" || databases[1] != "app" {
		t.Fatalf("unexpected databases: %v", databases)
	}
	if len(client.queryRequests) != 1 || client.queryRequests[0].GetSql() != "SHOW DATABASES" {
		t.Fatalf("unexpected database query: %+v", client.queryRequests)
	}
}

func TestManagedJDBCGatewayListsKingbaseDatabasesThroughPostgreSQLCatalog(t *testing.T) {
	client := &managedGatewayClient{queryResult: &jdbcproto.QueryResult{
		Columns: []string{"datname"},
		Rows: []*jdbcproto.Row{
			{Values: []string{"kingbase"}},
			{Values: []string{"application"}},
		},
	}}
	supervisor := &managedGatewaySupervisor{
		current: &JDBCAgentConnection{Client: client, Token: "token"},
	}
	gateway := NewManagedJDBCGateway(supervisor)
	gateway.SetProfileResolver(func(context.Context, config.DatabaseConfig) (config.JDBCDriverProfile, error) {
		return config.JDBCDriverProfile{ID: "kingbase", DriverClass: "com.kingbase8.Driver"}, nil
	})
	if err := gateway.ConnectDatabase(context.Background(), "kingbase-session", config.DatabaseConfig{DBType: "kingbase"}); err != nil {
		t.Fatalf("connect failed: %v", err)
	}

	databases, err := gateway.ListDatabases(context.Background(), "kingbase-session")
	if err != nil {
		t.Fatalf("list databases failed: %v", err)
	}
	if len(databases) != 2 || databases[0] != "kingbase" || databases[1] != "application" {
		t.Fatalf("unexpected databases: %v", databases)
	}
	if len(client.queryRequests) != 1 || client.queryRequests[0].GetSql() != "SELECT datname FROM pg_database WHERE datistemplate = false ORDER BY datname" {
		t.Fatalf("unexpected database query: %+v", client.queryRequests)
	}
}

type managedGatewaySupervisor struct {
	current      *JDBCAgentConnection
	restarted    *JDBCAgentConnection
	restartCalls int
}

func (s *managedGatewaySupervisor) Client(context.Context) (*JDBCAgentConnection, error) {
	return s.current, nil
}

func (s *managedGatewaySupervisor) Restart(context.Context) (*JDBCAgentConnection, error) {
	s.restartCalls++
	s.current = s.restarted
	return s.restarted, nil
}

type managedGatewayClient struct {
	openRequests  []*jdbcproto.OpenSessionRequest
	queryRequests []*jdbcproto.ExecuteQueryRequest
	queryErr      error
	queryResult   *jdbcproto.QueryResult
	queryCalls    int
}

func (c *managedGatewayClient) OpenSession(_ context.Context, request *jdbcproto.OpenSessionRequest) (*jdbcproto.OpenSessionResponse, error) {
	c.openRequests = append(c.openRequests, request)
	return &jdbcproto.OpenSessionResponse{SessionId: request.GetSessionId()}, nil
}

func (c *managedGatewayClient) ExecuteQuery(_ context.Context, request *jdbcproto.ExecuteQueryRequest) (*jdbcproto.QueryResult, error) {
	c.queryCalls++
	c.queryRequests = append(c.queryRequests, request)
	if c.queryErr != nil {
		return nil, c.queryErr
	}
	if c.queryResult != nil {
		return c.queryResult, nil
	}
	return &jdbcproto.QueryResult{}, nil
}

func (c *managedGatewayClient) ListTables(context.Context, *jdbcproto.ListTablesRequest) (*jdbcproto.ListTablesResponse, error) {
	return &jdbcproto.ListTablesResponse{}, nil
}

func (c *managedGatewayClient) ListColumns(context.Context, *jdbcproto.ListColumnsRequest) (*jdbcproto.ListColumnsResponse, error) {
	return &jdbcproto.ListColumnsResponse{}, nil
}

func (c *managedGatewayClient) CloseSession(context.Context, *jdbcproto.CloseSessionRequest) (*jdbcproto.CloseSessionResponse, error) {
	return &jdbcproto.CloseSessionResponse{Closed: true}, nil
}
