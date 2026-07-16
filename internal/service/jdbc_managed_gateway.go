package service

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"AHaSSHTools/internal/config"
)

type JDBCAgentConnectionProvider interface {
	Client(ctx context.Context) (*JDBCAgentConnection, error)
	Restart(ctx context.Context) (*JDBCAgentConnection, error)
}

type ManagedJDBCGateway struct {
	supervisor      JDBCAgentConnectionProvider
	profileResolver func(context.Context, config.DatabaseConfig) (config.JDBCDriverProfile, error)

	mu       sync.RWMutex
	sessions map[string]config.DatabaseConfig
}

func NewManagedJDBCGateway(supervisor JDBCAgentConnectionProvider) *ManagedJDBCGateway {
	return &ManagedJDBCGateway{
		supervisor: supervisor,
		sessions:   make(map[string]config.DatabaseConfig),
	}
}

func (s *ManagedJDBCGateway) SetProfileResolver(resolver func(context.Context, config.DatabaseConfig) (config.JDBCDriverProfile, error)) {
	s.profileResolver = resolver
}

func (s *ManagedJDBCGateway) ConnectDatabase(ctx context.Context, sessionID string, cfg config.DatabaseConfig) error {
	gateway, err := s.gateway(ctx, false)
	if err != nil {
		return err
	}
	err = gateway.ConnectDatabase(ctx, sessionID, cfg)
	if isJDBCAgentUnavailable(err) {
		gateway, restartErr := s.gateway(ctx, true)
		if restartErr != nil {
			return restartErr
		}
		err = gateway.ConnectDatabase(ctx, sessionID, cfg)
	}
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.sessions[sessionID] = cfg
	s.mu.Unlock()
	return nil
}

func (s *ManagedJDBCGateway) ExecuteQuery(ctx context.Context, sessionID, query string) (*QueryResult, error) {
	return managedGatewayCall(s, ctx, sessionID, func(gateway *JdbcGatewayService) (*QueryResult, error) {
		return gateway.ExecuteQuery(ctx, sessionID, query)
	})
}

func (s *ManagedJDBCGateway) ListTables(ctx context.Context, sessionID, database string) ([]string, error) {
	return managedGatewayCall(s, ctx, sessionID, func(gateway *JdbcGatewayService) ([]string, error) {
		return gateway.ListTables(ctx, sessionID, database)
	})
}

func (s *ManagedJDBCGateway) ListTablesInSchema(ctx context.Context, sessionID, database, schema string) ([]string, error) {
	return s.ListObjects(ctx, sessionID, database, schema, nil)
}

func (s *ManagedJDBCGateway) ListObjects(ctx context.Context, sessionID, database, schema string, types []string) ([]string, error) {
	return managedGatewayCall(s, ctx, sessionID, func(gateway *JdbcGatewayService) ([]string, error) {
		return gateway.ListObjects(ctx, sessionID, database, schema, types)
	})
}

func (s *ManagedJDBCGateway) ListSchemas(ctx context.Context, sessionID, database string) ([]string, error) {
	return managedGatewayCall(s, ctx, sessionID, func(gateway *JdbcGatewayService) ([]string, error) {
		return gateway.ListSchemas(ctx, sessionID, database)
	})
}

func (s *ManagedJDBCGateway) ListDatabases(ctx context.Context, sessionID string) ([]string, error) {
	cfg, ok := s.sessionConfig(sessionID)
	if !ok {
		return nil, fmt.Errorf("JDBC session 不存在: %s", sessionID)
	}

	var query string
	switch cfg.DBType {
	case "mysql":
		query = "SHOW DATABASES"
	case "postgresql", "kingbase":
		query = "SELECT datname FROM pg_database WHERE datistemplate = false ORDER BY datname"
	default:
		return nil, fmt.Errorf("数据库类型 %s 暂不支持列出数据库", cfg.DBType)
	}

	result, err := s.ExecuteQuery(ctx, sessionID, query)
	if err != nil {
		return nil, err
	}
	databases := make([]string, 0, len(result.Rows))
	for _, row := range result.Rows {
		if len(row) > 0 {
			databases = append(databases, fmt.Sprint(row[0]))
		}
	}
	return databases, nil
}

