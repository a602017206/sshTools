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

const elasticsearchQuerySizeLimit = 100

type elasticsearchMutationPayload struct {
	ID       string          `json:"id"`
	Document json.RawMessage `json:"document"`
}

func (s *elasticsearchNativeSession) ExecuteQuery(ctx context.Context, _, name, query string) (NativeQueryResult, error) {
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
