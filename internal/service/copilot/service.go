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

const systemPrompt = `你是面向当前 SSH 或数据库会话的 AI 运维助手。先理解用户目标与会话上下文；存在只读工具时优先用工具验证，不要猜测目录中的脚本、表结构或运行环境。对于服务启动、停止、重启等请求，先检查当前工作目录中的候选脚本，并说明选择依据。工具失败或信息不足时，明确缺少什么信息并给出安全下一步。只生成用户可审阅的 SQL 或 Shell 产物，绝不自动执行；危险操作必须标记 destructive=true。不得索要、回显或推断密码、私钥、API Key、Token、DSN。最终回复必须输出 JSON：{"type":"sql 或 shell","content":"...","summary":"...","destructive":false}。`

type RuntimeSettings struct{ MaxToolRounds, MaxToolResultChars int }

func NormalizeRuntimeSettings(v RuntimeSettings) RuntimeSettings {
	if v.MaxToolRounds < 1 {
		v.MaxToolRounds = MaxToolRounds
	}
	if v.MaxToolRounds > 8 {
		v.MaxToolRounds = 8
	}
	if v.MaxToolResultChars < 1000 {
		v.MaxToolResultChars = 1000
	}
	if v.MaxToolResultChars > 20000 {
		v.MaxToolResultChars = 20000
	}
	return v
}

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
	Model         string
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

type sessionGate struct {
	mu       sync.Mutex
	inflight map[string]context.CancelFunc
}

// Service orchestrates provider chat, read-only tools, and per-session isolation.
type Service struct {
	provider Provider
	schema   SchemaReader
	commands CommandRunner
	gate     *sessionGate
	runtime  RuntimeSettings
}

func NewService(provider Provider, schema SchemaReader, commands CommandRunner) *Service {
	return &Service{
		provider: provider,
		schema:   schema,
		commands: commands,
		gate:     &sessionGate{inflight: make(map[string]context.CancelFunc)},
		runtime:  NormalizeRuntimeSettings(RuntimeSettings{MaxToolRounds: MaxToolRounds, MaxToolResultChars: MaxToolResultChars}),
	}
}

// WithProvider returns a shallow copy that shares schema, commands, and in-flight sessions.
func (s *Service) WithProvider(p Provider) *Service {
	if s == nil {
		return &Service{
			provider: p,
			gate:     &sessionGate{inflight: make(map[string]context.CancelFunc)},
		}
	}
	return &Service{
		provider: p,
		schema:   s.schema,
		commands: s.commands,
		gate:     s.gate,
		runtime:  s.runtime,
	}
}

func (s *Service) WithRuntimeSettings(v RuntimeSettings) *Service {
	clone := *s
	clone.runtime = NormalizeRuntimeSettings(v)
	return &clone
}

func (s *Service) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	ctx, cancel := context.WithCancel(ctx)
	s.gate.mu.Lock()
	if _, busy := s.gate.inflight[req.SessionID]; busy {
		s.gate.mu.Unlock()
		cancel()
		return nil, fmt.Errorf("已有生成进行中")
	}
	s.gate.inflight[req.SessionID] = cancel
	s.gate.mu.Unlock()
	defer func() {
		s.gate.mu.Lock()
		delete(s.gate.inflight, req.SessionID)
		s.gate.mu.Unlock()
		cancel()
	}()

	tools := toolsForMode(req.Mode)
	messages := []Message{{Role: "system", Content: systemPrompt}}
	messages = append(messages, req.History...)
	messages = append(messages, Message{Role: "user", Content: buildUserPrompt(req)})

	var notes []string
	var last Message
	for round := 0; round < s.runtime.MaxToolRounds; round++ {
		msg, err := s.provider.Chat(ctx, req.Model, messages, tools)
		if err != nil {
			return nil, err
		}
		last = msg
		if len(msg.ToolCalls) == 0 {
			break
		}
		messages = append(messages, msg)
		for _, tc := range msg.ToolCalls {
			result, note := s.runTool(req.Mode, req.SessionID, req.WorkingDir, tc.Name, tc.Arguments)
			result = truncateToolResultTo(result, s.runtime.MaxToolResultChars)
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

	// 触顶后仍带 tool_calls：再请求一次不带工具的 Chat，让模型基于已有工具结果产出最终产物。
	if len(last.ToolCalls) > 0 {
		messages = append(messages, last)
		finalMsg, err := s.provider.Chat(ctx, req.Model, messages, nil)
		if err != nil {
			return nil, err
		}
		last = finalMsg
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
	if s == nil || s.gate == nil {
		return
	}
	s.gate.mu.Lock()
	defer s.gate.mu.Unlock()
	if c, ok := s.gate.inflight[sessionID]; ok {
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
