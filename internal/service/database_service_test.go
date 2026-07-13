package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"AHaSSHTools/internal/config"
	"github.com/DATA-DOG/go-sqlmock"
)

func newMockDatabaseService(t *testing.T, dbType, database string) (*DatabaseService, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	ds := NewDatabaseService(nil)
	ds.openFunc = func(driverName, dsn string) (*sql.DB, error) {
		return db, nil
	}

	ds.sessionStore["db-test"] = &DatabaseSession{
		ID:        "db-test",
		Config:    config.DatabaseConfig{DBType: dbType, Database: database},
		DB:        db,
		Connected: true,
	}

	cleanup := func() {
		_ = db.Close()
	}

	return ds, mock, cleanup
}

func TestDatabaseService_ExecuteQuery_ExecsNonSelect(t *testing.T) {
	ds, mock, cleanup := newMockDatabaseService(t, "mysql", "")
	defer cleanup()

	mock.ExpectExec("UPDATE users SET active=1").WillReturnResult(sqlmock.NewResult(0, 3))

	result, err := ds.ExecuteQuery("db-test", "UPDATE users SET active=1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Affected != 3 {
		t.Fatalf("expected affected=3, got %d", result.Affected)
	}
}

func TestDatabaseService_ExecuteQuery_SelectReturnsRows(t *testing.T) {
	ds, mock, cleanup := newMockDatabaseService(t, "mysql", "")
	defer cleanup()

	rows := sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "alice").AddRow(2, "bob")
	mock.ExpectQuery("SELECT id, name FROM users").WillReturnRows(rows)

	result, err := ds.ExecuteQuery("db-test", "SELECT id, name FROM users")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(result.Rows))
	}
}

