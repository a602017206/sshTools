package service

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/gocql/gocql"
)

type CassandraNativeClient interface {
	Ping(context.Context) error
	ListKeyspaces(context.Context) ([]string, error)
	ListTables(context.Context, string) ([]string, error)
	Close() error
}

type CassandraNativeClientFactory interface {
	New(context.Context, NativeDatabaseConfig) (CassandraNativeClient, error)
}

type CassandraNativeProvider struct{ factory CassandraNativeClientFactory }

func NewCassandraNativeProvider(factory CassandraNativeClientFactory) *CassandraNativeProvider {
	return &CassandraNativeProvider{factory: factory}
}

func NewDefaultCassandraNativeProvider() *CassandraNativeProvider {
	return NewCassandraNativeProvider(cassandraGoCQLClientFactory{})
}

func (p *CassandraNativeProvider) Test(ctx context.Context, cfg NativeDatabaseConfig) error {
	client, err := p.factory.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer client.Close()
	return client.Ping(ctx)
}

func (p *CassandraNativeProvider) Connect(ctx context.Context, cfg NativeDatabaseConfig) (NativeDatabaseClient, error) {
	client, err := p.factory.New(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx); err != nil {
		_ = client.Close()
		return nil, err
	}
	return &cassandraNativeSession{client: client}, nil
}

type cassandraNativeSession struct{ client CassandraNativeClient }

func (s *cassandraNativeSession) ListPrimaryResources(ctx context.Context) ([]NativeResource, error) {
	keyspaces, err := s.client.ListKeyspaces(ctx)
	if err != nil {
		return nil, err
	}
	sort.Strings(keyspaces)
	resources := make([]NativeResource, 0, len(keyspaces))
	for _, keyspace := range keyspaces {
		resources = append(resources, NativeResource{Kind: NativeResourceKindDatabase, Name: keyspace})
	}
	return resources, nil
}

func (s *cassandraNativeSession) ListSecondaryResources(ctx context.Context, keyspace string) ([]NativeResource, error) {
	tables, err := s.client.ListTables(ctx, keyspace)
	if err != nil {
		return nil, err
	}
	sort.Strings(tables)
	resources := make([]NativeResource, 0, len(tables))
	for _, table := range tables {
		resources = append(resources, NativeResource{Kind: NativeResourceKindTable, Name: table})
	}
	return resources, nil
}

func (s *cassandraNativeSession) Close() error { return s.client.Close() }

type cassandraGoCQLClientFactory struct{}

func (cassandraGoCQLClientFactory) New(ctx context.Context, cfg NativeDatabaseConfig) (CassandraNativeClient, error) {
	hosts := strings.FieldsFunc(strings.TrimSpace(cfg.Host), func(r rune) bool { return r == ',' || r == ';' })
	if len(hosts) == 0 {
		hosts = []string{"127.0.0.1"}
	}
	cluster := gocql.NewCluster(hosts...)
	if cfg.Port != 0 {
		cluster.Port = cfg.Port
	}
	if cfg.Timeout > 0 {
		cluster.Timeout = cfg.Timeout
		cluster.ConnectTimeout = cfg.Timeout
	}
	if strings.TrimSpace(cfg.User) != "" {
		cluster.Authenticator = gocql.PasswordAuthenticator{Username: cfg.User, Password: cfg.Password}
	}
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining > 0 && (cfg.Timeout == 0 || remaining < cluster.ConnectTimeout) {
			cluster.ConnectTimeout = remaining
		}
	}
	session, err := cluster.CreateSession()
	if err != nil {
		return nil, err
	}
	return &cassandraGoCQLClient{session: session}, nil
}

type cassandraGoCQLClient struct{ session *gocql.Session }

func (c *cassandraGoCQLClient) Ping(ctx context.Context) error {
	return c.session.Query("SELECT release_version FROM system.local").WithContext(ctx).Exec()
}

func (c *cassandraGoCQLClient) ListKeyspaces(ctx context.Context) ([]string, error) {
	iter := c.session.Query("SELECT keyspace_name FROM system_schema.keyspaces").WithContext(ctx).Iter()
	keyspaces := make([]string, 0)
	var keyspace string
	for iter.Scan(&keyspace) {
		keyspaces = append(keyspaces, keyspace)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return keyspaces, nil
}

func (c *cassandraGoCQLClient) ListTables(ctx context.Context, keyspace string) ([]string, error) {
	iter := c.session.Query("SELECT table_name FROM system_schema.tables WHERE keyspace_name = ?", keyspace).WithContext(ctx).Iter()
	tables := make([]string, 0)
	var table string
	for iter.Scan(&table) {
		tables = append(tables, table)
	}
	if err := iter.Close(); err != nil {
		return nil, err
	}
	return tables, nil
}

func (c *cassandraGoCQLClient) Close() error {
	c.session.Close()
	return nil
}
