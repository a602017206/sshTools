package service

import (
	"context"
	"errors"
	"testing"
)

func TestRabbitMQNativeProviderListsQueues(t *testing.T) {
	client := &fakeRabbitMQNativeClient{queues: []string{"orders", "events"}}
	provider := NewRabbitMQNativeProvider(&fakeRabbitMQNativeClientFactory{client: client})
	connected, err := provider.Connect(context.Background(), NativeDatabaseConfig{Type: NativeDatabaseTypeRabbitMQ})
	if err != nil {
		t.Fatalf("connect RabbitMQ: %v", err)
	}
	resources, err := connected.ListPrimaryResources(context.Background())
	if err != nil {
		t.Fatalf("list RabbitMQ queues: %v", err)
	}
	if len(resources) != 2 || resources[0].Name != "events" {
		t.Fatalf("unexpected queues: %#v", resources)
	}
	details, err := connected.DescribeResource(context.Background(), "", "orders")
	if err != nil {
		t.Fatalf("describe queue: %v", err)
	}
	if details.Name != "orders" {
		t.Fatalf("unexpected details: %#v", details)
	}
	if err := connected.Close(); err != nil {
		t.Fatalf("close RabbitMQ: %v", err)
	}
	if !client.closed {
		t.Fatal("expected RabbitMQ client to close")
	}
}

func TestRabbitMQNativeProviderPropagatesConnectivityFailure(t *testing.T) {
	client := &fakeRabbitMQNativeClient{pingErr: errors.New("amqp unavailable")}
	provider := NewRabbitMQNativeProvider(&fakeRabbitMQNativeClientFactory{client: client})
	err := provider.Test(context.Background(), NativeDatabaseConfig{Type: NativeDatabaseTypeRabbitMQ})
	if err == nil {
		t.Fatal("expected connectivity failure")
	}
}

type fakeRabbitMQNativeClient struct {
	queues  []string
	pingErr error
	closed  bool
}

func (c *fakeRabbitMQNativeClient) Ping(context.Context) error { return c.pingErr }
func (c *fakeRabbitMQNativeClient) ListQueues(context.Context) ([]string, error) {
	return append([]string{}, c.queues...), nil
}
func (c *fakeRabbitMQNativeClient) DescribeQueue(_ context.Context, name string) (NativeResourceDetails, error) {
	return NativeResourceDetails{Kind: NativeResourceKindCollection, Name: name, Summary: "ok", Content: `{"name":"` + name + `"}`}, nil
}
func (c *fakeRabbitMQNativeClient) Close() error { c.closed = true; return nil }

type fakeRabbitMQNativeClientFactory struct{ client RabbitMQNativeClient }

func (f *fakeRabbitMQNativeClientFactory) New(context.Context, NativeDatabaseConfig) (RabbitMQNativeClient, error) {
	return f.client, nil
}
