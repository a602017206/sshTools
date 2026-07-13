package service

import (
	"context"
	"errors"
	"testing"
)

func TestElasticsearchNativeProviderTestsConnectionAndListsIndices(t *testing.T) {
	client := &fakeElasticsearchNativeClient{indices: []string{".system", "logs-2026", "products"}}
	provider := NewElasticsearchNativeProvider(&fakeElasticsearchNativeClientFactory{client: client})
	cfg := NativeDatabaseConfig{Type: NativeDatabaseTypeElasticsearch, Host: "search.local", Port: 9200}

	if err := provider.Test(context.Background(), cfg); err != nil {
		t.Fatalf("test Elasticsearch connection: %v", err)
	}
	connected, err := provider.Connect(context.Background(), cfg)
	if err != nil {
		t.Fatalf("connect Elasticsearch: %v", err)
	}
	indices, err := connected.ListPrimaryResources(context.Background())
	if err != nil {
		t.Fatalf("list Elasticsearch indices: %v", err)
	}
	if got, want := resourceNames(indices), []string{".system", "logs-2026", "products"}; !equalStrings(got, want) {
		t.Fatalf("indices = %v, want %v", got, want)
	}
	secondary, err := connected.ListSecondaryResources(context.Background(), "products")
	if err != nil {
		t.Fatalf("list Elasticsearch secondary resources: %v", err)
	}
	if len(secondary) != 0 {
		t.Fatalf("secondary resources = %v, want empty", secondary)
	}
	if err := connected.Close(); err != nil {
		t.Fatalf("close Elasticsearch: %v", err)
	}
	if !client.closed {
		t.Fatal("expected Elasticsearch client to close")
	}
}

func TestElasticsearchNativeProviderPropagatesHealthFailure(t *testing.T) {
	client := &fakeElasticsearchNativeClient{pingErr: errors.New("unauthorized")}
	provider := NewElasticsearchNativeProvider(&fakeElasticsearchNativeClientFactory{client: client})
	err := provider.Test(context.Background(), NativeDatabaseConfig{Type: NativeDatabaseTypeElasticsearch})
	if !errors.Is(err, client.pingErr) {
		t.Fatalf("test error = %v", err)
	}
	if !client.closed {
		t.Fatal("expected failed test client to close")
	}
}

type fakeElasticsearchNativeClient struct {
	pingErr error
	indices []string
	closed  bool
}

func (c *fakeElasticsearchNativeClient) Ping(context.Context) error { return c.pingErr }

func (c *fakeElasticsearchNativeClient) ListIndices(context.Context) ([]string, error) {
	return c.indices, nil
}

func (c *fakeElasticsearchNativeClient) Close() error {
	c.closed = true
	return nil
}

type fakeElasticsearchNativeClientFactory struct {
	client ElasticsearchNativeClient
	err    error
}

func (f *fakeElasticsearchNativeClientFactory) New(NativeDatabaseConfig) (ElasticsearchNativeClient, error) {
	return f.client, f.err
}
