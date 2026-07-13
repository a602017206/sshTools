package service

import (
	"context"
	"errors"
	"testing"
)

func TestKafkaNativeProviderTestsConnectionAndListsTopics(t *testing.T) {
	client := &fakeKafkaNativeClient{topics: []string{"events", "orders"}}
	provider := NewKafkaNativeProvider(&fakeKafkaNativeClientFactory{client: client})
	connected, err := provider.Connect(context.Background(), NativeDatabaseConfig{Type: NativeDatabaseTypeKafka})
	if err != nil {
		t.Fatalf("connect Kafka: %v", err)
	}
	resources, err := connected.ListPrimaryResources(context.Background())
	if err != nil {
		t.Fatalf("list Kafka topics: %v", err)
	}
	if got, want := resourceNames(resources), []string{"events", "orders"}; !equalStrings(got, want) {
		t.Fatalf("topics = %v, want %v", got, want)
	}
	if err := connected.Close(); err != nil {
		t.Fatalf("close Kafka: %v", err)
	}
	if !client.closed {
		t.Fatal("expected Kafka client to close")
	}
}
func TestKafkaNativeProviderPropagatesConnectivityFailure(t *testing.T) {
	client := &fakeKafkaNativeClient{pingErr: errors.New("broker unavailable")}
	provider := NewKafkaNativeProvider(&fakeKafkaNativeClientFactory{client: client})
	err := provider.Test(context.Background(), NativeDatabaseConfig{Type: NativeDatabaseTypeKafka})
	if !errors.Is(err, client.pingErr) {
		t.Fatalf("test error = %v", err)
	}
}

type fakeKafkaNativeClient struct {
	topics  []string
	pingErr error
	closed  bool
}

func (c *fakeKafkaNativeClient) Ping(context.Context) error                   { return c.pingErr }
func (c *fakeKafkaNativeClient) ListTopics(context.Context) ([]string, error) { return c.topics, nil }
func (c *fakeKafkaNativeClient) Close() error                                 { c.closed = true; return nil }

type fakeKafkaNativeClientFactory struct{ client KafkaNativeClient }

func (f *fakeKafkaNativeClientFactory) New(context.Context, NativeDatabaseConfig) (KafkaNativeClient, error) {
	return f.client, nil
}
