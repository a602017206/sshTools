package service

import (
	"context"
	"testing"
)

func TestNativeDatabaseRegistryRegistersAndFindsProvider(t *testing.T) {
	registry := NewNativeDatabaseRegistry()
	provider := registryTestProvider{}
	if err := registry.Register(NativeDatabaseTypeRedis, provider); err != nil {
		t.Fatalf("register Redis provider: %v", err)
	}
	got, err := registry.Provider(NativeDatabaseTypeRedis)
	if err != nil {
		t.Fatalf("find Redis provider: %v", err)
	}
	if got != provider {
		t.Fatal("registry returned a different provider")
	}
}

func TestNativeDatabaseRegistryRejectsInvalidAndDuplicateProviderIDs(t *testing.T) {
	registry := NewNativeDatabaseRegistry()
	provider := registryTestProvider{}
	if err := registry.Register("", provider); err == nil {
		t.Fatal("expected empty provider ID rejection")
	}
	if err := registry.Register(NativeDatabaseTypeRedis, provider); err != nil {
		t.Fatal(err)
	}
	if err := registry.Register(NativeDatabaseTypeRedis, provider); err == nil {
		t.Fatal("expected duplicate provider ID rejection")
	}
	if _, err := registry.Provider(NativeDatabaseType("missing")); err == nil {
		t.Fatal("expected unknown provider error")
	}
}

type registryTestProvider struct{}

func (registryTestProvider) Test(context.Context, NativeDatabaseConfig) error { return nil }

func (registryTestProvider) Connect(context.Context, NativeDatabaseConfig) (NativeDatabaseClient, error) {
	return &fakeNativeDatabaseClient{}, nil
}
