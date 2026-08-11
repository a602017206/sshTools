package service

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type NativeDatabaseType string

const (
	NativeDatabaseTypeRedis         NativeDatabaseType = "redis"
	NativeDatabaseTypeMongoDB       NativeDatabaseType = "mongodb"
	NativeDatabaseTypeElasticsearch NativeDatabaseType = "elasticsearch"
	NativeDatabaseTypeMemcached     NativeDatabaseType = "memcached"
	NativeDatabaseTypeCassandra     NativeDatabaseType = "cassandra"
	NativeDatabaseTypeCouchbase     NativeDatabaseType = "couchbase"
	NativeDatabaseTypeInfluxDB      NativeDatabaseType = "influxdb"
	NativeDatabaseTypeNeo4j         NativeDatabaseType = "neo4j"
	NativeDatabaseTypeKafka         NativeDatabaseType = "kafka"
)

type NativeResourceKind string

const (
	NativeResourceKindDatabase   NativeResourceKind = "database"
	NativeResourceKindKey        NativeResourceKind = "key"
	NativeResourceKindCollection NativeResourceKind = "collection"
	NativeResourceKindIndex      NativeResourceKind = "index"
	NativeResourceKindStatistic  NativeResourceKind = "statistic"
	NativeResourceKindTable      NativeResourceKind = "table"
)

type NativeDatabaseConfig struct {
	Type     NativeDatabaseType
	Host     string
	Port     int
	User     string
	Password string
	Database string
	Timeout  time.Duration
}

type NativeResource struct {
	Kind NativeResourceKind `json:"kind"`
	Name string             `json:"name"`
}

type NativeDatabaseProvider interface {
	Test(context.Context, NativeDatabaseConfig) error
	Connect(context.Context, NativeDatabaseConfig) (NativeDatabaseClient, error)
}

type NativeDatabaseClient interface {
	ListPrimaryResources(context.Context) ([]NativeResource, error)
	ListSecondaryResources(context.Context, string) ([]NativeResource, error)
	Close() error
}

type NativeDatabaseSession struct {
	ID     string
	Config NativeDatabaseConfig
	client NativeDatabaseClient
}

type NativeDatabaseService struct {
	providers map[NativeDatabaseType]NativeDatabaseProvider
	sessions  map[string]*NativeDatabaseSession
	mu        sync.RWMutex
}

func NewNativeDatabaseService(providers map[NativeDatabaseType]NativeDatabaseProvider) *NativeDatabaseService {
	registered := make(map[NativeDatabaseType]NativeDatabaseProvider, len(providers))
	for databaseType, provider := range providers {
		registered[databaseType] = provider
	}
	return &NativeDatabaseService{
		providers: registered,
		sessions:  make(map[string]*NativeDatabaseSession),
	}
}

func (s *NativeDatabaseService) TestConnection(ctx context.Context, cfg NativeDatabaseConfig) error {
	provider, err := s.provider(cfg.Type)
	if err != nil {
		return err
	}
	if err := provider.Test(ctx, cfg); err != nil {
		return fmt.Errorf("测试 %s 连接失败: %w", nativeDatabaseTypeName(cfg.Type), err)
	}
	return nil
}

func (s *NativeDatabaseService) Connect(ctx context.Context, sessionID string, cfg NativeDatabaseConfig) error {
	if sessionID == "" {
		return fmt.Errorf("原生数据库会话 ID 不能为空")
	}
	provider, err := s.provider(cfg.Type)
	if err != nil {
		return err
	}
	s.mu.RLock()
	_, exists := s.sessions[sessionID]
	s.mu.RUnlock()
	if exists {
		return fmt.Errorf("原生数据库会话已存在: %s", sessionID)
	}
	client, err := provider.Connect(ctx, cfg)
	if err != nil {
		return fmt.Errorf("连接 %s 失败: %w", nativeDatabaseTypeName(cfg.Type), err)
	}
	if client == nil {
		return fmt.Errorf("连接 %s 失败: 客户端不可用", nativeDatabaseTypeName(cfg.Type))
	}
	s.mu.Lock()
	s.sessions[sessionID] = &NativeDatabaseSession{ID: sessionID, Config: cfg, client: client}
	s.mu.Unlock()
	return nil
}

func (s *NativeDatabaseService) ListPrimaryResources(ctx context.Context, sessionID string) ([]NativeResource, error) {
	session, err := s.session(sessionID)
	if err != nil {
		return nil, err
	}
	resources, err := session.client.ListPrimaryResources(ctx)
	if err != nil {
		return nil, fmt.Errorf("读取 %s 资源失败: %w", nativeDatabaseTypeName(session.Config.Type), err)
	}
	return resources, nil
}

func (s *NativeDatabaseService) ListSecondaryResources(ctx context.Context, sessionID, parent string) ([]NativeResource, error) {
	session, err := s.session(sessionID)
	if err != nil {
		return nil, err
	}
	resources, err := session.client.ListSecondaryResources(ctx, parent)
	if err != nil {
		return nil, fmt.Errorf("读取 %s 子资源失败: %w", nativeDatabaseTypeName(session.Config.Type), err)
	}
	return resources, nil
}

func (s *NativeDatabaseService) Close(sessionID string) error {
	s.mu.Lock()
	session, exists := s.sessions[sessionID]
	if exists {
		delete(s.sessions, sessionID)
	}
	s.mu.Unlock()
	if !exists {
		return fmt.Errorf("原生数据库会话不存在: %s", sessionID)
	}
	if err := session.client.Close(); err != nil {
		return fmt.Errorf("关闭 %s 会话失败: %w", nativeDatabaseTypeName(session.Config.Type), err)
	}
	return nil
}

// ListSessions returns active native database session IDs.
func (s *NativeDatabaseService) ListSessions() []string {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := make([]string, 0, len(s.sessions))
	for id := range s.sessions {
		ids = append(ids, id)
	}
	return ids
}

// CloseAll closes every active native database session.
func (s *NativeDatabaseService) CloseAll() error {
	if s == nil {
		return nil
	}
	ids := s.ListSessions()
	var errs []error
	for _, id := range ids {
		if err := s.Close(id); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

func (s *NativeDatabaseService) provider(databaseType NativeDatabaseType) (NativeDatabaseProvider, error) {
	provider, exists := s.providers[databaseType]
	if !exists || provider == nil {
		return nil, fmt.Errorf("不支持的原生数据库类型: %s", databaseType)
	}
	return provider, nil
}

func (s *NativeDatabaseService) session(sessionID string) (*NativeDatabaseSession, error) {
	s.mu.RLock()
	session, exists := s.sessions[sessionID]
	s.mu.RUnlock()
	if !exists || session == nil {
		return nil, fmt.Errorf("原生数据库会话不存在: %s", sessionID)
	}
	return session, nil
}

func nativeDatabaseTypeName(databaseType NativeDatabaseType) string {
	switch databaseType {
	case NativeDatabaseTypeRedis:
		return "Redis"
	case NativeDatabaseTypeMongoDB:
		return "MongoDB"
	case NativeDatabaseTypeElasticsearch:
		return "Elasticsearch"
	case NativeDatabaseTypeMemcached:
		return "Memcached"
	case NativeDatabaseTypeCassandra:
		return "Cassandra"
	case NativeDatabaseTypeCouchbase:
		return "Couchbase"
	case NativeDatabaseTypeInfluxDB:
		return "InfluxDB"
	case NativeDatabaseTypeNeo4j:
		return "Neo4j"
	case NativeDatabaseTypeKafka:
		return "Kafka"
	default:
		return string(databaseType)
	}
}
