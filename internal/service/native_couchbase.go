package service

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/couchbase/gocb/v2"
)

type CouchbaseNativeClient interface {
	Ping(context.Context) error
	ListBuckets(context.Context) ([]string, error)
	ListCollections(context.Context, string) ([]string, error)
	Close() error
}

type CouchbaseNativeClientFactory interface {
	New(context.Context, NativeDatabaseConfig) (CouchbaseNativeClient, error)
}

type CouchbaseNativeProvider struct{ factory CouchbaseNativeClientFactory }

func NewCouchbaseNativeProvider(factory CouchbaseNativeClientFactory) *CouchbaseNativeProvider {
	return &CouchbaseNativeProvider{factory: factory}
}

func NewDefaultCouchbaseNativeProvider() *CouchbaseNativeProvider {
	return NewCouchbaseNativeProvider(couchbaseGoClientFactory{})
}

func (p *CouchbaseNativeProvider) Test(ctx context.Context, cfg NativeDatabaseConfig) error {
	client, err := p.factory.New(ctx, cfg)
	if err != nil {
		return err
	}
	defer client.Close()
	return client.Ping(ctx)
}

func (p *CouchbaseNativeProvider) Connect(ctx context.Context, cfg NativeDatabaseConfig) (NativeDatabaseClient, error) {
	client, err := p.factory.New(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx); err != nil {
		_ = client.Close()
		return nil, err
	}
	return &couchbaseNativeSession{client: client}, nil
}

type couchbaseNativeSession struct{ client CouchbaseNativeClient }

func (s *couchbaseNativeSession) ListPrimaryResources(ctx context.Context) ([]NativeResource, error) {
	buckets, err := s.client.ListBuckets(ctx)
	if err != nil {
		return nil, err
	}
	sort.Strings(buckets)
	resources := make([]NativeResource, 0, len(buckets))
	for _, bucket := range buckets {
		resources = append(resources, NativeResource{Kind: NativeResourceKindDatabase, Name: bucket})
	}
	return resources, nil
}

func (s *couchbaseNativeSession) ListSecondaryResources(ctx context.Context, bucket string) ([]NativeResource, error) {
	collections, err := s.client.ListCollections(ctx, bucket)
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

func (s *couchbaseNativeSession) Close() error { return s.client.Close() }

type couchbaseGoClientFactory struct{}

func (couchbaseGoClientFactory) New(ctx context.Context, cfg NativeDatabaseConfig) (CouchbaseNativeClient, error) {
	host := strings.TrimSpace(cfg.Host)
	if host == "" {
		host = "127.0.0.1"
	}
	port := cfg.Port
	if port == 0 {
		port = 8091
	}
	cluster, err := gocb.Connect("couchbase://"+host+":"+strconv.Itoa(port), gocb.ClusterOptions{
		Authenticator: gocb.PasswordAuthenticator{Username: cfg.User, Password: cfg.Password},
	})
	if err != nil {
		return nil, err
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	if err := cluster.WaitUntilReady(timeout, &gocb.WaitUntilReadyOptions{Context: ctx, ServiceTypes: []gocb.ServiceType{gocb.ServiceTypeManagement}}); err != nil {
		_ = cluster.Close(nil)
		return nil, err
	}
	return &couchbaseGoClient{cluster: cluster, timeout: timeout}, nil
}

type couchbaseGoClient struct {
	cluster *gocb.Cluster
	timeout time.Duration
}

func (c *couchbaseGoClient) Ping(ctx context.Context) error {
	_, err := c.cluster.Ping(&gocb.PingOptions{Context: ctx, Timeout: c.timeout, ServiceTypes: []gocb.ServiceType{gocb.ServiceTypeManagement}})
	return err
}

func (c *couchbaseGoClient) ListBuckets(ctx context.Context) ([]string, error) {
	buckets, err := c.cluster.Buckets().GetAllBuckets(&gocb.GetAllBucketsOptions{Context: ctx, Timeout: c.timeout})
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(buckets))
	for name := range buckets {
		names = append(names, name)
	}
	return names, nil
}

func (c *couchbaseGoClient) ListCollections(ctx context.Context, bucketName string) ([]string, error) {
	scopes, err := c.cluster.Bucket(bucketName).Collections().GetAllScopes(&gocb.GetAllScopesOptions{Context: ctx, Timeout: c.timeout})
	if err != nil {
		return nil, err
	}
	collections := make([]string, 0)
	for _, scope := range scopes {
		for _, collection := range scope.Collections {
			collections = append(collections, scope.Name+"."+collection.Name)
		}
	}
	return collections, nil
}

func (c *couchbaseGoClient) Close() error { return c.cluster.Close(nil) }
