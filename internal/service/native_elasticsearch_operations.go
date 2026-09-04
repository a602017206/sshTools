package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	elasticsearchQuerySizeLimit = 100
	elasticsearchQueryFromLimit = 400
)

type elasticsearchMutationPayload struct {
	ID       string          `json:"id"`
	Document json.RawMessage `json:"document"`
}

type elasticsearchDevToolsQuery struct {
	Method string          `json:"method"`
	Path   string          `json:"path"`
	Body   json.RawMessage `json:"body"`
}

func (s *elasticsearchNativeSession) ExecuteQuery(ctx context.Context, _, name, query string) (NativeQueryResult, error) {
	query = strings.TrimSpace(query)
	if strings.HasPrefix(query, "{") {
		var envelope elasticsearchDevToolsQuery
		if err := json.Unmarshal([]byte(query), &envelope); err == nil && strings.TrimSpace(envelope.Path) != "" {
			return s.client.PerformRequest(ctx, envelope.Method, envelope.Path, string(envelope.Body))
		}
	}
	return s.client.SearchIndex(ctx, name, query)
}

func (s *elasticsearchNativeSession) MutateResource(ctx context.Context, _, name, operation, payload string) (NativeMutationResult, error) {
	switch strings.TrimSpace(operation) {
	case "index_document":
		return s.client.IndexDocument(ctx, name, payload)
	case "update_document":
		return s.client.UpdateDocument(ctx, name, payload)
	case "delete_document":
		return s.client.DeleteDocument(ctx, name, payload)
	case "create_index":
		return s.client.CreateIndex(ctx, name, payload)
	case "delete_index":
		return s.client.DeleteIndex(ctx, name)
	case "refresh_index":
		return s.client.RefreshIndex(ctx, name)
	default:
		return NativeMutationResult{}, fmt.Errorf("不支持的 Elasticsearch 操作: %s", operation)
	}
}

func clampElasticsearchQuerySize(body map[string]any, maxSize int) {
	sizeValue, exists := body["size"]
	if !exists {
		body["size"] = maxSize
		return
	}
	switch typed := sizeValue.(type) {
	case float64:
		if int(typed) > maxSize {
			body["size"] = maxSize
		}
	case int:
		if typed > maxSize {
			body["size"] = maxSize
		}
	case json.Number:
		if parsed, err := typed.Int64(); err == nil && int(parsed) > maxSize {
			body["size"] = maxSize
		}
	default:
		body["size"] = maxSize
	}
}

func clampElasticsearchQueryFrom(body map[string]any, maxFrom int) {
	fromValue, exists := body["from"]
	if !exists {
		return
	}
	switch typed := fromValue.(type) {
	case float64:
		if int(typed) > maxFrom {
			body["from"] = maxFrom
		}
	case int:
		if typed > maxFrom {
			body["from"] = maxFrom
		}
	case json.Number:
		if parsed, err := typed.Int64(); err == nil && int(parsed) > maxFrom {
			body["from"] = maxFrom
		}
	}
}

func normalizeElasticsearchQuery(query string) ([]byte, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		query = `{"query":{"match_all":{}}}`
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(query), &body); err != nil {
		return nil, fmt.Errorf("Elasticsearch 查询必须是合法 JSON: %w", err)
	}
	clampElasticsearchQuerySize(body, elasticsearchQuerySizeLimit)
	clampElasticsearchQueryFrom(body, elasticsearchQueryFromLimit)
	return json.Marshal(body)
}

func parseElasticsearchMutationPayload(payload string) (elasticsearchMutationPayload, error) {
	payload = strings.TrimSpace(payload)
	if payload == "" {
		return elasticsearchMutationPayload{}, fmt.Errorf("变更 payload 不能为空")
	}
	var parsed elasticsearchMutationPayload
	if err := json.Unmarshal([]byte(payload), &parsed); err != nil {
		return elasticsearchMutationPayload{}, fmt.Errorf("变更 payload 必须是合法 JSON: %w", err)
	}
	return parsed, nil
}