func (s *ManagedJDBCGateway) GetTableSchema(ctx context.Context, sessionID, table string) (*config.TableSchema, error) {
	return s.GetTableSchemaInSchema(ctx, sessionID, "", table)
}

// GetTableSchemaInSchema delegates schema-scoped metadata loading to the active JDBC agent.
func (s *ManagedJDBCGateway) GetTableSchemaInSchema(ctx context.Context, sessionID, schema, table string) (*config.TableSchema, error) {
	return managedGatewayCall(s, ctx, sessionID, func(gateway *JdbcGatewayService) (*config.TableSchema, error) {
		return gateway.GetTableSchemaInSchema(ctx, sessionID, schema, table)
	})
}

func (s *ManagedJDBCGateway) CloseDatabase(ctx context.Context, sessionID string) error {
	gateway, err := s.gateway(ctx, false)
	if err == nil {
		err = gateway.CloseDatabase(ctx, sessionID)
	}
	s.mu.Lock()
	delete(s.sessions, sessionID)
	s.mu.Unlock()
	return err
}

// ActiveSessionConfigs returns a snapshot of JDBC session configurations.
func (s *ManagedJDBCGateway) ActiveSessionConfigs() map[string]config.DatabaseConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sessions := make(map[string]config.DatabaseConfig, len(s.sessions))
	for sessionID, cfg := range s.sessions {
		sessions[sessionID] = cfg
	}
	return sessions
}

func managedGatewayCall[T any](s *ManagedJDBCGateway, ctx context.Context, sessionID string, call func(*JdbcGatewayService) (T, error)) (T, error) {
	var zero T
	gateway, err := s.gateway(ctx, false)
	if err != nil {
		return zero, err
	}
	result, err := call(gateway)
	if !isJDBCAgentUnavailable(err) {
		return result, err
	}

	cfg, ok := s.sessionConfig(sessionID)
	if !ok {
		return zero, err
	}
	recoveredGateway, restartErr := s.gateway(ctx, true)
	if restartErr != nil {
		return zero, restartErr
	}
	if reopenErr := recoveredGateway.ConnectDatabase(ctx, sessionID, cfg); reopenErr != nil {
		return zero, fmt.Errorf("恢复 JDBC session 失败: %w", reopenErr)
	}
	return call(recoveredGateway)
}

func (s *ManagedJDBCGateway) gateway(ctx context.Context, restart bool) (*JdbcGatewayService, error) {
	if s.supervisor == nil {
		return nil, &JDBCError{Code: JDBCErrorAgentUnavailable, Message: "JDBC agent supervisor 未配置"}
	}
	var (
		connection *JDBCAgentConnection
		err        error
	)
	if restart {
		connection, err = s.supervisor.Restart(ctx)
	} else {
		connection, err = s.supervisor.Client(ctx)
	}
	if err != nil {
		return nil, err
	}
	if connection == nil || connection.Client == nil {
		return nil, &JDBCError{Code: JDBCErrorAgentUnavailable, Message: "JDBC agent client 不可用"}
	}
	gateway := NewJdbcGatewayService(connection.Client, connection.Token)
	gateway.SetProfileResolver(s.profileResolver)
	return gateway, nil
}

func (s *ManagedJDBCGateway) sessionConfig(sessionID string) (config.DatabaseConfig, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cfg, ok := s.sessions[sessionID]
	return cfg, ok
}

func isJDBCAgentUnavailable(err error) bool {
	var jdbcErr *JDBCError
	return errors.As(err, &jdbcErr) && jdbcErr.Code == JDBCErrorAgentUnavailable
}
