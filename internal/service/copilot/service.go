package copilot

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"AHaSSHTools/internal/config"
)

const (
	MaxToolRounds        = 4
	MaxToolResultChars   = 8000
	APIKeyCredentialID   = "copilot:api_key"
	maxReplyFallbackText = "未能在有限工具轮次内生成完整结果"
)

const systemPrompt = "你是 SQL/Shell 助手。最终回复必须输出产物 JSON：{\"type\":\"sql 或 shell\",\"content\":\"...\",\"summary\":\"...\",\"destructive\":false}。只使用提供的只读工具获取环境信息，不要索要密码。"

// SchemaReader loads database metadata for copilot tools.
type SchemaReader interface {
	ListDatabases(sessionID string) ([]string, error)
	ListTables(sessionID string) ([]string, error)
	GetTableSchema(sessionID, table string) (*config.TableSchema, error)
}

// CommandRunner runs a command on an existing session without using the user PTY.
type CommandRunner interface {
	ExecuteCommand(sessionID, cmd string, timeout time.Duration) (stdout, stderr string, err error)
}

// ChatRequest is the user-facing copilot turn. It must not carry passwords.
type ChatRequest struct {
	SessionID     string
	Mode          string // ssh | database
	Message       string
	History       []Message
	EditorContent string
	TerminalTail  string
	Host          string
	User          string
	DBType        string
	Database      string
	WorkingDir    string
}

// ChatResponse is the model reply plus any parsed artifact.
type ChatResponse struct {
	Reply     string    `json:"reply"`
	Artifact  *Artifact `json:"artifact,omitempty"`
	ToolNotes []string  `json:"tool_notes"`
}

// Service orchestrates provider chat, read-only tools, and per-session isolation.
type Service struct {
	provider Provider
	schema   SchemaReader
	commands CommandRunner

	mu       sync.Mutex
	inflight map[string]context.CancelFunc
}

func NewService(provider Provider, schema SchemaReader, commands CommandRunner) *Service {
	return &Service{
		provider: provider,
		schema:   schema,
		commands: commands,
		inflight: make(map[string]context.CancelFunc),
	}
}

func (s *Service) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	ctx, cancel := context.WithCancel(ctx)
	s.mu.Lock()
	if _, busy := s.inflight[req.SessionID]; busy {
		s.mu.Unlock()
		cancel()
		return nil, fmt.Errorf("已有生成进行中")
	}
	s.inflight[req.SessionID] = cancel
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.inflight, req.SessionID)
		s.mu.Unlock()
		cancel()
	}()

	tools := toolsForMode(req.Mode)
	messages := []Message{{Role: "system", Content: systemPrompt}}
	messages = append(messages, req.History...)
	messages = append(messages, Message{Role: "user", Content: buildUserPrompt(req)})

	var notes []string
	var last Message
	for round := 0; ; round++ {
		msg, err := s.provider.Chat(ctx, "", messages, tools)
		if err != nil {
			return nil, err
		}
		last = msg
		if len(msg.ToolCalls) == 0 || round >= MaxToolRounds {
			break
		}
		messages = append(messages, msg)
		for _, tc := range msg.ToolCalls {
			result, note := s.runTool(req.Mode, req.SessionID, tc.Name, tc.Arguments)
			result = truncateToolResult(result)
			if note != "" {
				notes = append(notes, note)
			}
			messages = append(messages, Message{
				Role:       "tool",
				ToolCallID: tc.ID,
				Content:    result,
			})
		}
	}

	reply := last.Content
	var artifact *Artifact
	if len(last.ToolCalls) == 0 {
		if art, ok := ParseArtifact(reply); ok {
			artifact = art
		}
	}
	if strings.TrimSpace(reply) == "" {
		reply = maxReplyFallbackText
	}
	if notes == nil {
		notes = []string{}
	}
	return &ChatResponse{Reply: reply, Artifact: artifact, ToolNotes: notes}, nil
}

func (s *Service) Cancel(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if c, ok := s.inflight[sessionID]; ok {
		c()
	}
}

func buildUserPrompt(req ChatRequest) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Host: %s\n", req.Host)
	fmt.Fprintf(&b, "User: %s\n", req.User)
	fmt.Fprintf(&b, "DBType: %s\n", req.DBType)
	fmt.Fprintf(&b, "Database: %s\n", req.Database)
	fmt.Fprintf(&b, "WorkingDir: %s\n", req.WorkingDir)
	fmt.Fprintf(&b, "EditorContent: %s\n", Redact(req.EditorContent))
	fmt.Fprintf(&b, "TerminalTail: %s\n", Redact(req.TerminalTail))
	b.WriteString("\n")
	b.WriteString(req.Message)
	return b.String()
}