func validateElasticsearchDevTools(method, path string) error {
	method = strings.ToUpper(strings.TrimSpace(method))
	if method != http.MethodGet && method != http.MethodPost && method != http.MethodPut && method != http.MethodHead {
		return fmt.Errorf("Dev Tools 仅允许 GET/POST/PUT/HEAD")
	}
	path = strings.TrimSpace(path)
	if path == "" || !strings.HasPrefix(path, "/") {
		return fmt.Errorf("Dev Tools path 必须以 / 开头")
	}
	if strings.Contains(path, "://") || strings.Contains(path, "..") {
		return fmt.Errorf("Dev Tools path 非法")
	}
	allowed := []string{
		"/_search", "/_mapping", "/_stats", "/_doc", "/_update", "/_refresh",
		"/_cluster/health", "/_cluster/stats", "/_nodes", "/_cat/",
	}
	lower := strings.ToLower(path)
	for _, prefix := range allowed {
		if strings.Contains(lower, prefix) || strings.HasPrefix(lower, prefix) {
			return nil
		}
	}
	// /index/_search style
	parts := strings.Split(strings.Trim(lower, "/"), "/")
	if len(parts) >= 2 {
		action := parts[len(parts)-1]
		switch action {
		case "_search", "_mapping", "_stats", "_refresh", "_doc":
			return nil
		}
		if len(parts) >= 3 && (parts[1] == "_doc" || parts[1] == "_update" || parts[1] == "_search") {
			return nil
		}
	}
	return fmt.Errorf("Dev Tools 路径不在白名单内: %s", path)
}

