package service

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"AHaSSHTools/internal/config"
	"AHaSSHTools/internal/service/jdbcproto"
)

type JdbcAgentClient interface {
	OpenSession(ctx context.Context, request *jdbcproto.OpenSessionRequest) (*jdbcproto.OpenSessionResponse, error)
	ExecuteQuery(ctx context.Context, request *jdbcproto.ExecuteQueryRequest) (*jdbcproto.QueryResult, error)
	ListTables(ctx context.Context, request *jdbcproto.ListTablesRequest) (*jdbcproto.ListTablesResponse, error)
	ListColumns(ctx context.Context, request *jdbcproto.ListColumnsRequest) (*jdbcproto.ListColumnsResponse, error)
	CloseSession(ctx context.Context, request *jdbcproto.CloseSessionRequest) (*jdbcproto.CloseSessionResponse, error)
}

type JdbcGatewayService struct {
	client          JdbcAgentClient
	token           string
	profileResolver func(context.Context, config.DatabaseConfig) (config.JDBCDriverProfile, error)
}

func NewJdbcGatewayService(client JdbcAgentClient, token string) *JdbcGatewayService {
	return &JdbcGatewayService{client: client, token: token}
}

func (s *JdbcGatewayService) SetProfileResolver(resolver func(context.Context, config.DatabaseConfig) (config.JDBCDriverProfile, error)) {
	s.profileResolver = resolver
}

func (s *JdbcGatewayService) ConnectDatabase(ctx context.Context, sessionID string, cfg config.DatabaseConfig) error {
	if s.client == nil {
		return MapJDBCAgentError("AGENT_UNAVAILABLE: JDBC agent client not configured")
	}
	if s.profileResolver == nil {
		return MapJDBCAgentError("DRIVER_MISSING: JDBC driver profile resolver not configured")
	}
	profile, err := s.profileResolver(ctx, cfg)
	if err != nil {
		return mapJdbcGatewayError(err)
	}
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
	if s.client == nil {
		return nil, MapJDBCAgentError("AGENT_UNAVAILABLE: JDBC agent client not configured")
	}
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
	if s.client == nil {
		return nil, MapJDBCAgentError("AGENT_UNAVAILABLE: JDBC agent client not configured")
	}
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
	if s.client == nil {
		return nil, MapJDBCAgentError("AGENT_UNAVAILABLE: JDBC agent client not configured")
	}
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
			ColumnSize: int(column.GetColumnSize()),
			DecimalDigits: int(column.GetDecimalDigits()),
			DefaultValue: column.GetDefaultValue(),
			HasDefault: column.GetHasDefault(),
		})
	}
	return &config.TableSchema{TableName: table, Columns: columns}, nil
}

func (s *JdbcGatewayService) CloseDatabase(ctx context.Context, sessionID string) error {
	if s.client == nil {
		return MapJDBCAgentError("AGENT_UNAVAILABLE: JDBC agent client not configured")
	}
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
	var jdbcErr *JDBCError
	if errors.As(err, &jdbcErr) {
		return err
	}
	return newJDBCError(err.Error(), err)
}
