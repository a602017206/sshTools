package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	redisScanPageSize = 200
	redisScanKeyLimit = 1000
	redisPreviewLimit = 4096
	redisBatchDeleteLimit = 100
)

type RedisNativeClient interface {
	Ping(context.Context) error
	Keyspace(context.Context) (map[int]int, error)
	Scan(context.Context, uint64, string, int, int64) ([]string, uint64, error)
	DescribeKey(context.Context, string) (NativeResourceDetails, error)
	SetKey(context.Context, string, string) (NativeMutationResult, error)
	SaveKeyValue(context.Context, string, string) (NativeMutationResult, error)
	DeleteKey(context.Context, string) (NativeMutationResult, error)
	DeleteKeys(context.Context, []string) (NativeMutationResult, error)
	DoCommand(context.Context, []any) (NativeQueryResult, error)
	Close() error
}

type RedisNativeClientFactory interface {
	New(NativeDatabaseConfig, int) (RedisNativeClient, error)
}

type RedisNativeProvider struct {
	factory RedisNativeClientFactory
}

func NewRedisNativeProvider(factory RedisNativeClientFactory) *RedisNativeProvider {
	return &RedisNativeProvider{factory: factory}
}

func NewDefaultRedisNativeProvider() *RedisNativeProvider {
	return NewRedisNativeProvider(redisGoClientFactory{})
}

func (p *RedisNativeProvider) Test(ctx context.Context, cfg NativeDatabaseConfig) error {
	database, err := redisDatabaseNumber(cfg.Database)
	if err != nil {
		return err
	}
	client, err := p.factory.New(cfg, database)
	if err != nil {
		return err
	}
	defer client.Close()
	return client.Ping(ctx)
}

func (p *RedisNativeProvider) Connect(ctx context.Context, cfg NativeDatabaseConfig) (NativeDatabaseClient, error) {
	database, err := redisDatabaseNumber(cfg.Database)
	if err != nil {
		return nil, err
	}
	client, err := p.factory.New(cfg, database)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx); err != nil {
		_ = client.Close()
		return nil, err
	}
	return &redisNativeSession{factory: p.factory, config: cfg, database: database, client: client}, nil
}

type redisNativeSession struct {
	factory  RedisNativeClientFactory
	config   NativeDatabaseConfig
	database int
	client   RedisNativeClient
}

func (s *redisNativeSession) ListPrimaryResources(ctx context.Context) ([]NativeResource, error) {
	keyspace, err := s.client.Keyspace(ctx)
	if err != nil {
		return nil, err
	}
	databases := make([]int, 0, len(keyspace))
	for database := range keyspace {
		databases = append(databases, database)
	}
	if len(databases) == 0 {
		databases = append(databases, s.database)
	}
	sort.Ints(databases)
	resources := make([]NativeResource, 0, len(databases))
	for _, database := range databases {
		resources = append(resources, NativeResource{Kind: NativeResourceKindDatabase, Name: strconv.Itoa(database)})
	}
	return resources, nil
}

func (s *redisNativeSession) redisClientFor(ctx context.Context, parent string) (RedisNativeClient, func(), error) {
	database, err := redisDatabaseNumber(parent)
	if err != nil {
		return nil, nil, err
	}
	if database == s.database {
		return s.client, func() {}, nil
	}
	client, err := s.factory.New(s.config, database)
	if err != nil {
		return nil, nil, err
	}
	return client, func() { _ = client.Close() }, nil
}

func (s *redisNativeSession) ListSecondaryResources(ctx context.Context, parent string) ([]NativeResource, error) {
	client, cleanup, err := s.redisClientFor(ctx, parent)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	keys := make([]string, 0)
	var cursor uint64
	for len(keys) < redisScanKeyLimit {
		page, next, scanErr := client.Scan(ctx, cursor, "*", redisScanPageSize, redisScanKeyLimit-int64(len(keys)))
		if scanErr != nil {
			return nil, scanErr
		}
		keys = append(keys, page...)
		cursor = next
		if cursor == 0 {
			break
		}
	}
	sort.Strings(keys)
	resources := make([]NativeResource, 0, len(keys))
	for _, key := range keys {
		resources = append(resources, NativeResource{Kind: NativeResourceKindKey, Name: key})
	}
	return resources, nil
}

