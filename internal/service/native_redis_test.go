package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRedisNativeProviderTestsConnectionAndBrowsesKeys(t *testing.T) {
	primaryClient := &fakeRedisNativeClient{
		keyspace: map[int]int{0: 2, 3: 1},
	}
	keyClient := &fakeRedisNativeClient{
		pages: []redisScanPage{
			{keys: []string{"cache:a", "cache:b"}, cursor: 5},
			{keys: []string{"cache:c"}, cursor: 0},
		},
	}
	factory := &fakeRedisNativeClientFactory{clients: map[int]*fakeRedisNativeClient{
		0: primaryClient,
		3: keyClient,
	}}
	provider := NewRedisNativeProvider(factory)
	cfg := NativeDatabaseConfig{
		Type:     NativeDatabaseTypeRedis,
		Host:     "127.0.0.1",
		Port:     6379,
		Timeout:  time.Second,
		Database: "0",
	}

	if err := provider.Test(context.Background(), cfg); err != nil {
		t.Fatalf("test Redis connection: %v", err)
	}
	client, err := provider.Connect(context.Background(), cfg)
	if err != nil {
		t.Fatalf("connect Redis: %v", err)
	}
	primary, err := client.ListPrimaryResources(context.Background())
	if err != nil {
		t.Fatalf("list Redis logical databases: %v", err)
	}
	if got, want := resourceNames(primary), []string{"0", "3"}; !equalStrings(got, want) {
		t.Fatalf("logical databases = %v, want %v", got, want)
	}

	keys, err := client.ListSecondaryResources(context.Background(), "3")
	if err != nil {
		t.Fatalf("scan Redis keys: %v", err)
	}
	if got, want := resourceNames(keys), []string{"cache:a", "cache:b", "cache:c"}; !equalStrings(got, want) {
		t.Fatalf("keys = %v, want %v", got, want)
	}
	if keyClient.scanCalls != 2 {
		t.Fatalf("scan calls = %d, want 2", keyClient.scanCalls)
	}
	if err := client.Close(); err != nil {
		t.Fatalf("close Redis client: %v", err)
	}
	if !primaryClient.closed {
		t.Fatal("expected primary Redis client to close")
	}
}

func TestRedisNativeProviderRejectsInvalidLogicalDatabase(t *testing.T) {
	provider := NewRedisNativeProvider(&fakeRedisNativeClientFactory{})
	if _, err := provider.Connect(context.Background(), NativeDatabaseConfig{
		Type: NativeDatabaseTypeRedis, Database: "not-a-number",
	}); err == nil {
		t.Fatal("expected invalid Redis logical database error")
	}
}

func TestRedisNativeProviderWrapsConnectionFailure(t *testing.T) {
	factory := &fakeRedisNativeClientFactory{clients: map[int]*fakeRedisNativeClient{
		0: {pingErr: errors.New("connection refused")},
	}}
	provider := NewRedisNativeProvider(factory)
	err := provider.Test(context.Background(), NativeDatabaseConfig{Type: NativeDatabaseTypeRedis})
	if !errors.Is(err, factory.clients[0].pingErr) {
		t.Fatalf("test error = %v", err)
	}
}

func TestRedisNativeSessionDescribesKeyWithBoundedPreview(t *testing.T) {
	client := &fakeRedisNativeClient{keyDetails: NativeResourceDetails{
		Kind:    NativeResourceKindKey,
		Name:    "cache:profile:42",
		Summary: "string · 58 秒后过期",
		Content: `{"type":"string","ttlSeconds":58,"value":"alice"}`,
	}}
	provider := NewRedisNativeProvider(&fakeRedisNativeClientFactory{clients: map[int]*fakeRedisNativeClient{0: client}})
	connected, err := provider.Connect(context.Background(), NativeDatabaseConfig{Type: NativeDatabaseTypeRedis})
	if err != nil {
		t.Fatalf("connect Redis: %v", err)
	}
	details, err := connected.DescribeResource(context.Background(), "0", "cache:profile:42")
	if err != nil {
		t.Fatalf("describe Redis key: %v", err)
	}
	if details.Summary != "string · 58 秒后过期" || details.Content == "" {
		t.Fatalf("details = %#v", details)
	}
}

type redisScanPage struct {
	keys   []string
	cursor uint64
	err    error
}

type fakeRedisNativeClient struct {
	pingErr    error
	keyspace   map[int]int
	pages      []redisScanPage
	keyDetails NativeResourceDetails
	scanCalls  int
	closed     bool
}

func (c *fakeRedisNativeClient) DescribeKey(context.Context, string) (NativeResourceDetails, error) {
	return c.keyDetails, nil
}

func (c *fakeRedisNativeClient) SetKey(context.Context, string, string) (NativeMutationResult, error) {
	return NativeMutationResult{Summary: "saved"}, nil
}

func (c *fakeRedisNativeClient) SaveKeyValue(context.Context, string, string) (NativeMutationResult, error) {
	return NativeMutationResult{Summary: "saved"}, nil
}

func (c *fakeRedisNativeClient) DeleteKey(context.Context, string) (NativeMutationResult, error) {
	return NativeMutationResult{Summary: "deleted"}, nil
}

func (c *fakeRedisNativeClient) Ping(context.Context) error { return c.pingErr }

func (c *fakeRedisNativeClient) Keyspace(context.Context) (map[int]int, error) {
	return c.keyspace, nil
}

func (c *fakeRedisNativeClient) Scan(_ context.Context, _ uint64, _ int, _ int64) ([]string, uint64, error) {
	page := c.pages[c.scanCalls]
	c.scanCalls++
	return page.keys, page.cursor, page.err
}

func (c *fakeRedisNativeClient) Close() error {
	c.closed = true
	return nil
}

type fakeRedisNativeClientFactory struct {
	clients map[int]*fakeRedisNativeClient
}

func (f *fakeRedisNativeClientFactory) New(_ NativeDatabaseConfig, database int) (RedisNativeClient, error) {
	if f.clients == nil || f.clients[database] == nil {
		return nil, errors.New("no fake Redis client")
	}
	return f.clients[database], nil
}

func resourceNames(resources []NativeResource) []string {
	names := make([]string, 0, len(resources))
	for _, resource := range resources {
		names = append(names, resource.Name)
	}
	return names
}

func equalStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range got {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}
