package service

import (
	"context"
	"errors"
	"testing"
)

func TestCassandraNativeProviderTestsConnectionAndBrowsesKeyspaces(t *testing.T) {
	client := &fakeCassandraNativeClient{
		keyspaces: map[string][]string{
			"inventory": {"products", "warehouses"},
			"system":    {"local"},
		},
	}
	provider := NewCassandraNativeProvider(&fakeCassandraNativeClientFactory{client: client})
	connected, err := provider.Connect(context.Background(), NativeDatabaseConfig{Type: NativeDatabaseTypeCassandra})
	if err != nil {
		t.Fatalf("connect Cassandra: %v", err)
	}

	keyspaces, err := connected.ListPrimaryResources(context.Background())
	if err != nil {
		t.Fatalf("list Cassandra keyspaces: %v", err)
	}
	if got, want := resourceNames(keyspaces), []string{"inventory", "system"}; !equalStrings(got, want) {
		t.Fatalf("keyspaces = %v, want %v", got, want)
	}
	tables, err := connected.ListSecondaryResources(context.Background(), "inventory")
	if err != nil {
		t.Fatalf("list Cassandra tables: %v", err)
	}
	if got, want := resourceNames(tables), []string{"products", "warehouses"}; !equalStrings(got, want) {
		t.Fatalf("tables = %v, want %v", got, want)
	}
	if err := connected.Close(); err != nil {
		t.Fatalf("close Cassandra: %v", err)
	}
	if !client.closed {
		t.Fatal("expected Cassandra client to close")
	}
}

func TestCassandraNativeProviderPropagatesConnectionFailure(t *testing.T) {
	client := &fakeCassandraNativeClient{connectErr: errors.New("authentication failed")}
	provider := NewCassandraNativeProvider(&fakeCassandraNativeClientFactory{client: client})
	err := provider.Test(context.Background(), NativeDatabaseConfig{Type: NativeDatabaseTypeCassandra})
	if !errors.Is(err, client.connectErr) {
		t.Fatalf("test error = %v", err)
	}
}

type fakeCassandraNativeClient struct {
	keyspaces  map[string][]string
	connectErr error
	closed     bool
}

func (c *fakeCassandraNativeClient) Ping(context.Context) error { return c.connectErr }
func (c *fakeCassandraNativeClient) ListKeyspaces(context.Context) ([]string, error) {
	keyspaces := make([]string, 0, len(c.keyspaces))
	for keyspace := range c.keyspaces {
		keyspaces = append(keyspaces, keyspace)
	}
	return keyspaces, nil
}
func (c *fakeCassandraNativeClient) ListTables(_ context.Context, keyspace string) ([]string, error) {
	return c.keyspaces[keyspace], nil
}
func (c *fakeCassandraNativeClient) Close() error { c.closed = true; return nil }

type fakeCassandraNativeClientFactory struct{ client CassandraNativeClient }

func (f *fakeCassandraNativeClientFactory) New(context.Context, NativeDatabaseConfig) (CassandraNativeClient, error) {
	return f.client, nil
}
