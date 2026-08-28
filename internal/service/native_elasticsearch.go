package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
)

type ElasticsearchNativeClient interface {
	Ping(context.Context) error
	ListIndices(context.Context) ([]string, error)
	DescribeIndex(context.Context, string) (NativeResourceDetails, error)
	SearchIndex(context.Context, string, string) (NativeQueryResult, error)
	IndexDocument(context.Context, string, string) (NativeMutationResult, error)
	UpdateDocument(context.Context, string, string) (NativeMutationResult, error)
	DeleteDocument(context.Context, string, string) (NativeMutationResult, error)
	Close() error
}

type ElasticsearchNativeClientFactory interface {
	New(NativeDatabaseConfig) (ElasticsearchNativeClient, error)
}

type ElasticsearchNativeProvider struct {
	factory ElasticsearchNativeClientFactory
}

func NewElasticsearchNativeProvider(factory ElasticsearchNativeClientFactory) *ElasticsearchNativeProvider {
	return &ElasticsearchNativeProvider{factory: factory}
}

func NewDefaultElasticsearchNativeProvider() *ElasticsearchNativeProvider {
	return NewElasticsearchNativeProvider(elasticsearchGoClientFactory{})
}

func (p *ElasticsearchNativeProvider) Test(ctx context.Context, cfg NativeDatabaseConfig) error {
	client, err := p.factory.New(cfg)
	if err != nil {
		return err
	}
	defer client.Close()
	return client.Ping(ctx)
}

func (p *ElasticsearchNativeProvider) Connect(ctx context.Context, cfg NativeDatabaseConfig) (NativeDatabaseClient, error) {
	client, err := p.factory.New(cfg)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx); err != nil {
		_ = client.Close()
		return nil, err
	}
	return &elasticsearchNativeSession{client: client}, nil
}

type elasticsearchNativeSession struct {
	client ElasticsearchNativeClient
}

func (s *elasticsearchNativeSession) ListPrimaryResources(ctx context.Context) ([]NativeResource, error) {
	indices, err := s.client.ListIndices(ctx)
	if err != nil {
		return nil, err
	}
	sort.Strings(indices)
	resources := make([]NativeResource, 0, len(indices))
	for _, index := range indices {
		resources = append(resources, NativeResource{Kind: NativeResourceKindIndex, Name: index})
	}
	return resources, nil
}

func (*elasticsearchNativeSession) ListSecondaryResources(context.Context, string) ([]NativeResource, error) {
	return []NativeResource{}, nil
}

func (s *elasticsearchNativeSession) Close() error {
	return s.client.Close()
}

func (s *elasticsearchNativeSession) DescribeResource(ctx context.Context, _ string, name string) (NativeResourceDetails, error) {
	return s.client.DescribeIndex(ctx, name)
}

type elasticsearchGoClientFactory struct{}

func (elasticsearchGoClientFactory) New(cfg NativeDatabaseConfig) (ElasticsearchNativeClient, error) {
	host := strings.TrimSpace(cfg.Host)
	if host == "" {
		host = "127.0.0.1"
	}
	port := cfg.Port
	if port == 0 {
		port = 9200
	}
	endpoint := (&url.URL{Scheme: "http", Host: host + ":" + strconv.Itoa(port)}).String()
	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{endpoint},
		Username:  cfg.User,
		Password:  cfg.Password,
	})
	if err != nil {
		return nil, err
	}
	return &elasticsearchGoClient{client: client}, nil
}

type elasticsearchGoClient struct {
	client *elasticsearch.Client
}

func (c *elasticsearchGoClient) Ping(ctx context.Context) error {
	response, err := c.perform(ctx, http.MethodGet, "/", nil)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	return elasticsearchResponseError(response)
}

func (c *elasticsearchGoClient) ListIndices(ctx context.Context) ([]string, error) {
	response, err := c.perform(ctx, http.MethodGet, "/_cat/indices?format=json&h=index", nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if err := elasticsearchResponseError(response); err != nil {
		return nil, err
	}
	var rows []struct {
		Index string `json:"index"`
	}
	if err := json.NewDecoder(response.Body).Decode(&rows); err != nil {
		return nil, fmt.Errorf("解析 Elasticsearch 索引列表失败: %w", err)
	}
	indices := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.Index != "" {
			indices = append(indices, row.Index)
		}
	}
	return indices, nil
}

func (c *elasticsearchGoClient) Close() error {
	c.client.Close(context.Background())
	return nil
}

func (c *elasticsearchGoClient) DescribeIndex(ctx context.Context, name string) (NativeResourceDetails, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return NativeResourceDetails{}, fmt.Errorf("Elasticsearch 索引名不能为空")
	}
	response, err := c.perform(ctx, http.MethodGet, "/"+url.PathEscape(name)+"/_search?size=20&track_total_hits=false", nil)
	if err != nil {
		return NativeResourceDetails{}, err
	}
	defer response.Body.Close()
	if err := elasticsearchResponseError(response); err != nil {
		return NativeResourceDetails{}, err
	}
	var payload struct {
		Hits struct {
			Total json.RawMessage   `json:"total"`
			Hits  []json.RawMessage `json:"hits"`
		} `json:"hits"`
	}
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return NativeResourceDetails{}, fmt.Errorf("解析 Elasticsearch 文档预览失败: %w", err)
	}
	content, err := json.Marshal(map[string]any{"total": json.RawMessage(payload.Hits.Total), "documents": payload.Hits.Hits})
	if err != nil {
		return NativeResourceDetails{}, err
	}
	return NativeResourceDetails{Kind: NativeResourceKindIndex, Name: name, Summary: fmt.Sprintf("%d 条文档预览", len(payload.Hits.Hits)), Content: string(content)}, nil
}

func elasticsearchResponseError(response *http.Response) error {
	if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
	return fmt.Errorf("Elasticsearch 返回 HTTP %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
}