func (c *elasticsearchGoClient) SearchIndex(ctx context.Context, index, query string) (NativeQueryResult, error) {
	index = strings.TrimSpace(index)
	if index == "" {
		return NativeQueryResult{}, fmt.Errorf("Elasticsearch 索引名不能为空")
	}
	body, err := normalizeElasticsearchQuery(query)
	if err != nil {
		return NativeQueryResult{}, err
	}
	response, err := c.perform(ctx, http.MethodPost, "/"+url.PathEscape(index)+"/_search", bytes.NewReader(body))
	if err != nil {
		return NativeQueryResult{}, err
	}
	defer response.Body.Close()
	if err := elasticsearchResponseError(response); err != nil {
		return NativeQueryResult{}, err
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return NativeQueryResult{}, fmt.Errorf("读取 Elasticsearch 查询结果失败: %w", err)
	}
	var payload struct {
		Hits struct {
			Total json.RawMessage   `json:"total"`
			Hits  []json.RawMessage `json:"hits"`
		} `json:"hits"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return NativeQueryResult{}, fmt.Errorf("解析 Elasticsearch 查询结果失败: %w", err)
	}
	content, err := json.Marshal(map[string]any{
		"total": payload.Hits.Total,
		"hits":  payload.Hits.Hits,
		"raw":   json.RawMessage(raw),
	})
	if err != nil {
		return NativeQueryResult{}, err
	}
	return NativeQueryResult{
		Summary: fmt.Sprintf("返回 %d 条命中", len(payload.Hits.Hits)),
		Content: string(content),
	}, nil
}

func (c *elasticsearchGoClient) PerformRequest(ctx context.Context, method, path, body string) (NativeQueryResult, error) {
	method = strings.ToUpper(strings.TrimSpace(method))
	if method == "" {
		method = http.MethodGet
	}
	if err := validateElasticsearchDevTools(method, path); err != nil {
		return NativeQueryResult{}, err
	}
	var reader io.Reader
	if strings.TrimSpace(body) != "" && body != "null" {
		reader = strings.NewReader(body)
	}
	response, err := c.perform(ctx, method, path, reader)
	if err != nil {
		return NativeQueryResult{}, err
	}
	defer response.Body.Close()
	if err := elasticsearchResponseError(response); err != nil {
		return NativeQueryResult{}, err
	}
	raw, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return NativeQueryResult{}, err
	}
	content, err := json.Marshal(map[string]any{
		"status": response.StatusCode,
		"body":   json.RawMessage(raw),
	})
	if err != nil {
		content, _ = json.Marshal(map[string]any{"status": response.StatusCode, "text": string(raw)})
	}
	return NativeQueryResult{
		Summary: fmt.Sprintf("%s %s → %d", method, path, response.StatusCode),
		Content: string(content),
	}, nil
}

func (c *elasticsearchGoClient) CreateIndex(ctx context.Context, index, payload string) (NativeMutationResult, error) {
	index = strings.TrimSpace(index)
	if index == "" {
		return NativeMutationResult{}, fmt.Errorf("索引名不能为空")
	}
	body := strings.TrimSpace(payload)
	if body == "" {
		body = "{}"
	}
	if !json.Valid([]byte(body)) {
		return NativeMutationResult{}, fmt.Errorf("create_index payload 必须是合法 JSON")
	}
	response, err := c.perform(ctx, http.MethodPut, "/"+url.PathEscape(index), strings.NewReader(body))
	if err != nil {
		return NativeMutationResult{}, err
	}
	defer response.Body.Close()
	if err := elasticsearchResponseError(response); err != nil {
		return NativeMutationResult{}, err
	}
	raw, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
	return NativeMutationResult{Summary: "索引已创建", Content: strings.TrimSpace(string(raw))}, nil
}

func (c *elasticsearchGoClient) DeleteIndex(ctx context.Context, index string) (NativeMutationResult, error) {
	index = strings.TrimSpace(index)
	if index == "" {
		return NativeMutationResult{}, fmt.Errorf("索引名不能为空")
	}
	response, err := c.perform(ctx, http.MethodDelete, "/"+url.PathEscape(index), nil)
	if err != nil {
		return NativeMutationResult{}, err
	}
	defer response.Body.Close()
	if err := elasticsearchResponseError(response); err != nil {
		return NativeMutationResult{}, err
	}
	raw, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
	return NativeMutationResult{Summary: "索引已删除", Content: strings.TrimSpace(string(raw))}, nil
}

func (c *elasticsearchGoClient) RefreshIndex(ctx context.Context, index string) (NativeMutationResult, error) {
	index = strings.TrimSpace(index)
	if index == "" {
		return NativeMutationResult{}, fmt.Errorf("索引名不能为空")
	}
	response, err := c.perform(ctx, http.MethodPost, "/"+url.PathEscape(index)+"/_refresh", nil)
	if err != nil {
		return NativeMutationResult{}, err
	}
	defer response.Body.Close()
	if err := elasticsearchResponseError(response); err != nil {
		return NativeMutationResult{}, err
	}
	raw, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
	return NativeMutationResult{Summary: "索引已刷新", Content: strings.TrimSpace(string(raw))}, nil
}

func (c *elasticsearchGoClient) IndexDocument(ctx context.Context, index, payload string) (NativeMutationResult, error) {
	parsed, err := parseElasticsearchMutationPayload(payload)
	if err != nil {
		return NativeMutationResult{}, err
	}
	if len(parsed.Document) == 0 {
		return NativeMutationResult{}, fmt.Errorf("document 不能为空")
	}
	path := "/" + url.PathEscape(index) + "/_doc"
	if strings.TrimSpace(parsed.ID) != "" {
		path += "/" + url.PathEscape(parsed.ID)
	}
	response, err := c.perform(ctx, http.MethodPut, path, bytes.NewReader(parsed.Document))
	if err != nil {
		return NativeMutationResult{}, err
	}
	defer response.Body.Close()
	if err := elasticsearchResponseError(response); err != nil {
		return NativeMutationResult{}, err
	}
	body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
	return NativeMutationResult{Summary: "文档已写入", Content: strings.TrimSpace(string(body))}, nil
}

func (c *elasticsearchGoClient) UpdateDocument(ctx context.Context, index, payload string) (NativeMutationResult, error) {
	parsed, err := parseElasticsearchMutationPayload(payload)
	if err != nil {
		return NativeMutationResult{}, err
	}
	id := strings.TrimSpace(parsed.ID)
	if id == "" {
		return NativeMutationResult{}, fmt.Errorf("update_document 需要 id")
	}
	if len(parsed.Document) == 0 {
		return NativeMutationResult{}, fmt.Errorf("document 不能为空")
	}
	updateBody, err := json.Marshal(map[string]json.RawMessage{"doc": parsed.Document})
	if err != nil {
		return NativeMutationResult{}, err
	}
	path := "/" + url.PathEscape(index) + "/_update/" + url.PathEscape(id)
	response, err := c.perform(ctx, http.MethodPost, path, bytes.NewReader(updateBody))
	if err != nil {
		return NativeMutationResult{}, err
	}
	defer response.Body.Close()
	if err := elasticsearchResponseError(response); err != nil {
		return NativeMutationResult{}, err
	}
	body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
	return NativeMutationResult{Summary: "文档已更新", Content: strings.TrimSpace(string(body))}, nil
}

func (c *elasticsearchGoClient) DeleteDocument(ctx context.Context, index, payload string) (NativeMutationResult, error) {
	parsed, err := parseElasticsearchMutationPayload(payload)
	if err != nil {
		return NativeMutationResult{}, err
	}
	id := strings.TrimSpace(parsed.ID)
	if id == "" {
		return NativeMutationResult{}, fmt.Errorf("delete_document 需要 id")
	}
	path := "/" + url.PathEscape(index) + "/_doc/" + url.PathEscape(id)
	response, err := c.perform(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return NativeMutationResult{}, err
	}
	defer response.Body.Close()
	if err := elasticsearchResponseError(response); err != nil {
		return NativeMutationResult{}, err
	}
	body, _ := io.ReadAll(io.LimitReader(response.Body, 4096))
	return NativeMutationResult{Summary: "文档已删除", Content: strings.TrimSpace(string(body))}, nil
}

func (c *elasticsearchGoClient) perform(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, method, path, body)
	if err != nil {
		return nil, err
	}
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	return c.client.Perform(request)
}
