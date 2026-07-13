package service

import (
	"context"
	"errors"
	"testing"
)

func TestNeo4jNativeProviderTestsConnectionAndListsDatabases(t *testing.T) {
	client := &fakeNeo4jNativeClient{databases: []string{"neo4j", "system"}}
	provider := NewNeo4jNativeProvider(&fakeNeo4jNativeClientFactory{client: client})
	connected, err := provider.Connect(context.Background(), NativeDatabaseConfig{Type: NativeDatabaseTypeNeo4j})
	if err != nil { t.Fatalf("connect Neo4j: %v", err) }
	resources, err := connected.ListPrimaryResources(context.Background())
	if err != nil { t.Fatalf("list Neo4j databases: %v", err) }
	if got, want := resourceNames(resources), []string{"neo4j", "system"}; !equalStrings(got, want) { t.Fatalf("databases = %v, want %v", got, want) }
	if err := connected.Close(); err != nil { t.Fatalf("close Neo4j: %v", err) }
	if !client.closed { t.Fatal("expected Neo4j client to close") }
}
func TestNeo4jNativeProviderPropagatesConnectivityFailure(t *testing.T) { client := &fakeNeo4jNativeClient{pingErr: errors.New("authentication failed")}; provider := NewNeo4jNativeProvider(&fakeNeo4jNativeClientFactory{client: client}); err := provider.Test(context.Background(), NativeDatabaseConfig{Type: NativeDatabaseTypeNeo4j}); if !errors.Is(err, client.pingErr) { t.Fatalf("test error = %v", err) } }
type fakeNeo4jNativeClient struct { databases []string; pingErr error; closed bool }
func (c *fakeNeo4jNativeClient) Ping(context.Context) error { return c.pingErr }
func (c *fakeNeo4jNativeClient) ListDatabases(context.Context) ([]string, error) { return c.databases, nil }
func (c *fakeNeo4jNativeClient) Close() error { c.closed = true; return nil }
type fakeNeo4jNativeClientFactory struct{ client Neo4jNativeClient }
func (f *fakeNeo4jNativeClientFactory) New(context.Context, NativeDatabaseConfig) (Neo4jNativeClient, error) { return f.client, nil }
