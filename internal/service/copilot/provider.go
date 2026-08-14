package copilot

import (
	"context"
	"encoding/json"
)

// Message is a chat turn exchanged with a Provider.
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall is a model-requested function invocation.
type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolSpec describes a function the model may call.
type ToolSpec struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

// Provider is an LLM chat backend that may request tool calls.
type Provider interface {
	Chat(ctx context.Context, model string, messages []Message, tools []ToolSpec) (Message, error)
}
