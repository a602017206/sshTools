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
	DescribeCluster(context.Context) (NativeResourceDetails, error)
	SearchIndex(context.Context, string, string) (NativeQueryResult, error)
	IndexDocument(context.Context, string, string) (NativeMutationResult, error)
	UpdateDocument(context.Context, string, string) (NativeMutationResult, error)
	DeleteDocument(context.Context, string, string) (NativeMutationResult, error)
	CreateIndex(context.Context, string, string) (NativeMutationResult, error)
	DeleteIndex(context.Context, string) (NativeMutationResult, error)
	RefreshIndex(context.Context, string) (NativeMutationResult, error)
	PerformRequest(context.Context, string, string, string) (NativeQueryResult, error)
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

func (s *elasticsearchNativeSession) DescribeSession(ctx context.Context) (NativeResourceDetails, error) {
	return s.client.DescribeCluster(ctx)
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

	stats, err := c.fetchIndexCatStats(ctx, name)
	if err != nil {
		return NativeResourceDetails{}, err
	}
	mapping, err := c.fetchIndexMapping(ctx, name)
	if err != nil {
		return NativeResourceDetails{}, err
	}

	content, err := json.Marshal(map[string]any{
		"stats":   stats,
		"mapping": mapping,
	})
	if err != nil {
		return NativeResourceDetails{}, err
	}
	summary := fmt.Sprintf("%s · %s 文档 · %s",
		orDefault(stats["health"], "unknown"),
		orDefault(stats["docsCount"], "0"),
		orDefault(stats["storeSize"], "-"),
	)
	return NativeResourceDetails{
		Kind:    NativeResourceKindIndex,
		Name:    name,
		Summary: summary,
		Content: string(content),
	}, nil
}

func (c *elasticsearchGoClient) DescribeCluster(ctx context.Context) (NativeResourceDetails, error) {
	info, err := c.fetchClusterInfo(ctx)
	if err != nil {
		return NativeResourceDetails{}, err
	}
	health, err := c.fetchClusterHealth(ctx)
	if err != nil {
		return NativeResourceDetails{}, err
	}
	nodes, err := c.fetchClusterNodes(ctx)
	if err != nil {
		return NativeResourceDetails{}, err
	}

	payload := map[string]any{
		"clusterName": orDefault(info["cluster_name"], ""),
		"version":     nestedString(info, "version", "number"),
		"tagline":     orDefault(info["tagline"], ""),
		"health":      orDefault(health["status"], "unknown"),
		"nodeCount":   len(nodes),
		"nodes":       nodes,
		"numberOfNodes":         health["number_of_nodes"],
		"numberOfDataNodes":     health["number_of_data_nodes"],
		"activePrimaryShards":   health["active_primary_shards"],
		"activeShards":          health["active_shards"],
		"relocatingShards":      health["relocating_shards"],
		"initializingShards":    health["initializing_shards"],
		"unassignedShards":      health["unassigned_shards"],
	}
	content, err := json.Marshal(payload)
	if err != nil {
		return NativeResourceDetails{}, err
	}
	nodeCount := len(nodes)
	if value, ok := health["number_of_nodes"].(float64); ok {
		nodeCount = int(value)
	}
	summary := fmt.Sprintf("%s · %s · %d 节点",
		orDefault(payload["clusterName"], "cluster"),
		orDefault(payload["health"], "unknown"),
		nodeCount,
	)
	return NativeResourceDetails{
		Kind:    NativeResourceKindIndex,
		Name:    orDefault(payload["clusterName"], "cluster"),
		Summary: summary,
		Content: string(content),
	}, nil
}

func (c *elasticsearchGoClient) fetchIndexCatStats(ctx context.Context, name string) (map[string]string, error) {
	path := "/_cat/indices/" + url.PathEscape(name) + "?format=json&h=index,health,status,pri,rep,docs.count,docs.deleted,store.size,pri.store.size"
	response, err := c.perform(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if err := elasticsearchResponseError(response); err != nil {
		return nil, err
	}
	var rows []map[string]string
	if err := json.NewDecoder(response.Body).Decode(&rows); err != nil {
		return nil, fmt.Errorf("解析 Elasticsearch 索引统计失败: %w", err)
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("未找到索引: %s", name)
	}
	row := rows[0]
	return map[string]string{
		"index":         row["index"],
		"health":        row["health"],
		"status":        row["status"],
		"primaries":     row["pri"],
		"replicas":      row["rep"],
		"docsCount":     row["docs.count"],
		"docsDeleted":   row["docs.deleted"],
		"storeSize":     row["store.size"],
		"priStoreSize":  row["pri.store.size"],
	}, nil
}

func (c *elasticsearchGoClient) fetchIndexMapping(ctx context.Context, name string) (json.RawMessage, error) {
	response, err := c.perform(ctx, http.MethodGet, "/"+url.PathEscape(name)+"/_mapping", nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if err := elasticsearchResponseError(response); err != nil {
		return nil, err
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, 2<<20))
	if err != nil {
		return nil, fmt.Errorf("读取 Elasticsearch mapping 失败: %w", err)
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("解析 Elasticsearch mapping 失败: %w", err)
	}
	if mapping, ok := payload[name]; ok {
		return mapping, nil
	}
	for _, mapping := range payload {
		return mapping, nil
	}
	return json.RawMessage("{}"), nil
}

func (c *elasticsearchGoClient) fetchClusterInfo(ctx context.Context) (map[string]any, error) {
	response, err := c.perform(ctx, http.MethodGet, "/", nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if err := elasticsearchResponseError(response); err != nil {
		return nil, err
	}
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("解析 Elasticsearch 集群信息失败: %w", err)
	}
	return payload, nil
}

func (c *elasticsearchGoClient) fetchClusterHealth(ctx context.Context) (map[string]any, error) {
	response, err := c.perform(ctx, http.MethodGet, "/_cluster/health", nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if err := elasticsearchResponseError(response); err != nil {
		return nil, err
	}
	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("解析 Elasticsearch 集群健康状态失败: %w", err)
	}
	return payload, nil
}

func (c *elasticsearchGoClient) fetchClusterNodes(ctx context.Context) ([]map[string]string, error) {
	response, err := c.perform(ctx, http.MethodGet, "/_cat/nodes?format=json&h=name,ip,node.role,master,heap.percent,ram.percent,cpu,load_1m,node.total,version", nil)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if err := elasticsearchResponseError(response); err != nil {
		return nil, err
	}
	var rows []map[string]string
	if err := json.NewDecoder(response.Body).Decode(&rows); err != nil {
		return nil, fmt.Errorf("解析 Elasticsearch 节点列表失败: %w", err)
	}
	return rows, nil
}

func orDefault(value any, fallback string) string {
	switch typed := value.(type) {
	case string:
		if strings.TrimSpace(typed) == "" {
			return fallback
		}
		return typed
	case nil:
		return fallback
	default:
		text := fmt.Sprint(typed)
		if strings.TrimSpace(text) == "" || text == "<nil>" {
			return fallback
		}
		return text
	}
}

func nestedString(payload map[string]any, keys ...string) string {
	current := any(payload)
	for _, key := range keys {
		object, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		current = object[key]
	}
	return orDefault(current, "")
}

func elasticsearchResponseError(response *http.Response) error {
	if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
	return fmt.Errorf("Elasticsearch 返回 HTTP %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
}
