package main

import (
	"context"
	"testing"

	"AHaSSHTools/internal/service"
)

func TestNativeDatabaseAPIsConnectBrowseAndCloseWithoutJDBC(t *testing.T) {
	provider := &appNativeDatabaseProvider{client: &appNativeDatabaseClient{
		primary: []service.NativeResource{{Kind: service.NativeResourceKindDatabase, Name: "0"}},
		secondary: map[string][]service.NativeResource{
			"0": {{Kind: service.NativeResourceKindKey, Name: "cache:session"}},
		},
	}}
	app := &App{nativeDatabaseService: service.NewNativeDatabaseService(map[service.NativeDatabaseType]service.NativeDatabaseProvider{
		service.NativeDatabaseTypeRedis: provider,
	})}

	if err := app.TestNativeDatabaseConnection("127.0.0.1", 6379, "", "", "redis", "0"); err != nil {
		t.Fatalf("test native database: %v", err)
	}
	if err := app.ConnectNativeDatabase("redis-session", "127.0.0.1", 6379, "", "", "redis", "0"); err != nil {
		t.Fatalf("connect native database: %v", err)
	}
	primary, err := app.ListNativeDatabaseResources("redis-session")
	if err != nil || len(primary) != 1 || primary[0].Name != "0" {
		t.Fatalf("primary resources = %+v, %v", primary, err)
	}
	secondary, err := app.ListNativeDatabaseChildResources("redis-session", "0")
	if err != nil || len(secondary) != 1 || secondary[0].Name != "cache:session" {
		t.Fatalf("secondary resources = %+v, %v", secondary, err)
	}
	if err := app.CloseNativeDatabase("redis-session"); err != nil {
		t.Fatalf("close native database: %v", err)
	}
}

func TestNativeDatabaseRequestClearsLegacyRedisUsername(t *testing.T) {
	app := &App{nativeDatabaseService: service.NewNativeDatabaseService(nil)}

	_, cfg, _, cancel, err := app.nativeDatabaseRequest(
		"192.168.195.185", 6379, "Root", "secret", "redis", "0",
	)
	if err != nil {
		t.Fatalf("create Redis request: %v", err)
	}
	defer cancel()
	if cfg.User != "" {
		t.Fatalf("Redis username = %q, want empty for password-only authentication", cfg.User)
	}
}

type appNativeDatabaseProvider struct {
	client service.NativeDatabaseClient
}

func (p *appNativeDatabaseProvider) Test(context.Context, service.NativeDatabaseConfig) error {
	return nil
}

func (p *appNativeDatabaseProvider) Connect(context.Context, service.NativeDatabaseConfig) (service.NativeDatabaseClient, error) {
	return p.client, nil
}

type appNativeDatabaseClient struct {
	primary   []service.NativeResource
	secondary map[string][]service.NativeResource
}

func (c *appNativeDatabaseClient) ListPrimaryResources(context.Context) ([]service.NativeResource, error) {
	return c.primary, nil
}

func (c *appNativeDatabaseClient) ListSecondaryResources(_ context.Context, parent string) ([]service.NativeResource, error) {
	return c.secondary[parent], nil
}

func (*appNativeDatabaseClient) DescribeResource(context.Context, string, string) (service.NativeResourceDetails, error) {
	return service.NativeResourceDetails{}, nil
}

func (*appNativeDatabaseClient) Close() error { return nil }
