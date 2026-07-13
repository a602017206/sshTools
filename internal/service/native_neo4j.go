package service

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Neo4jNativeClient interface {
	Ping(context.Context) error
	ListDatabases(context.Context) ([]string, error)
	Close() error
}
type Neo4jNativeClientFactory interface {
	New(context.Context, NativeDatabaseConfig) (Neo4jNativeClient, error)
}
type Neo4jNativeProvider struct{ factory Neo4jNativeClientFactory }

func NewNeo4jNativeProvider(factory Neo4jNativeClientFactory) *Neo4jNativeProvider {
	return &Neo4jNativeProvider{factory: factory}
}
func NewDefaultNeo4jNativeProvider() *Neo4jNativeProvider {
	return NewNeo4jNativeProvider(neo4jGoClientFactory{})
}
func (p *Neo4jNativeProvider) Test(ctx context.Context, cfg NativeDatabaseConfig) error {
	client, err := p.factory.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer client.Close()
	return client.Ping(ctx)
}
func (p *Neo4jNativeProvider) Connect(ctx context.Context, cfg NativeDatabaseConfig) (NativeDatabaseClient, error) {
	client, err := p.factory.New(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx); err != nil {
		_ = client.Close()
		return nil, err
	}
	return &neo4jNativeSession{client: client}, nil
}

type neo4jNativeSession struct{ client Neo4jNativeClient }

func (s *neo4jNativeSession) ListPrimaryResources(ctx context.Context) ([]NativeResource, error) {
	databases, err := s.client.ListDatabases(ctx)
	if err != nil {
		return nil, err
	}
	sort.Strings(databases)
	resources := make([]NativeResource, 0, len(databases))
	for _, database := range databases {
		resources = append(resources, NativeResource{Kind: NativeResourceKindDatabase, Name: database})
	}
	return resources, nil
}
func (*neo4jNativeSession) ListSecondaryResources(context.Context, string) ([]NativeResource, error) {
	return []NativeResource{}, nil
}
func (s *neo4jNativeSession) Close() error { return s.client.Close() }

type neo4jGoClientFactory struct{}

func (neo4jGoClientFactory) New(_ context.Context, cfg NativeDatabaseConfig) (Neo4jNativeClient, error) {
	host := strings.TrimSpace(cfg.Host)
	if host == "" {
		host = "127.0.0.1"
	}
	port := cfg.Port
	if port == 0 {
		port = 7687
	}
	driver, err := neo4j.NewDriverWithContext("neo4j://"+host+":"+strconv.Itoa(port), neo4j.BasicAuth(cfg.User, cfg.Password, ""))
	if err != nil {
		return nil, err
	}
	return &neo4jGoClient{driver: driver}, nil
}

type neo4jGoClient struct{ driver neo4j.DriverWithContext }

func (c *neo4jGoClient) Ping(ctx context.Context) error { return c.driver.VerifyConnectivity(ctx) }
func (c *neo4jGoClient) ListDatabases(ctx context.Context) ([]string, error) {
	result, err := neo4j.ExecuteQuery(ctx, c.driver, "SHOW DATABASES YIELD name RETURN name", nil, neo4j.EagerResultTransformer, neo4j.ExecuteQueryWithDatabase("system"))
	if err != nil {
		return nil, err
	}
	databases := make([]string, 0, len(result.Records))
	for _, record := range result.Records {
		name, ok := record.Get("name")
		if !ok {
			continue
		}
		value, ok := name.(string)
		if !ok {
			return nil, fmt.Errorf("Neo4j 数据库名称格式无效")
		}
		databases = append(databases, value)
	}
	return databases, nil
}
func (c *neo4jGoClient) Close() error { return c.driver.Close(context.Background()) }
