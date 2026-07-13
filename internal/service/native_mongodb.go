package service

import (
	"context"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoNativeClient interface {
	Ping(context.Context) error
	ListDatabaseNames(context.Context) ([]string, error)
	ListCollectionNames(context.Context, string) ([]string, error)
	Close(context.Context) error
}

type MongoNativeClientFactory interface {
	New(context.Context, NativeDatabaseConfig) (MongoNativeClient, error)
}

type MongoNativeProvider struct {
	factory MongoNativeClientFactory
}

func NewMongoNativeProvider(factory MongoNativeClientFactory) *MongoNativeProvider {
	return &MongoNativeProvider{factory: factory}
}

func NewDefaultMongoNativeProvider() *MongoNativeProvider {
	return NewMongoNativeProvider(mongoGoClientFactory{})
}

func (p *MongoNativeProvider) Test(ctx context.Context, cfg NativeDatabaseConfig) error {
	client, err := p.factory.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer client.Close(context.Background())
	return client.Ping(ctx)
}

func (p *MongoNativeProvider) Connect(ctx context.Context, cfg NativeDatabaseConfig) (NativeDatabaseClient, error) {
	client, err := p.factory.New(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx); err != nil {
		_ = client.Close(context.Background())
		return nil, err
	}
	return &mongoNativeSession{client: client}, nil
}

type mongoNativeSession struct {
	client MongoNativeClient
}

func (s *mongoNativeSession) ListPrimaryResources(ctx context.Context) ([]NativeResource, error) {
	databases, err := s.client.ListDatabaseNames(ctx)
	if err != nil {
		return nil, err
	}
	sort.Strings(databases)
	resources := make([]NativeResource, 0, len(databases))
	for _, database := range databases {
		resources = append(resources, NativeResource{Kind: NativeResourceKindDatabase, Name: database})
	}
	return resources, nil
}

func (s *mongoNativeSession) ListSecondaryResources(ctx context.Context, database string) ([]NativeResource, error) {
	collections, err := s.client.ListCollectionNames(ctx, database)
	if err != nil {
		return nil, err
	}
	sort.Strings(collections)
	resources := make([]NativeResource, 0, len(collections))
	for _, collection := range collections {
		resources = append(resources, NativeResource{Kind: NativeResourceKindCollection, Name: collection})
	}
	return resources, nil
}

func (s *mongoNativeSession) Close() error {
	return s.client.Close(context.Background())
}

type mongoGoClientFactory struct{}

func (mongoGoClientFactory) New(_ context.Context, cfg NativeDatabaseConfig) (MongoNativeClient, error) {
	client, err := mongo.Connect(options.Client().ApplyURI(mongoNativeURI(cfg)))
	if err != nil {
		return nil, err
	}
	return &mongoGoClient{client: client}, nil
}

func mongoNativeURI(cfg NativeDatabaseConfig) string {
	host := strings.TrimSpace(cfg.Host)
	if host == "" {
		host = "127.0.0.1"
	}
	port := cfg.Port
	if port == 0 {
		port = 27017
	}
	uri := &url.URL{Scheme: "mongodb", Host: net.JoinHostPort(host, strconv.Itoa(port))}
	if cfg.User != "" {
		uri.User = url.UserPassword(cfg.User, cfg.Password)
	}
	return uri.String()
}

type mongoGoClient struct {
	client *mongo.Client
}

func (c *mongoGoClient) Ping(ctx context.Context) error {
	return c.client.Ping(ctx, nil)
}

func (c *mongoGoClient) ListDatabaseNames(ctx context.Context) ([]string, error) {
	return c.client.ListDatabaseNames(ctx, bson.D{})
}

func (c *mongoGoClient) ListCollectionNames(ctx context.Context, database string) ([]string, error) {
	return c.client.Database(database).ListCollectionNames(ctx, bson.D{})
}

func (c *mongoGoClient) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}
