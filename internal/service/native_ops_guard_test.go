package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestValidateRedisCLIDeniesDangerousCommands(t *testing.T) {
	if err := validateRedisCLI([]any{"FLUSHALL"}, false); err == nil {
		t.Fatal("expected FLUSHALL to be denied")
	}
	if err := validateRedisCLI([]any{"GET", "k"}, true); err != nil {
		t.Fatalf("GET should be allowed readonly: %v", err)
	}
	if err := validateRedisCLI([]any{"SET", "k", "v"}, true); err == nil {
		t.Fatal("SET should be denied in readonly mode")
	}
}

func TestRedisExecuteQueryCLI(t *testing.T) {
	client := &fakeRedisNativeClient{}
	session := &redisNativeSession{database: 0, client: client, factory: &fakeRedisNativeClientFactory{clients: map[int]*fakeRedisNativeClient{0: client}}}
	result, err := session.ExecuteQuery(context.Background(), "0", "", `{"mode":"cli","command":"PING","readOnly":true}`)
	if err != nil {
		t.Fatalf("ExecuteQuery: %v", err)
	}
	if !strings.Contains(result.Summary, "PING") && result.Summary != "ok" {
		t.Fatalf("summary = %q", result.Summary)
	}
}

func TestValidateElasticsearchDevTools(t *testing.T) {
	if err := validateElasticsearchDevTools("GET", "/logs/_search"); err != nil {
		t.Fatalf("expected allowed: %v", err)
	}
	if err := validateElasticsearchDevTools("DELETE", "/logs"); err == nil {
		t.Fatal("DELETE index via Dev Tools should be denied")
	}
	if err := validateElasticsearchDevTools("GET", "http://evil/_search"); err == nil {
		t.Fatal("absolute URL should be denied")
	}
}

func TestNormalizeElasticsearchQueryClampsFrom(t *testing.T) {
	raw, err := normalizeElasticsearchQuery(`{"query":{"match_all":{}},"from":9999,"size":500}`)
	if err != nil {
		t.Fatal(err)
	}
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	if int(body["size"].(float64)) > elasticsearchQuerySizeLimit {
		t.Fatalf("size not clamped: %v", body["size"])
	}
	if int(body["from"].(float64)) > elasticsearchQueryFromLimit {
		t.Fatalf("from not clamped: %v", body["from"])
	}
}
