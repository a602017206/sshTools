package copilot

import (
	"context"
	"errors"
	"strings"
	"testing"

	"AHaSSHTools/internal/config"
)

// fakeProvider returns a scripted sequence of messages and records the tools
// passed to each Chat call.
type fakeProvider struct {
	msgs      []Message
	errs      []error
	calls     int
	toolsCalls [][]ToolSpec
}

func (p *fakeProvider) Chat(ctx context.Context, model string, messages []Message, tools []ToolSpec) (Message, error) {
	i := p.calls
	if i >= len(p.msgs) {
		return Message{}, errors.New("no more scripted messages")
	}
	p.toolsCalls = append(p.toolsCalls, tools)
	msg := p.msgs[i]
	if i < len(p.errs) && p.errs[i] != nil {
		p.calls++
		return msg, p.errs[i]
	}
	p.calls++
	return msg, nil
}

type schemaStub struct{}

func (schemaStub) ListDatabases(sessionID string) ([]string, error) { return []string{"db"}, nil }
func (schemaStub) ListTables(sessionID string) ([]string, error)    { return []string{"t"}, nil }
func (schemaStub) GetTableSchema(sessionID, table string) (*config.TableSchema, error) {
	return &config.TableSchema{}, nil
}

func TestChatParsesArtifactWhenNoToolCalls(t *testing.T) {
	p := &fakeProvider{msgs: []Message{{
		Role:    "assistant",
		Content: `{"type":"sql","content":"SELECT 1","summary":"ping","destructive":false}`,
	}}}
	svc := NewService(p, schemaStub{}, nil)
	resp, err := svc.Chat(context.Background(), ChatRequest{SessionID: "s1", Mode: "database"})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if resp.Artifact == nil || resp.Artifact.Content != "SELECT 1" {
		t.Fatalf("expected artifact SELECT 1, got %+v", resp.Artifact)
	}
}

func TestChatRequestsFinalReplyAfterMaxToolRounds(t *testing.T) {
	// 前 MaxToolRounds 轮全部带 tool_calls（触顶），第 5 次无 tools 调用应产出最终产物。
	toolCall := ToolCall{ID: "c1", Name: "list_tables", Arguments: "{}"}
	rounds := make([]Message, 0, MaxToolRounds+1)
	for i := 0; i < MaxToolRounds; i++ {
		rounds = append(rounds, Message{Role: "assistant", ToolCalls: []ToolCall{toolCall}})
	}
	rounds = append(rounds, Message{
		Role:    "assistant",
		Content: `{"type":"sql","content":"SELECT 2","summary":"final","destructive":false}`,
	})
	p := &fakeProvider{msgs: rounds}
	svc := NewService(p, schemaStub{}, nil)
	resp, err := svc.Chat(context.Background(), ChatRequest{SessionID: "s2", Mode: "database"})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if p.calls != MaxToolRounds+1 {
		t.Fatalf("provider calls = %d, want %d (final no-tools reply)", p.calls, MaxToolRounds+1)
	}
	// 最后一次调用必须不带 tools
	lastTools := p.toolsCalls[len(p.toolsCalls)-1]
	if len(lastTools) != 0 {
		t.Fatalf("final call tools = %d, want 0", len(lastTools))
	}
	if resp.Artifact == nil || resp.Artifact.Content != "SELECT 2" {
		t.Fatalf("expected final artifact SELECT 2, got %+v", resp.Artifact)
	}
	if !strings.Contains(resp.Reply, "SELECT 2") {
		t.Fatalf("expected reply to contain final content, got %q", resp.Reply)
	}
}

func TestChatStopsWhenModelStopsCallingTools(t *testing.T) {
	// 模型第一轮就给最终产物，不应触发额外的无 tools 调用。
	p := &fakeProvider{msgs: []Message{{
		Role:    "assistant",
		Content: `{"type":"sql","content":"SELECT 3","summary":"done","destructive":false}`,
	}}}
	svc := NewService(p, schemaStub{}, nil)
	resp, err := svc.Chat(context.Background(), ChatRequest{SessionID: "s3", Mode: "database"})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if p.calls != 1 {
		t.Fatalf("provider calls = %d, want 1", p.calls)
	}
	if resp.Artifact == nil || resp.Artifact.Content != "SELECT 3" {
		t.Fatalf("expected artifact SELECT 3, got %+v", resp.Artifact)
	}
}
