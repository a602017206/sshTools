package copilot

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestOpenAICompatibleSendsPassthroughModelAndAuth(t *testing.T) {
	var (
		gotMethod string
		gotPath   string
		gotAuth   string
		gotBody   map[string]any
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
			http.Error(w, "read body", http.StatusInternalServerError)
			return
		}
		if err := json.Unmarshal(raw, &gotBody); err != nil {
			t.Errorf("decode body: %v", err)
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices":[{
				"message":{
					"role":"assistant",
					"content":null,
					"tool_calls":[{
						"id":"call_1",
						"type":"function",
						"function":{"name":"list_tables","arguments":"{\"schema\":\"public\"}"}
					}]
				}
			}]
		}`))
	}))
	defer srv.Close()

	p := NewOpenAICompatible(srv.URL, "sk-test", srv.Client())
	msg, err := p.Chat(context.Background(), "deepseek-chat", []Message{
		{Role: "user", Content: "列出表"},
	}, []ToolSpec{{
		Name:        "list_tables",
		Description: "List tables",
		Parameters:  json.RawMessage(`{"type":"object","properties":{}}`),
	}})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}

	if gotMethod != http.MethodPost {
		t.Fatalf("method = %q, want POST", gotMethod)
	}
	if gotPath != "/v1/chat/completions" {
		t.Fatalf("url path = %q, want /v1/chat/completions", gotPath)
	}
	if gotAuth != "Bearer sk-test" {
		t.Fatalf("Authorization = %q, want %q", gotAuth, "Bearer sk-test")
	}
	model, _ := gotBody["model"].(string)
	if model != "deepseek-chat" {
		t.Fatalf("body.model = %q, want deepseek-chat (must not be a hardcoded gpt name)", model)
	}
	if strings.HasPrefix(model, "gpt-") {
		t.Fatalf("body.model %q looks like a hardcoded gpt name", model)
	}
	if len(msg.ToolCalls) != 1 {
		t.Fatalf("ToolCalls len = %d, want 1: %+v", len(msg.ToolCalls), msg.ToolCalls)
	}
	if msg.ToolCalls[0].Name != "list_tables" {
		t.Fatalf("ToolCall.Name = %q, want list_tables", msg.ToolCalls[0].Name)
	}
	if msg.ToolCalls[0].Arguments != `{"schema":"public"}` {
		t.Fatalf("ToolCall.Arguments = %q, want {\"schema\":\"public\"}", msg.ToolCalls[0].Arguments)
	}
	if msg.ToolCalls[0].ID != "call_1" {
		t.Fatalf("ToolCall.ID = %q, want call_1", msg.ToolCalls[0].ID)
	}
}

func TestOpenAICompatibleKeepsExistingV1Prefix(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"ok"}}]}`))
	}))
	defer srv.Close()

	p := NewOpenAICompatible(srv.URL+"/v1/", "sk-test", srv.Client())
	if _, err := p.Chat(context.Background(), "deepseek-chat", []Message{
		{Role: "user", Content: "hi"},
	}, nil); err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if gotPath != "/v1/chat/completions" {
		t.Fatalf("url path = %q, want /v1/chat/completions (must not double /v1)", gotPath)
	}
}

func TestOpenAICompatible401ErrorIsReadableWithoutAPIKey(t *testing.T) {
	const apiKey = "sk-test"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":{"message":"Incorrect API key provided"}}`, http.StatusUnauthorized)
	}))
	defer srv.Close()

	p := NewOpenAICompatible(srv.URL, apiKey, srv.Client())
	_, err := p.Chat(context.Background(), "deepseek-chat", []Message{
		{Role: "user", Content: "hi"},
	}, nil)
	if err == nil {
		t.Fatal("expected error on HTTP 401")
	}
	text := err.Error()
	if !strings.Contains(text, "401") {
		t.Fatalf("error should be readable and mention 401: %q", text)
	}
	if strings.Contains(text, apiKey) || strings.Contains(text, "Bearer "+apiKey) {
		t.Fatalf("error must not include API key: %q", text)
	}
}
