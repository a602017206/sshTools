package service

import (
	"context"
	"encoding/json"
	"testing"
)

func TestRedisNativeSessionMutatesKeyValue(t *testing.T) {
	client := &fakeRedisNativeClient{}
	provider := NewRedisNativeProvider(&fakeRedisNativeClientFactory{clients: map[int]*fakeRedisNativeClient{0: client}})
	connected, err := provider.Connect(context.Background(), NativeDatabaseConfig{Type: NativeDatabaseTypeRedis})
	if err != nil {
		t.Fatalf("connect Redis: %v", err)
	}
	mutator, ok := connected.(NativeResourceMutator)
	if !ok {
		t.Fatal("expected Redis session to support mutations")
	}
	result, err := mutator.MutateResource(context.Background(), "0", "cache:user", "save", `{"type":"list","items":["a","b"]}`)
	if err != nil || result.Summary == "" {
		t.Fatalf("save list key: result=%+v err=%v", result, err)
	}
	result, err = mutator.MutateResource(context.Background(), "0", "cache:user", "delete", `{}`)
	if err != nil || result.Summary == "" {
		t.Fatalf("delete key: result=%+v err=%v", result, err)
	}
}

func TestElasticsearchNativeSessionExecutesQueryAndMutations(t *testing.T) {
	client := &fakeElasticsearchNativeClient{}
	provider := NewElasticsearchNativeProvider(&fakeElasticsearchNativeClientFactory{client: client})
	connected, err := provider.Connect(context.Background(), NativeDatabaseConfig{Type: NativeDatabaseTypeElasticsearch})
	if err != nil {
		t.Fatalf("connect Elasticsearch: %v", err)
	}
	executor, ok := connected.(NativeQueryExecutor)
	if !ok {
		t.Fatal("expected Elasticsearch session to support queries")
	}
	queryResult, err := executor.ExecuteQuery(context.Background(), "", "products", `{"query":{"match_all":{}}}`)
	if err != nil || queryResult.Content == "" {
		t.Fatalf("execute query: result=%+v err=%v", queryResult, err)
	}
	mutator, ok := connected.(NativeResourceMutator)
	if !ok {
		t.Fatal("expected Elasticsearch session to support mutations")
	}
	_, err = mutator.MutateResource(context.Background(), "", "products", "index_document", `{"id":"p-1","document":{"name":"Keyboard"}}`)
	if err != nil {
		t.Fatalf("index document: %v", err)
	}
}

func TestNativeDatabaseServiceRejectsUnsupportedOperations(t *testing.T) {
	stub := &unsupportedOperationsProvider{client: &unsupportedOperationsClient{}}
	service := NewNativeDatabaseService(map[NativeDatabaseType]NativeDatabaseProvider{
		NativeDatabaseTypeKafka: stub,
	})
	if err := service.Connect(context.Background(), "kafka-session", NativeDatabaseConfig{Type: NativeDatabaseTypeKafka}); err != nil {
		t.Fatalf("connect stub session: %v", err)
	}
	_, err := service.ExecuteQuery(context.Background(), "kafka-session", "", "events", `{}`)
	if err == nil {
		t.Fatal("expected unsupported query error")
	}
	_, err = service.MutateResource(context.Background(), "kafka-session", "", "events", "delete_document", `{}`)
	if err == nil {
		t.Fatal("expected unsupported mutation error")
	}
}

type unsupportedOperationsProvider struct {
	client NativeDatabaseClient
}

func (p *unsupportedOperationsProvider) Test(context.Context, NativeDatabaseConfig) error { return nil }

func (p *unsupportedOperationsProvider) Connect(context.Context, NativeDatabaseConfig) (NativeDatabaseClient, error) {
	return p.client, nil
}

type unsupportedOperationsClient struct{}

func (*unsupportedOperationsClient) ListPrimaryResources(context.Context) ([]NativeResource, error) {
	return nil, nil
}

func (*unsupportedOperationsClient) ListSecondaryResources(context.Context, string) ([]NativeResource, error) {
	return nil, nil
}

func (*unsupportedOperationsClient) DescribeResource(context.Context, string, string) (NativeResourceDetails, error) {
	return NativeResourceDetails{}, nil
}

func (*unsupportedOperationsClient) Close() error { return nil }

func TestNormalizeElasticsearchQueryDefaultsAndClampsSize(t *testing.T) {
	body, err := normalizeElasticsearchQuery("")
	if err != nil {
		t.Fatalf("normalize empty query: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("parse query: %v", err)
	}
	if parsed["size"].(float64) != float64(elasticsearchQuerySizeLimit) {
		t.Fatalf("size = %v", parsed["size"])
	}
	body, err = normalizeElasticsearchQuery(`{"query":{"match_all":{}},"size":500}`)
	if err != nil {
		t.Fatalf("normalize large query: %v", err)
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		t.Fatalf("parse clamped query: %v", err)
	}
	if parsed["size"].(float64) != float64(elasticsearchQuerySizeLimit) {
		t.Fatalf("clamped size = %v", parsed["size"])
	}
}
