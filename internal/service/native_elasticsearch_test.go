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

func TestElasticsearchNativeSessionDescribesIndexWithDocumentPreview(t *testing.T) {
	client := &fakeElasticsearchNativeClient{indexDetails: NativeResourceDetails{
		Kind: NativeResourceKindIndex, Name: "products", Summary: "12 文档 · 8 KiB",
		Content: `{"documents":[{"_id":"p-1","_source":{"name":"Keyboard"}}]}`,
	}}
	provider := NewElasticsearchNativeProvider(&fakeElasticsearchNativeClientFactory{client: client})
	connected, err := provider.Connect(context.Background(), NativeDatabaseConfig{Type: NativeDatabaseTypeElasticsearch})
	if err != nil {
		t.Fatalf("connect Elasticsearch: %v", err)
	}
	details, err := connected.DescribeResource(context.Background(), "", "products")
	if err != nil {
		t.Fatalf("describe Elasticsearch index: %v", err)
	}
	if details.Name != "products" || details.Content == "" {
		t.Fatalf("details = %#v", details)
	}
}

type fakeElasticsearchNativeClient struct {
	pingErr      error
	indices      []string
	indexDetails NativeResourceDetails
	closed       bool
}

func (c *fakeElasticsearchNativeClient) Ping(context.Context) error { return c.pingErr }

func (c *fakeElasticsearchNativeClient) ListIndices(context.Context) ([]string, error) {
	return c.indices, nil
}
func (c *fakeElasticsearchNativeClient) DescribeIndex(context.Context, string) (NativeResourceDetails, error) {
	return c.indexDetails, nil
}

func (c *fakeElasticsearchNativeClient) SearchIndex(context.Context, string, string) (NativeQueryResult, error) {
	return NativeQueryResult{Summary: "1 hit", Content: `{"hits":[]}`}, nil
}

func (c *fakeElasticsearchNativeClient) IndexDocument(context.Context, string, string) (NativeMutationResult, error) {
	return NativeMutationResult{Summary: "indexed"}, nil
}

func (c *fakeElasticsearchNativeClient) UpdateDocument(context.Context, string, string) (NativeMutationResult, error) {
	return NativeMutationResult{Summary: "updated"}, nil
}

func (c *fakeElasticsearchNativeClient) DeleteDocument(context.Context, string, string) (NativeMutationResult, error) {
	return NativeMutationResult{Summary: "deleted"}, nil
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
