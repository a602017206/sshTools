package copilot

import (
	"strings"
	"testing"
)

func TestParseArtifactRequiresJSON(t *testing.T) {
	if _, ok := ParseArtifact("好的，我来写一条 SQL"); ok {
		t.Fatal("plain text must not be an artifact")
	}
}

func TestParseArtifactExtractsJSON(t *testing.T) {
	raw := `{"type":"sql","content":"SELECT 1","summary":"ping","destructive":false}`
	got, ok := ParseArtifact(raw)
	if !ok {
		t.Fatal("expected artifact")
	}
	if got.Type != "sql" || got.Content != "SELECT 1" || got.Summary != "ping" {
		t.Fatalf("unexpected artifact: %+v", got)
	}
	if got.Destructive {
		t.Fatal("SELECT 1 must not be destructive")
	}
}

func TestParseArtifactAcceptsMarkdownFence(t *testing.T) {
	raw := "说明如下：\n```json\n{\"type\":\"shell\",\"content\":\"ls -la\",\"summary\":\"list files\",\"destructive\":false}\n```\n"
	got, ok := ParseArtifact(raw)
	if !ok {
		t.Fatal("expected artifact from fenced JSON")
	}
	if got.Type != "shell" || got.Content != "ls -la" || got.Summary != "list files" {
		t.Fatalf("unexpected artifact: %+v", got)
	}
}

func TestParseArtifactRulesOverrideModelDestructive(t *testing.T) {
	raw := `{"type":"sql","content":"DROP TABLE users","summary":"drop","destructive":false}`
	got, ok := ParseArtifact(raw)
	if !ok {
		t.Fatal("expected artifact")
	}
	if !got.Destructive {
		t.Fatal("rules must override model destructive=false")
	}
}

func TestParseArtifactRejectsEmptyContent(t *testing.T) {
	if _, ok := ParseArtifact(`{"type":"sql","content":"","summary":"x"}`); ok {
		t.Fatal("empty content must be rejected")
	}
}

func TestParseArtifactSkipsLeadingBareDSL(t *testing.T) {
	raw := "说明：\n```json\n{\"query\":{\"match_all\":{}},\"size\":20}\n```\n" +
		`{"type":"native_query","content":{"query":{"match_all":{}},"size":20},"summary":"最近一周","destructive":false}`
	got, ok := ParseArtifact(raw)
	if !ok {
		t.Fatal("expected trailing native_query artifact")
	}
	if got.Type != "native_query" || got.Summary != "最近一周" {
		t.Fatalf("unexpected artifact: %+v", got)
	}
}

func TestExtractArtifactStripsEnvelopeAndPromotesBareQuery(t *testing.T) {
	raw := "先看这个：\n{\"query\":{\"match_all\":{}},\"size\":5}\n"
	art, display := ExtractArtifact(raw, "elasticsearch")
	if art == nil || art.Type != "native_query" {
		t.Fatalf("expected promoted native_query, got %+v", art)
	}
	if strings.Contains(display, `"query"`) {
		t.Fatalf("display should drop bare query JSON, got %q", display)
	}

	wrapped := "说明文字\n{\"type\":\"native_query\",\"content\":\"GET k\",\"summary\":\"读键\",\"destructive\":false}\n"
	art2, display2 := ExtractArtifact(wrapped, "redis")
	if art2 == nil || art2.Content != "GET k" {
		t.Fatalf("unexpected art: %+v", art2)
	}
	if strings.Contains(display2, `"type"`) {
		t.Fatalf("display should strip artifact envelope, got %q", display2)
	}
}