func TestDatabaseService_CloseDatabase_RemovesSession(t *testing.T) {
	ds, mock, cleanup := newMockDatabaseService(t, "mysql", "")
	defer cleanup()

	mock.ExpectClose()

	if err := ds.CloseDatabase("db-test"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := ds.GetSession("db-test"); err == nil {
		t.Fatalf("expected session to be removed")
	}
}

func TestDatabaseServiceDelegatesConnectToJdbcGateway(t *testing.T) {
	gateway := &fakeJdbcGateway{}
	ds := NewDatabaseServiceWithGateway(nil, gateway)
	err := ds.ConnectDatabase("db-test", "localhost", 1521, "scott", "tiger", "oracle", "orcl")
	if err != nil {
		t.Fatalf("connect failed: %v", err)
	}
	if gateway.lastDBType != "oracle" {
		t.Fatalf("expected oracle, got %s", gateway.lastDBType)
	}
}

func TestDatabaseServiceTestConnectionUsesGateway(t *testing.T) {
	gateway := &fakeJdbcGateway{}
	ds := NewDatabaseServiceWithGateway(nil, gateway)

	err := ds.TestConnection("db.example", 1521, "scott", "tiger", "oracle", "orcl")
	if err != nil {
		t.Fatalf("test connection failed: %v", err)
	}
	if gateway.connectCalls != 1 || gateway.closeCalls != 1 {
		t.Fatalf("expected temporary gateway session, connect=%d close=%d", gateway.connectCalls, gateway.closeCalls)
	}
	if gateway.lastDBType != "oracle" {
		t.Fatalf("expected oracle gateway config, got %s", gateway.lastDBType)
	}
}

func TestDatabaseServiceTestConnectionAddsMySQLHandshakeTimeouts(t *testing.T) {
	gateway := &fakeJdbcGateway{}
	ds := NewDatabaseServiceWithGateway(nil, gateway)

	if err := ds.TestConnection("db.example", 3306, "root", "secret", "mysql", "app"); err != nil {
		t.Fatal(err)
	}
	if gateway.lastConfig.Properties["connectTimeout"] != "8000" {
		t.Fatalf("unexpected MySQL connect timeout: %+v", gateway.lastConfig.Properties)
	}
	if gateway.lastConfig.Properties["socketTimeout"] != "8000" {
		t.Fatalf("unexpected MySQL socket timeout: %+v", gateway.lastConfig.Properties)
	}
}

func TestDatabaseServiceTestConnectionExplainsMySQLHandshakeTimeout(t *testing.T) {
	gateway := &fakeJdbcGateway{connectErr: context.DeadlineExceeded}
	ds := NewDatabaseServiceWithGateway(nil, gateway)

	err := ds.TestConnection("192.168.121.158", 22306, "root", "secret", "mysql", "")
	var jdbcErr *JDBCError
	if !errors.As(err, &jdbcErr) || jdbcErr.Code != JDBCErrorDBConnectFailed {
		t.Fatalf("expected DB_CONNECT_FAILED, got %v", err)
	}
	for _, expected := range []string{"192.168.121.158:22306", "MySQL 服务端握手", "TCP 代理或端口转发"} {
		if !strings.Contains(err.Error(), expected) {
			t.Fatalf("error missing %q: %v", expected, err)
		}
	}
}

func TestDatabaseServiceTestConnectionExplainsMySQLNoPacketsError(t *testing.T) {
	gateway := &fakeJdbcGateway{connectErr: errors.New("Communications link failure: The driver has not received any packets from the server")}
	ds := NewDatabaseServiceWithGateway(nil, gateway)

	err := ds.TestConnection("db.example", 22306, "root", "secret", "mysql", "")
	if err == nil || !strings.Contains(err.Error(), "未收到 MySQL 服务端握手") {
		t.Fatalf("expected MySQL handshake diagnosis, got %v", err)
	}
}

type fakeJdbcGateway struct {
	lastDBType   string
	lastConfig   config.DatabaseConfig
	lastQuery    string
	queryResult  *QueryResult
	connectCalls int
	closeCalls   int
	connectErr   error
}

func (g *fakeJdbcGateway) ConnectDatabase(ctx context.Context, sessionID string, cfg config.DatabaseConfig) error {
	g.connectCalls++
	g.lastDBType = cfg.DBType
	g.lastConfig = cfg
	return g.connectErr
}

func (g *fakeJdbcGateway) ExecuteQuery(ctx context.Context, sessionID string, query string) (*QueryResult, error) {
	g.lastQuery = query
	if g.queryResult != nil {
		return g.queryResult, nil
	}
	return &QueryResult{}, nil
}

func (g *fakeJdbcGateway) ListTables(ctx context.Context, sessionID, database string) ([]string, error) {
	return []string{}, nil
}

func (g *fakeJdbcGateway) ListDatabases(ctx context.Context, sessionID string) ([]string, error) {
	return []string{}, nil
}

func (g *fakeJdbcGateway) GetTableSchema(ctx context.Context, sessionID, table string) (*config.TableSchema, error) {
	return &config.TableSchema{TableName: table}, nil
}

func (g *fakeJdbcGateway) CloseDatabase(ctx context.Context, sessionID string) error {
	g.closeCalls++
	return nil
}

func TestDatabaseServiceGetTableDDLUsesJDBCGatewayForMySQL(t *testing.T) {
	gateway := &fakeJdbcGateway{queryResult: &QueryResult{
		Columns: []string{"Table", "Create Table"},
		Rows:    [][]interface{}{{"users", "CREATE TABLE `users` (`id` bigint NOT NULL)"}},
	}}
	ds := NewDatabaseServiceWithGateway(nil, gateway)
	ds.sessionStore["mysql-session"] = &DatabaseSession{
		ID:        "mysql-session",
		Config:    config.DatabaseConfig{DBType: "mysql"},
		Connected: true,
	}

	ddl, err := ds.GetTableDDL("mysql-session", "app", "users")
	if err != nil {
		t.Fatalf("get table DDL failed: %v", err)
	}
	if gateway.lastQuery != "SHOW CREATE TABLE `app`.`users`" {
		t.Fatalf("unexpected query: %s", gateway.lastQuery)
	}
	if ddl.TableName != "users" || ddl.DDL != "CREATE TABLE `users` (`id` bigint NOT NULL)" || ddl.DBType != "mysql" {
		t.Fatalf("unexpected DDL: %+v", ddl)
	}
}

func TestDatabaseService_ListDatabases_MySQL(t *testing.T) {
	ds, mock, cleanup := newMockDatabaseService(t, "mysql", "")
	defer cleanup()

	rows := sqlmock.NewRows([]string{"Database"}).AddRow("db1").AddRow("db2")
	mock.ExpectQuery("SHOW DATABASES").WillReturnRows(rows)

	result, err := ds.ListDatabases("db-test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 databases, got %d", len(result))
	}
}

func TestDatabaseService_ListDatabases_Postgres(t *testing.T) {
	ds, mock, cleanup := newMockDatabaseService(t, "postgresql", "postgres")
	defer cleanup()

	rows := sqlmock.NewRows([]string{"datname"}).AddRow("postgres").AddRow("appdb")
	mock.ExpectQuery("SELECT datname FROM pg_database").WillReturnRows(rows)

	result, err := ds.ListDatabases("db-test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 databases, got %d", len(result))
	}
}

func TestDatabaseService_ListTablesInDatabase_MySQL(t *testing.T) {
	ds, mock, cleanup := newMockDatabaseService(t, "mysql", "")
	defer cleanup()

	rows := sqlmock.NewRows([]string{"Tables_in_appdb"}).AddRow("users").AddRow("orders")
	mock.ExpectQuery("SHOW TABLES FROM `appdb`").WillReturnRows(rows)

	result, err := ds.ListTablesInDatabase("db-test", "appdb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 tables, got %d", len(result))
	}
}

func TestDatabaseService_ListTablesInDatabase_PostgresCurrent(t *testing.T) {
	ds, mock, cleanup := newMockDatabaseService(t, "postgresql", "appdb")
	defer cleanup()

	rows := sqlmock.NewRows([]string{"tablename"}).AddRow("users")
	mock.ExpectQuery("SELECT tablename FROM pg_tables").WillReturnRows(rows)

	result, err := ds.ListTablesInDatabase("db-test", "appdb")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 table, got %d", len(result))
	}
}
