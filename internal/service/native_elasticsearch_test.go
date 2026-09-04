package service

import (
	"context"
	"errors"
	"strings"
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

func TestElasticsearchNativeSessionDescribesIndexMetadata(t *testing.T) {
	client := &fakeElasticsearchNativeClient{indexDetails: NativeResourceDetails{
		Kind: NativeResourceKindIndex, Name: "products", Summary: "green · 12 文档 · 8kb",
		Content: `{"stats":{"docsCount":"12","storeSize":"8kb","health":"green"},"mapping":{"mappings":{"properties":{"name":{"type":"text"}}}}}`,
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
	if details.Name != "products" || !strings.Contains(details.Content, `"mapping"`) || strings.Contains(details.Content, `"documents"`) {
		t.Fatalf("details = %#v", details)
	}
}

func TestElasticsearchNativeSessionDescribesClusterOverview(t *testing.T) {
	client := &fakeElasticsearchNativeClient{clusterDetails: NativeResourceDetails{
		Name: "demo-cluster", Summary: "demo-cluster · green · 3 节点",
		Content: `{"clusterName":"demo-cluster","nodeCount":3,"health":"green","version":"8.15.0"}`,
	}}
	provider := NewElasticsearchNativeProvider(&fakeElasticsearchNativeClientFactory{client: client})
	connected, err := provider.Connect(context.Background(), NativeDatabaseConfig{Type: NativeDatabaseTypeElasticsearch})
	if err != nil {
		t.Fatalf("connect Elasticsearch: %v", err)
	}
	inspector, ok := connected.(NativeSessionInspector)
	if !ok {
		t.Fatal("expected Elasticsearch session to expose session overview")
	}
	details, err := inspector.DescribeSession(context.Background())
	if err != nil {
		t.Fatalf("describe cluster: %v", err)
	}
	if details.Name != "demo-cluster" || !strings.Contains(details.Content, `"nodeCount":3`) {
		t.Fatalf("details = %#v", details)
	}
}

type fakeElasticsearchNativeClient struct {
	pingErr         error
	indices         []string
	indexDetails    NativeResourceDetails
	clusterDetails  NativeResourceDetails
	closed          bool
}

func (c *fakeElasticsearchNativeClient) Ping(context.Context) error { return c.pingErr }

func (c *fakeElasticsearchNativeClient) ListIndices(context.Context) ([]string, error) {
	return c.indices, nil
}
func (c *fakeElasticsearchNativeClient) DescribeIndex(context.Context, string) (NativeResourceDetails, error) {
	return c.indexDetails, nil
}

func (c *fakeElasticsearchNativeClient) DescribeCluster(context.Context) (NativeResourceDetails, error) {
	return c.clusterDetails, nil
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

func (c *fakeElasticsearchNativeClient) CreateIndex(context.Context, string, string) (NativeMutationResult, error) {
	return NativeMutationResult{Summary: "created"}, nil
}

func (c *fakeElasticsearchNativeClient) DeleteIndex(context.Context, string) (NativeMutationResult, error) {
	return NativeMutationResult{Summary: "deleted"}, nil
}

func (c *fakeElasticsearchNativeClient) RefreshIndex(context.Context, string) (NativeMutationResult, error) {
	return NativeMutationResult{Summary: "refreshed"}, nil
}

func (c *fakeElasticsearchNativeClient) PerformRequest(context.Context, string, string, string) (NativeQueryResult, error) {
	return NativeQueryResult{Summary: "ok", Content: `{"status":200}`}, nil
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
