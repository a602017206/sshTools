package service

import (
	"context"
	"errors"
	"testing"
)

func TestInfluxDBNativeProviderTestsConnectionAndListsBuckets(t *testing.T) {
	client := &fakeInfluxDBNativeClient{buckets: []string{"metrics", "telemetry"}}
	provider := NewInfluxDBNativeProvider(&fakeInfluxDBNativeClientFactory{client: client})
	connected, err := provider.Connect(context.Background(), NativeDatabaseConfig{Type: NativeDatabaseTypeInfluxDB})
	if err != nil {
		t.Fatalf("connect InfluxDB: %v", err)
	}
	resources, err := connected.ListPrimaryResources(context.Background())
	if err != nil {
		t.Fatalf("list InfluxDB buckets: %v", err)
	}
	if got, want := resourceNames(resources), []string{"metrics", "telemetry"}; !equalStrings(got, want) {
		t.Fatalf("buckets = %v, want %v", got, want)
	}
	if err := connected.Close(); err != nil {
		t.Fatalf("close InfluxDB: %v", err)
	}
	if !client.closed {
		t.Fatal("expected InfluxDB client to close")
	}
}

func TestInfluxDBNativeProviderPropagatesHealthFailure(t *testing.T) {
	client := &fakeInfluxDBNativeClient{pingErr: errors.New("invalid token")}
	provider := NewInfluxDBNativeProvider(&fakeInfluxDBNativeClientFactory{client: client})
	err := provider.Test(context.Background(), NativeDatabaseConfig{Type: NativeDatabaseTypeInfluxDB})
	if !errors.Is(err, client.pingErr) {
		t.Fatalf("test error = %v", err)
	}
}

type fakeInfluxDBNativeClient struct {
	buckets []string
	pingErr error
	closed  bool
}

func (c *fakeInfluxDBNativeClient) Ping(context.Context) error { return c.pingErr }
func (c *fakeInfluxDBNativeClient) ListBuckets(context.Context) ([]string, error) {
	return c.buckets, nil
}
func (c *fakeInfluxDBNativeClient) Close() error { c.closed = true; return nil }

type fakeInfluxDBNativeClientFactory struct{ client InfluxDBNativeClient }

func (f *fakeInfluxDBNativeClientFactory) New(context.Context, NativeDatabaseConfig) (InfluxDBNativeClient, error) {
	return f.client, nil
}
