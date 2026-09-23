package copilot

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestModelsURLAddsV1AndStripsChatCompletions(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"https://api.deepseek.com", "https://api.deepseek.com/v1/models"},
		{"https://api.deepseek.com/v1", "https://api.deepseek.com/v1/models"},
		{"https://api.deepseek.com/v1/", "https://api.deepseek.com/v1/models"},
		{"https://api.deepseek.com/v1/chat/completions", "https://api.deepseek.com/v1/models"},
	}
	for _, tt := range tests {
		if got := modelsURL(tt.in); got != tt.want {
			t.Errorf("modelsURL(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestListModelsParsesOpenAIDataAndSkipsEmptyIDs(t *testing.T) {
	var (
		gotMethod string
		gotPath   string
		gotAuth   string
	)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"deepseek-chat"},{"id":""},{"id":"deepseek-reasoner"},{"id":"deepseek-chat"}]}`))
	}))
	defer srv.Close()

	got, err := ListModels(context.Background(), srv.URL, "sk-test", srv.Client())
	if err != nil {
		t.Fatalf("ListModels: %v", err)
	}
	if gotMethod != http.MethodGet {
		t.Fatalf("method = %q, want GET", gotMethod)
	}
	if gotPath != "/v1/models" {
		t.Fatalf("path = %q, want /v1/models", gotPath)
	}
	if gotAuth != "Bearer sk-test" {
		t.Fatalf("Authorization = %q, want Bearer sk-test", gotAuth)
	}
	if len(got) != 2 {
		t.Fatalf("models = %+v, want 2 unique ids", got)
	}
	if got[0].ID != "deepseek-chat" || got[0].Name != "deepseek-chat" {
		t.Fatalf("first = %+v, want id/name deepseek-chat", got[0])
	}
	if got[1].ID != "deepseek-reasoner" {
		t.Fatalf("second = %+v, want deepseek-reasoner", got[1])
	}
}

func TestListModelsKeepsExistingV1Prefix(t *testing.T) {
	var gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		_, _ = io.WriteString(w, `{"data":[{"id":"llama3.1"}]}`)
	}))
	defer srv.Close()

	if _, err := ListModels(context.Background(), srv.URL+"/v1/", "", srv.Client()); err != nil {
		t.Fatalf("ListModels: %v", err)
	}
	if gotPath != "/v1/models" {
		t.Fatalf("path = %q, want /v1/models (must not double /v1)", gotPath)
	}
}

func TestListModelsOmitsAuthorizationWhenKeyEmpty(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_, _ = io.WriteString(w, `{"data":[{"id":"llama3.1"}]}`)
	}))
	defer srv.Close()

	if _, err := ListModels(context.Background(), srv.URL, "  ", srv.Client()); err != nil {
		t.Fatalf("ListModels: %v", err)
	}
	if gotAuth != "" {
		t.Fatalf("Authorization = %q, want empty for Ollama-style local endpoints", gotAuth)
	}
}

func TestListModels401RedactsAPIKey(t *testing.T) {
	const apiKey = "sk-secret-key"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":{"message":"Incorrect API key provided: sk-secret-key"}}`, http.StatusUnauthorized)
	}))
	defer srv.Close()

	_, err := ListModels(context.Background(), srv.URL, apiKey, srv.Client())
	if err == nil {
		t.Fatal("expected error on HTTP 401")
	}
	text := err.Error()
	if !strings.Contains(text, "401") {
		t.Fatalf("error should mention 401: %q", text)
	}
	if strings.Contains(text, apiKey) || strings.Contains(text, "Bearer "+apiKey) {
		t.Fatalf("error must not include API key: %q", text)
	}
}

func TestListModelsRequiresBaseURL(t *testing.T) {
	_, err := ListModels(context.Background(), "  ", "sk-test", nil)
	if err == nil || !strings.Contains(err.Error(), "Base URL") {
		t.Fatalf("expected Base URL error, got %v", err)
	}
}

func TestResolveListModelsAPIKey(t *testing.T) {
	got, err := ResolveListModelsAPIKey(" draft-key ", "stored", "deepseek")
	if err != nil || got != "draft-key" {
		t.Fatalf("draft should win: got %q err %v", got, err)
	}
	got, err = ResolveListModelsAPIKey("", "stored-key", "openai")
	if err != nil || got != "stored-key" {
		t.Fatalf("stored fallback: got %q err %v", got, err)
	}
	if _, err := ResolveListModelsAPIKey("", "", "deepseek"); err == nil || !strings.Contains(err.Error(), "API Key") {
		t.Fatalf("cloud provider without key should fail, got %v", err)
	}
	got, err = ResolveListModelsAPIKey("", "", "ollama")
	if err != nil || got != "" {
		t.Fatalf("ollama may have empty key: got %q err %v", got, err)
	}
}
