package copilot

import "testing"

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

func TestParseArtifactRejectsUnknownType(t *testing.T) {
	if _, ok := ParseArtifact(`{"type":"python","content":"print(1)","summary":"x"}`); ok {
		t.Fatal("unknown type must be rejected")
	}
}
