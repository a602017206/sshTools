package service

import (
	"context"
	"errors"
	"testing"
)

func TestCouchbaseNativeProviderTestsConnectionAndBrowsesBuckets(t *testing.T) {
	client := &fakeCouchbaseNativeClient{collections: map[string][]string{
		"orders": {"_default._default", "inventory.products"},
		"users":  {"identity.accounts"},
	}}
	provider := NewCouchbaseNativeProvider(&fakeCouchbaseNativeClientFactory{client: client})
	connected, err := provider.Connect(context.Background(), NativeDatabaseConfig{Type: NativeDatabaseTypeCouchbase})
	if err != nil {
		t.Fatalf("connect Couchbase: %v", err)
	}
	buckets, err := connected.ListPrimaryResources(context.Background())
	if err != nil {
		t.Fatalf("list Couchbase buckets: %v", err)
	}
	if got, want := resourceNames(buckets), []string{"orders", "users"}; !equalStrings(got, want) {
		t.Fatalf("buckets = %v, want %v", got, want)
	}
	collections, err := connected.ListSecondaryResources(context.Background(), "orders")
	if err != nil {
		t.Fatalf("list Couchbase collections: %v", err)
	}
	if got, want := resourceNames(collections), []string{"_default._default", "inventory.products"}; !equalStrings(got, want) {
		t.Fatalf("collections = %v, want %v", got, want)
	}
	if err := connected.Close(); err != nil {
		t.Fatalf("close Couchbase: %v", err)
	}
	if !client.closed {
		t.Fatal("expected Couchbase client to close")
	}
}

func TestCouchbaseNativeProviderPropagatesHealthFailure(t *testing.T) {
	client := &fakeCouchbaseNativeClient{pingErr: errors.New("not authorized")}
	provider := NewCouchbaseNativeProvider(&fakeCouchbaseNativeClientFactory{client: client})
	err := provider.Test(context.Background(), NativeDatabaseConfig{Type: NativeDatabaseTypeCouchbase})
	if !errors.Is(err, client.pingErr) {
		t.Fatalf("test error = %v", err)
	}
}

type fakeCouchbaseNativeClient struct {
	collections map[string][]string
	pingErr     error
	closed      bool
}

func (c *fakeCouchbaseNativeClient) Ping(context.Context) error { return c.pingErr }
func (c *fakeCouchbaseNativeClient) ListBuckets(context.Context) ([]string, error) {
	buckets := make([]string, 0, len(c.collections))
	for bucket := range c.collections {
		buckets = append(buckets, bucket)
	}
	return buckets, nil
}
func (c *fakeCouchbaseNativeClient) ListCollections(_ context.Context, bucket string) ([]string, error) {
	return c.collections[bucket], nil
}
func (c *fakeCouchbaseNativeClient) Close() error { c.closed = true; return nil }

type fakeCouchbaseNativeClientFactory struct{ client CouchbaseNativeClient }

func (f *fakeCouchbaseNativeClientFactory) New(context.Context, NativeDatabaseConfig) (CouchbaseNativeClient, error) {
	return f.client, nil
}