func (s *redisNativeSession) ListSecondaryResourcesPage(ctx context.Context, parent, pattern, cursor string, limit int) (NativeResourcePage, error) {
	client, cleanup, err := s.redisClientFor(ctx, parent)
	if err != nil {
		return NativeResourcePage{}, err
	}
	defer cleanup()
	if limit <= 0 || limit > redisScanPageSize {
		limit = redisScanPageSize
	}
	match := strings.TrimSpace(pattern)
	if match == "" {
		match = "*"
	}
	start, _ := strconv.ParseUint(strings.TrimSpace(cursor), 10, 64)
	keys, next, scanErr := client.Scan(ctx, start, match, limit, int64(limit))
	if scanErr != nil {
		return NativeResourcePage{}, scanErr
	}
	sort.Strings(keys)
	items := make([]NativeResource, 0, len(keys))
	for _, key := range keys {
		items = append(items, NativeResource{Kind: NativeResourceKindKey, Name: key})
	}
	return NativeResourcePage{
		Items:      items,
		NextCursor: strconv.FormatUint(next, 10),
		HasMore:    next != 0,
		Truncated:  false,
	}, nil
}

func (s *redisNativeSession) Close() error {
	return s.client.Close()
}

func (s *redisNativeSession) DescribeResource(ctx context.Context, parent, name string) (NativeResourceDetails, error) {
	client, cleanup, err := s.redisClientFor(ctx, parent)
	if err != nil {
		return NativeResourceDetails{}, err
	}
	defer cleanup()
	return client.DescribeKey(ctx, name)
}

func redisDatabaseNumber(value string) (int, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, nil
	}
	database, err := strconv.Atoi(value)
	if err != nil || database < 0 {
		return 0, fmt.Errorf("Redis 逻辑数据库必须是非负整数: %s", value)
	}
	return database, nil
}

type redisGoClientFactory struct{}

func (redisGoClientFactory) New(cfg NativeDatabaseConfig, database int) (RedisNativeClient, error) {
	host := strings.TrimSpace(cfg.Host)
	if host == "" {
		host = "127.0.0.1"
	}
	port := cfg.Port
	if port == 0 {
		port = 6379
	}
	return &redisGoClient{client: redis.NewClient(&redis.Options{
		Addr:     net.JoinHostPort(host, strconv.Itoa(port)),
		Username: cfg.User,
		Password: cfg.Password,
		DB:       database,
	})}, nil
}

type redisGoClient struct {
	client *redis.Client
}

func (c *redisGoClient) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *redisGoClient) Keyspace(ctx context.Context) (map[int]int, error) {
	info, err := c.client.Info(ctx, "keyspace").Result()
	if err != nil {
		return nil, err
	}
	keyspace := make(map[int]int)
	for _, line := range strings.Split(info, "\n") {
		parts := strings.SplitN(strings.TrimSpace(line), ":", 2)
		if len(parts) != 2 || !strings.HasPrefix(parts[0], "db") {
			continue
		}
		database, parseErr := strconv.Atoi(strings.TrimPrefix(parts[0], "db"))
		if parseErr != nil || database < 0 {
			continue
		}
		keyspace[database] = 0
	}
	return keyspace, nil
}

func (c *redisGoClient) Scan(ctx context.Context, cursor uint64, match string, count int, remaining int64) ([]string, uint64, error) {
	if remaining < int64(count) {
		count = int(remaining)
	}
	if strings.TrimSpace(match) == "" {
		match = "*"
	}
	keys, next, err := c.client.Scan(ctx, cursor, match, int64(count)).Result()
	return keys, next, err
}

func (c *redisGoClient) Close() error {
	return c.client.Close()
}

func (c *redisGoClient) DescribeKey(ctx context.Context, name string) (NativeResourceDetails, error) {
	kind, err := c.client.Type(ctx, name).Result()
	if err != nil {
		return NativeResourceDetails{}, err
	}
	ttl, err := c.client.TTL(ctx, name).Result()
	if err != nil {
		return NativeResourceDetails{}, err
	}
	data := map[string]any{"type": kind, "ttlSeconds": ttlSeconds(ttl)}
	if err := c.readKeyPreview(ctx, kind, name, data); err != nil {
		return NativeResourceDetails{}, err
	}
	content, err := json.Marshal(data)
	if err != nil {
		return NativeResourceDetails{}, err
	}
	return NativeResourceDetails{Kind: NativeResourceKindKey, Name: name, Summary: redisKeySummary(kind, ttl), Content: string(content)}, nil
}

func ttlSeconds(ttl time.Duration) int64 {
	if ttl == -1 {
		return -1
	}
	if ttl == -2 {
		return -2
	}
	return int64(ttl.Seconds())
}
