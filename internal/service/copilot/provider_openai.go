package copilot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const maxChatResponseBytes = 8 << 20

// OpenAICompatible talks to an OpenAI Chat Completions compatible API.
type OpenAICompatible struct {
	endpoint string
	apiKey   string
	client   *http.Client
}

var _ Provider = (*OpenAICompatible)(nil)

// NewOpenAICompatible builds a provider. baseURL may or may not already include /v1.
func NewOpenAICompatible(baseURL, apiKey string, client *http.Client) *OpenAICompatible {
	if client == nil {
		client = http.DefaultClient
	}
	return &OpenAICompatible{
		endpoint: chatCompletionsURL(baseURL),
		apiKey:   apiKey,
		client:   client,
	}
}

func chatCompletionsURL(baseURL string) string {
	base := strings.TrimRight(baseURL, "/")
	if !strings.HasSuffix(base, "/v1") {
		base += "/v1"
	}
	return base + "/chat/completions"
}

// Chat posts model, messages, and tools to /v1/chat/completions.
// Timeout is controlled by ctx. Errors never include request headers or the API key.
func (p *OpenAICompatible) Chat(ctx context.Context, model string, messages []Message, tools []ToolSpec) (Message, error) {
	raw, err := json.Marshal(openaiChatRequest{
		Model:    model,
		Messages: toOpenAIMessages(messages),
		Tools:    toOpenAITools(tools),
	})
	if err != nil {
		return Message{}, fmt.Errorf("openai chat: encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint, bytes.NewReader(raw))
	if err != nil {
		return Message{}, fmt.Errorf("openai chat: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return Message{}, fmt.Errorf("openai chat: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxChatResponseBytes))
	if err != nil {
		return Message{}, fmt.Errorf("openai chat: read response: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return Message{}, httpStatusError(resp.StatusCode, body, p.apiKey)
	}

	var parsed openaiChatResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return Message{}, fmt.Errorf("openai chat: decode response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return Message{}, fmt.Errorf("openai chat: empty choices")
	}
	return fromOpenAIMessage(parsed.Choices[0].Message), nil
}

func httpStatusError(status int, body []byte, apiKey string) error {
	text := strings.TrimSpace(string(body))
	if apiKey != "" {
		text = strings.ReplaceAll(text, apiKey, "[REDACTED]")
	}
	if text == "" {
		return fmt.Errorf("openai chat: HTTP %d", status)
	}
	return fmt.Errorf("openai chat: HTTP %d: %s", status, text)
}

type openaiChatRequest struct {
	Model    string          `json:"model"`
	Messages []openaiMessage `json:"messages"`
	Tools    []openaiTool    `json:"tools,omitempty"`
}

type openaiChatResponse struct {
	Choices []struct {
		Message openaiMessage `json:"message"`
	} `json:"choices"`
}

type openaiMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
	ToolCalls  []openaiToolCall `json:"tool_calls,omitempty"`
}

type openaiToolCall struct {
	ID       string             `json:"id"`
	Type     string             `json:"type"`
	Function openaiFunctionCall `json:"function"`
}

type openaiFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openaiTool struct {
	Type     string             `json:"type"`
	Function openaiFunctionSpec `json:"function"`
}

type openaiFunctionSpec struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Parameters  json.RawMessage `json:"parameters,omitempty"`
}

func toOpenAIMessages(msgs []Message) []openaiMessage {
	out := make([]openaiMessage, 0, len(msgs))
	for _, m := range msgs {
		om := openaiMessage{
			Role:       m.Role,
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
		}
		if len(m.ToolCalls) > 0 {
			om.ToolCalls = make([]openaiToolCall, len(m.ToolCalls))
			for i, tc := range m.ToolCalls {
				om.ToolCalls[i] = openaiToolCall{
					ID:   tc.ID,
					Type: "function",
					Function: openaiFunctionCall{
						Name:      tc.Name,
						Arguments: tc.Arguments,
					},
				}
			}
		}
		out = append(out, om)
	}
	return out
}

func toOpenAITools(tools []ToolSpec) []openaiTool {
	if len(tools) == 0 {
		return nil
	}
	out := make([]openaiTool, len(tools))
	for i, t := range tools {
		out[i] = openaiTool{
			Type: "function",
			Function: openaiFunctionSpec{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		}
	}
	return out
}

func fromOpenAIMessage(m openaiMessage) Message {
	msg := Message{
		Role:       m.Role,
		Content:    m.Content,
		ToolCallID: m.ToolCallID,
	}
	if len(m.ToolCalls) == 0 {
		return msg
	}
	msg.ToolCalls = make([]ToolCall, len(m.ToolCalls))
	for i, tc := range m.ToolCalls {
		msg.ToolCalls[i] = ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		}
	}
	return msg
}
