package service

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestNativeDatabaseServiceConnectsAndBrowsesResources(t *testing.T) {
	client := &fakeNativeDatabaseClient{
		primary: []NativeResource{{Kind: NativeResourceKindDatabase, Name: "0"}},
		secondary: map[string][]NativeResource{
			"0": {{Kind: NativeResourceKindKey, Name: "session:1"}},
		},
	}
	provider := &fakeNativeDatabaseProvider{client: client}
	service := NewNativeDatabaseService(map[NativeDatabaseType]NativeDatabaseProvider{
		NativeDatabaseTypeRedis: provider,
	})

	err := service.Connect(context.Background(), "redis-session", NativeDatabaseConfig{
		Type:    NativeDatabaseTypeRedis,
		Host:    "127.0.0.1",
		Port:    6379,
		Timeout: time.Second,
	})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	if provider.connectCalls != 1 {
		t.Fatalf("connect calls = %d, want 1", provider.connectCalls)
	}

	primary, err := service.ListPrimaryResources(context.Background(), "redis-session")
	if err != nil {
		t.Fatalf("list primary resources: %v", err)
	}
	if len(primary) != 1 || primary[0].Name != "0" {
		t.Fatalf("unexpected primary resources: %+v", primary)
	}

	secondary, err := service.ListSecondaryResources(context.Background(), "redis-session", "0")
	if err != nil {
		t.Fatalf("list secondary resources: %v", err)
	}
	if len(secondary) != 1 || secondary[0].Name != "session:1" {
		t.Fatalf("unexpected secondary resources: %+v", secondary)
	}

	if err := service.Close("redis-session"); err != nil {
		t.Fatalf("close: %v", err)
	}
	if !client.closed {
		t.Fatal("expected native client to be closed")
	}
}

func TestNativeDatabaseServiceCloseAllClosesEverySession(t *testing.T) {
	clientA := &fakeNativeDatabaseClient{}
	clientB := &fakeNativeDatabaseClient{}
	provider := &sequencedNativeProvider{clients: []*fakeNativeDatabaseClient{clientA, clientB}}
	service := NewNativeDatabaseService(map[NativeDatabaseType]NativeDatabaseProvider{
		NativeDatabaseTypeRedis: provider,
	})

	if err := service.Connect(context.Background(), "a", NativeDatabaseConfig{Type: NativeDatabaseTypeRedis}); err != nil {
		t.Fatal(err)
	}
	if err := service.Connect(context.Background(), "b", NativeDatabaseConfig{Type: NativeDatabaseTypeRedis}); err != nil {
		t.Fatal(err)
	}
	if got := service.ListSessions(); len(got) != 2 {
		t.Fatalf("expected 2 sessions, got %v", got)
	}
	if err := service.CloseAll(); err != nil {
		t.Fatalf("CloseAll: %v", err)
	}
	if got := service.ListSessions(); len(got) != 0 {
		t.Fatalf("expected no sessions, got %v", got)
	}
	if !clientA.closed || !clientB.closed {
		t.Fatal("expected all native clients closed")
	}
}

func TestNativeDatabaseServiceRejectsUnsupportedTypeAndUnknownSession(t *testing.T) {
	service := NewNativeDatabaseService(nil)
	if err := service.Connect(context.Background(), "unsupported", NativeDatabaseConfig{Type: "unknown"}); err == nil {
		t.Fatal("expected unsupported type error")
	}
	if _, err := service.ListPrimaryResources(context.Background(), "missing"); err == nil {
		t.Fatal("expected missing session error")
	}
}

func TestNativeDatabaseServiceTestConnectionDelegatesToProvider(t *testing.T) {
	provider := &fakeNativeDatabaseProvider{testErr: errors.New("unavailable")}
	service := NewNativeDatabaseService(map[NativeDatabaseType]NativeDatabaseProvider{
		NativeDatabaseTypeMongoDB: provider,
	})

	err := service.TestConnection(context.Background(), NativeDatabaseConfig{Type: NativeDatabaseTypeMongoDB})
	if !errors.Is(err, provider.testErr) {
		t.Fatalf("test connection error = %v, want %v", err, provider.testErr)
	}
	if provider.testCalls != 1 {
		t.Fatalf("test calls = %d, want 1", provider.testCalls)
	}
}

type fakeNativeDatabaseProvider struct {
	client       NativeDatabaseClient
	connectCalls int
	testCalls    int
	testErr      error
}

func (p *fakeNativeDatabaseProvider) Test(context.Context, NativeDatabaseConfig) error {
	p.testCalls++
	return p.testErr
}

func (p *fakeNativeDatabaseProvider) Connect(context.Context, NativeDatabaseConfig) (NativeDatabaseClient, error) {
	p.connectCalls++
	return p.client, nil
}

type fakeNativeDatabaseClient struct {
	primary   []NativeResource
	secondary map[string][]NativeResource
	closed    bool
}

func (c *fakeNativeDatabaseClient) ListPrimaryResources(context.Context) ([]NativeResource, error) {
	return c.primary, nil
}

func (c *fakeNativeDatabaseClient) ListSecondaryResources(_ context.Context, parent string) ([]NativeResource, error) {
	return c.secondary[parent], nil
}

func (*fakeNativeDatabaseClient) DescribeResource(context.Context, string, string) (NativeResourceDetails, error) {
	return NativeResourceDetails{}, nil
}

func (c *fakeNativeDatabaseClient) Close() error {
	c.closed = true
	return nil
}

type sequencedNativeProvider struct {
	clients []*fakeNativeDatabaseClient
	next    int
}

func (*sequencedNativeProvider) Test(context.Context, NativeDatabaseConfig) error { return nil }

func (p *sequencedNativeProvider) Connect(context.Context, NativeDatabaseConfig) (NativeDatabaseClient, error) {
	if p.next >= len(p.clients) {
		return nil, errors.New("no more fake clients")
	}
	client := p.clients[p.next]
	p.next++
	return client, nil
}
