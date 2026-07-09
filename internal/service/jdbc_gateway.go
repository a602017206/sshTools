package service

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"AHaSSHTools/internal/config"
	"AHaSSHTools/internal/service/jdbcproto"
)

type JDBCError struct {
	Code    string
	Message string
	Err     error
}

func (e *JDBCError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Code
}

func (e *JDBCError) Unwrap() error {
	return e.Err
}

type JdbcAgentClient interface {
	OpenSession(ctx context.Context, request *jdbcproto.OpenSessionRequest) (*jdbcproto.OpenSessionResponse, error)
	ExecuteQuery(ctx context.Context, request *jdbcproto.ExecuteQueryRequest) (*jdbcproto.QueryResult, error)
	ListTables(ctx context.Context, request *jdbcproto.ListTablesRequest) (*jdbcproto.ListTablesResponse, error)
	ListColumns(ctx context.Context, request *jdbcproto.ListColumnsRequest) (*jdbcproto.ListColumnsResponse, error)
	CloseSession(ctx context.Context, request *jdbcproto.CloseSessionRequest) (*jdbcproto.CloseSessionResponse, error)
}

type JdbcGatewayService struct {
	client JdbcAgentClient
	token  string
}

func NewJdbcGatewayService(client JdbcAgentClient, token string) *JdbcGatewayService {
	return &JdbcGatewayService{client: client, token: token}
}

func (s *JdbcGatewayService) ConnectDatabase(ctx context.Context, sessionID string, cfg config.DatabaseConfig, profile config.JDBCDriverProfile) error {
	request := &jdbcproto.OpenSessionRequest{
		Token:     s.token,
		SessionId: sessionID,
		Profile: &jdbcproto.DriverProfile{
			Id:          profile.ID,
			DriverClass: profile.DriverClass,
			UrlTemplate: profile.URLTemplate,
			JarPaths:    profileJarPaths(profile),
		},
		Host:       cfg.Host,
		Port:       int32(cfg.Port),
		Database:   cfg.Database,
		User:       cfg.User,
		Password:   cfg.Password,
		Properties: cfg.Properties,
	}
	if _, err := s.client.OpenSession(ctx, request); err != nil {
		return mapJdbcGatewayError(err)
	}
	return nil
}

func (s *JdbcGatewayService) ExecuteQuery(ctx context.Context, sessionID string, query string) (*QueryResult, error) {
	result, err := s.client.ExecuteQuery(ctx, &jdbcproto.ExecuteQueryRequest{
		Token:     s.token,
		SessionId: sessionID,
		Sql:       query,
	})
	if err != nil {
		return nil, mapJdbcGatewayError(err)
	}
	rows := make([][]interface{}, 0, len(result.GetRows()))
	for _, row := range result.GetRows() {
		values := make([]interface{}, 0, len(row.GetValues()))
		for _, value := range row.GetValues() {
			values = append(values, value)
		}
		rows = append(rows, values)
	}
	return &QueryResult{
		Columns:  result.GetColumns(),
		Rows:     rows,
		Affected: int(result.GetAffected()),
	}, nil
}

func (s *JdbcGatewayService) ListTables(ctx context.Context, sessionID, database string) ([]string, error) {
	result, err := s.client.ListTables(ctx, &jdbcproto.ListTablesRequest{
		Token:     s.token,
		SessionId: sessionID,
		Catalog:   database,
	})
	if err != nil {
		return nil, mapJdbcGatewayError(err)
	}
	return result.GetTables(), nil
}

func (s *JdbcGatewayService) ListDatabases(context.Context, string) ([]string, error) {
	return nil, fmt.Errorf("JDBC agent 暂未实现数据库列表")
}

func (s *JdbcGatewayService) GetTableSchema(ctx context.Context, sessionID, table string) (*config.TableSchema, error) {
	result, err := s.client.ListColumns(ctx, &jdbcproto.ListColumnsRequest{
		Token:     s.token,
		SessionId: sessionID,
		Table:     table,
	})
	if err != nil {
		return nil, mapJdbcGatewayError(err)
	}
	columns := make([]config.ColumnSchema, 0, len(result.GetColumns()))
	for _, column := range result.GetColumns() {
		columns = append(columns, config.ColumnSchema{
			Name:         column.GetName(),
			Type:         column.GetType(),
			Nullable:     column.GetNullable(),
			IsPrimaryKey: column.GetPrimaryKey(),
		})
	}
	return &config.TableSchema{TableName: table, Columns: columns}, nil
}

func (s *JdbcGatewayService) CloseDatabase(ctx context.Context, sessionID string) error {
	_, err := s.client.CloseSession(ctx, &jdbcproto.CloseSessionRequest{
		Token:     s.token,
		SessionId: sessionID,
	})
	if err != nil {
		return mapJdbcGatewayError(err)
	}
	return nil
}

func profileJarPaths(profile config.JDBCDriverProfile) []string {
	paths := make([]string, 0, len(profile.Jars))
	for _, jar := range profile.Jars {
		if filepath.IsAbs(jar.Name) {
			paths = append(paths, jar.Name)
			continue
		}
		paths = append(paths, filepath.Join(profile.InstallPath, "jars", jar.Name))
	}
	return paths
}

func mapJdbcGatewayError(err error) error {
	if err == nil {
		return nil
	}
	message := err.Error()
	code := "DB_CONNECT_FAILED"
	if strings.Contains(message, "DRIVER_MISSING") {
		code = "DRIVER_MISSING"
	}
	return &JDBCError{Code: code, Message: message, Err: err}
}
