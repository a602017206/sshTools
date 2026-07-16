package service

import (
	"context"
	"errors"
	"testing"

	"AHaSSHTools/internal/config"
	"AHaSSHTools/internal/service/jdbcproto"
)

func TestJdbcGatewayConnectDatabaseOpensAgentSession(t *testing.T) {
	client := &fakeJdbcAgentClient{}
	gateway := NewJdbcGatewayService(client, "secret")
	profile := config.JDBCDriverProfile{
		ID:          "h2-2.2.224",
		DriverClass: "org.h2.Driver",
		URLTemplate: "jdbc:h2:mem:{database}",
		InstallPath: "/tmp/h2",
		Jars:        []config.JDBCJar{{Name: "h2.jar"}},
	}
	gateway.SetProfileResolver(func(ctx context.Context, cfg config.DatabaseConfig) (config.JDBCDriverProfile, error) {
		return profile, nil
	})

	err := gateway.ConnectDatabase(context.Background(), "db-test", config.DatabaseConfig{
		DBType:          "h2",
		Database:        "testdb",
		DriverProfileID: "h2-2.2.224",
	})
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	if client.openRequest.GetSessionId() != "db-test" {
		t.Fatalf("unexpected session id: %s", client.openRequest.GetSessionId())
	}
	if client.openRequest.GetProfile().GetDriverClass() != "org.h2.Driver" {
		t.Fatalf("unexpected driver class: %s", client.openRequest.GetProfile().GetDriverClass())
	}
	if len(client.openRequest.GetProfile().GetJarPaths()) != 1 {
		t.Fatalf("expected jar path")
	}
}

func TestJdbcGatewayExecuteQueryConvertsRows(t *testing.T) {
	client := &fakeJdbcAgentClient{
		queryResult: &jdbcproto.QueryResult{
			Columns: []string{"ID", "NAME"},
			Rows: []*jdbcproto.Row{
				{Values: []string{"1", "ok"}},
			},
			Affected: 0,
		},
	}
	gateway := NewJdbcGatewayService(client, "secret")

	result, err := gateway.ExecuteQuery(context.Background(), "db-test", "select 1")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if result.Columns[0] != "ID" || result.Rows[0][1] != "ok" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestJdbcGatewayMapsDriverMissingError(t *testing.T) {
	client := &fakeJdbcAgentClient{openErr: errors.New("DRIVER_MISSING: h2")}
	gateway := NewJdbcGatewayService(client, "secret")

	gateway.SetProfileResolver(func(ctx context.Context, cfg config.DatabaseConfig) (config.JDBCDriverProfile, error) {
		return config.JDBCDriverProfile{}, errors.New("DRIVER_MISSING: h2")
	})

	err := gateway.ConnectDatabase(context.Background(), "db-test", config.DatabaseConfig{})
	if err == nil {
		t.Fatalf("expected error")
	}
	var jdbcErr *JDBCError
	if !errors.As(err, &jdbcErr) {
		t.Fatalf("expected JDBCError, got %T", err)
	}
	if jdbcErr.Code != "DRIVER_MISSING" {
		t.Fatalf("unexpected code: %s", jdbcErr.Code)
	}
}

type fakeJdbcAgentClient struct {
	openRequest    *jdbcproto.OpenSessionRequest
	columnsRequest *jdbcproto.ListColumnsRequest
	schemasRequest *jdbcproto.ListSchemasRequest
	querySQL       string
	queryResult    *jdbcproto.QueryResult
	openErr        error
}

func (f *fakeJdbcAgentClient) OpenSession(ctx context.Context, request *jdbcproto.OpenSessionRequest) (*jdbcproto.OpenSessionResponse, error) {
	f.openRequest = request
	if f.openErr != nil {
		return nil, f.openErr
	}
	return &jdbcproto.OpenSessionResponse{SessionId: request.GetSessionId()}, nil
}

func (f *fakeJdbcAgentClient) ExecuteQuery(ctx context.Context, request *jdbcproto.ExecuteQueryRequest) (*jdbcproto.QueryResult, error) {
	f.querySQL = request.GetSql()
	if f.queryResult != nil {
		return f.queryResult, nil
	}
	return &jdbcproto.QueryResult{}, nil
}

func (f *fakeJdbcAgentClient) ListSchemas(ctx context.Context, request *jdbcproto.ListSchemasRequest) (*jdbcproto.ListSchemasResponse, error) {
	f.schemasRequest = request
	return &jdbcproto.ListSchemasResponse{Schemas: []string{"PUBLIC"}}, nil
}

func (f *fakeJdbcAgentClient) ListTables(ctx context.Context, request *jdbcproto.ListTablesRequest) (*jdbcproto.ListTablesResponse, error) {
	return &jdbcproto.ListTablesResponse{}, nil
}

func (f *fakeJdbcAgentClient) ListColumns(ctx context.Context, request *jdbcproto.ListColumnsRequest) (*jdbcproto.ListColumnsResponse, error) {
	f.columnsRequest = request
	return &jdbcproto.ListColumnsResponse{}, nil
}

func TestJdbcGatewayGetTableSchemaInSchemaPassesSchemaToAgent(t *testing.T) {
	client := &fakeJdbcAgentClient{}
	gateway := NewJdbcGatewayService(client, "test-token")

	_, err := gateway.GetTableSchemaInSchema(context.Background(), "session-1", "pems", "users")
	if err != nil {
		t.Fatalf("get table schema failed: %v", err)
	}
	if client.columnsRequest == nil {
		t.Fatal("expected ListColumns request")
	}
	if client.columnsRequest.GetSchema() != "pems" {
		t.Fatalf("schema = %q, want pems", client.columnsRequest.GetSchema())
	}
}

func TestJdbcGatewayListSchemasPassesCatalogToAgent(t *testing.T) {
	client := &fakeJdbcAgentClient{}
	gateway := NewJdbcGatewayService(client, "test-token")

	schemas, err := gateway.ListSchemas(context.Background(), "session-1", "app")
	if err != nil {
		t.Fatalf("list schemas failed: %v", err)
	}
	if len(schemas) != 1 || schemas[0] != "PUBLIC" {
		t.Fatalf("schemas = %v", schemas)
	}
	if client.schemasRequest.GetCatalog() != "app" {
		t.Fatalf("catalog = %q, want app", client.schemasRequest.GetCatalog())
	}
}

func (f *fakeJdbcAgentClient) CloseSession(ctx context.Context, request *jdbcproto.CloseSessionRequest) (*jdbcproto.CloseSessionResponse, error) {
	return &jdbcproto.CloseSessionResponse{Closed: true}, nil
}
