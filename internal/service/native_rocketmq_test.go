package service

import (
	"context"
	"errors"
	"testing"
)

func TestRocketMQNativeProviderListsTopics(t *testing.T) {
	client := &fakeRocketMQNativeClient{topics: []string{"orders", "events"}}
	provider := NewRocketMQNativeProvider(&fakeRocketMQNativeClientFactory{client: client})
	connected, err := provider.Connect(context.Background(), NativeDatabaseConfig{Type: NativeDatabaseTypeRocketMQ})
	if err != nil {
		t.Fatalf("connect RocketMQ: %v", err)
	}
	resources, err := connected.ListPrimaryResources(context.Background())
	if err != nil {
		t.Fatalf("list RocketMQ topics: %v", err)
	}
	if len(resources) != 2 || resources[0].Name != "events" {
		t.Fatalf("unexpected topics: %#v", resources)
	}
	details, err := connected.DescribeResource(context.Background(), "", "orders")
	if err != nil {
		t.Fatalf("describe topic: %v", err)
	}
	if details.Name != "orders" {
		t.Fatalf("unexpected details: %#v", details)
	}
	if err := connected.Close(); err != nil {
		t.Fatalf("close RocketMQ: %v", err)
	}
	if !client.closed {
		t.Fatal("expected RocketMQ client to close")
	}
}

func TestRocketMQNativeProviderPropagatesConnectivityFailure(t *testing.T) {
	client := &fakeRocketMQNativeClient{pingErr: errors.New("nameserver unavailable")}
	provider := NewRocketMQNativeProvider(&fakeRocketMQNativeClientFactory{client: client})
	err := provider.Test(context.Background(), NativeDatabaseConfig{Type: NativeDatabaseTypeRocketMQ})
	if err == nil {
		t.Fatal("expected connectivity failure")
	}
}

type fakeRocketMQNativeClient struct {
	topics  []string
	pingErr error
	closed  bool
}

func (c *fakeRocketMQNativeClient) Ping(context.Context) error { return c.pingErr }
func (c *fakeRocketMQNativeClient) ListTopics(context.Context) ([]string, error) {
	return append([]string{}, c.topics...), nil
}
func (c *fakeRocketMQNativeClient) DescribeTopic(_ context.Context, name string) (NativeResourceDetails, error) {
	return NativeResourceDetails{Kind: NativeResourceKindCollection, Name: name, Summary: "ok", Content: `{"topic":"` + name + `"}`}, nil
}
func (c *fakeRocketMQNativeClient) Close() error { c.closed = true; return nil }

type fakeRocketMQNativeClientFactory struct{ client RocketMQNativeClient }

func (f *fakeRocketMQNativeClientFactory) New(context.Context, NativeDatabaseConfig) (RocketMQNativeClient, error) {
	return f.client, nil
}
