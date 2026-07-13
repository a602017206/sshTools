package service

import (
	"context"
	"errors"
	"testing"
)

func TestMemcachedNativeProviderTestsConnectionAndListsStatistics(t *testing.T) {
	client := &fakeMemcachedNativeClient{stats: map[string]string{"curr_items": "4", "version": "1.6.0"}}
	provider := NewMemcachedNativeProvider(&fakeMemcachedNativeClientFactory{client: client})
	connected, err := provider.Connect(context.Background(), NativeDatabaseConfig{Type: NativeDatabaseTypeMemcached})
	if err != nil {
		t.Fatalf("connect Memcached: %v", err)
	}
	resources, err := connected.ListPrimaryResources(context.Background())
	if err != nil {
		t.Fatalf("list Memcached statistics: %v", err)
	}
	if got, want := resourceNames(resources), []string{"curr_items=4", "version=1.6.0"}; !equalStrings(got, want) {
		t.Fatalf("statistics = %v, want %v", got, want)
	}
	if err := connected.Close(); err != nil {
		t.Fatalf("close Memcached: %v", err)
	}
	if !client.closed {
		t.Fatal("expected Memcached client to close")
	}
}

func TestMemcachedNativeProviderPropagatesStatisticsFailure(t *testing.T) {
	client := &fakeMemcachedNativeClient{statsErr: errors.New("connection refused")}
	provider := NewMemcachedNativeProvider(&fakeMemcachedNativeClientFactory{client: client})
	err := provider.Test(context.Background(), NativeDatabaseConfig{Type: NativeDatabaseTypeMemcached})
	if !errors.Is(err, client.statsErr) {
		t.Fatalf("test error = %v", err)
	}
}

type fakeMemcachedNativeClient struct {
	stats    map[string]string
	statsErr error
	closed   bool
}

func (c *fakeMemcachedNativeClient) Stats(context.Context) (map[string]string, error) {
	return c.stats, c.statsErr
}
func (c *fakeMemcachedNativeClient) Close() error { c.closed = true; return nil }

type fakeMemcachedNativeClientFactory struct{ client MemcachedNativeClient }

func (f *fakeMemcachedNativeClientFactory) New(context.Context, NativeDatabaseConfig) (MemcachedNativeClient, error) {
	return f.client, nil
}
