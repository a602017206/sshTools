package service

import (
	"context"
	"errors"
	"testing"
)

func TestMongoNativeProviderTestsConnectionAndBrowsesCollections(t *testing.T) {
	client := &fakeMongoNativeClient{
		databases: []string{"admin", "inventory"},
		collections: map[string][]string{
			"inventory": {"orders", "products"},
		},
	}
	provider := NewMongoNativeProvider(&fakeMongoNativeClientFactory{client: client})
	cfg := NativeDatabaseConfig{Type: NativeDatabaseTypeMongoDB, Host: "mongo.local", Port: 27017}

	if err := provider.Test(context.Background(), cfg); err != nil {
		t.Fatalf("test MongoDB connection: %v", err)
	}
	connected, err := provider.Connect(context.Background(), cfg)
	if err != nil {
		t.Fatalf("connect MongoDB: %v", err)
	}
	databases, err := connected.ListPrimaryResources(context.Background())
	if err != nil {
		t.Fatalf("list MongoDB databases: %v", err)
	}
	if got, want := resourceNames(databases), []string{"admin", "inventory"}; !equalStrings(got, want) {
		t.Fatalf("databases = %v, want %v", got, want)
	}
	collections, err := connected.ListSecondaryResources(context.Background(), "inventory")
	if err != nil {
		t.Fatalf("list MongoDB collections: %v", err)
	}
	if got, want := resourceNames(collections), []string{"orders", "products"}; !equalStrings(got, want) {
		t.Fatalf("collections = %v, want %v", got, want)
	}
	if err := connected.Close(); err != nil {
		t.Fatalf("close MongoDB: %v", err)
	}
	if !client.closed {
		t.Fatal("expected MongoDB client to close")
	}
}

func TestMongoNativeProviderPropagatesPingFailure(t *testing.T) {
	client := &fakeMongoNativeClient{pingErr: errors.New("authentication failed")}
	provider := NewMongoNativeProvider(&fakeMongoNativeClientFactory{client: client})
	err := provider.Test(context.Background(), NativeDatabaseConfig{Type: NativeDatabaseTypeMongoDB})
	if !errors.Is(err, client.pingErr) {
		t.Fatalf("test error = %v", err)
	}
	if !client.closed {
		t.Fatal("expected failed test client to close")
	}
}

type fakeMongoNativeClient struct {
	pingErr     error
	databases   []string
	collections map[string][]string
	closed      bool
}

func (c *fakeMongoNativeClient) Ping(context.Context) error { return c.pingErr }

func (c *fakeMongoNativeClient) ListDatabaseNames(context.Context) ([]string, error) {
	return c.databases, nil
}

func (c *fakeMongoNativeClient) ListCollectionNames(_ context.Context, database string) ([]string, error) {
	return c.collections[database], nil
}

func (c *fakeMongoNativeClient) Close(context.Context) error {
	c.closed = true
	return nil
}

type fakeMongoNativeClientFactory struct {
	client MongoNativeClient
	err    error
}

func (f *fakeMongoNativeClientFactory) New(context.Context, NativeDatabaseConfig) (MongoNativeClient, error) {
	return f.client, f.err
}
