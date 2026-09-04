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

const systemPrompt = `你是面向当前 SSH 或数据库会话的 AI 运维助手。先理解用户目标与会话上下文；存在只读工具时优先用工具验证，不要猜测目录中的脚本、表结构或运行环境。用户消息会附带 Database、Schema、ObjectKind、ObjectName、ObjectParent、EditorContent，表示此刻打开的库、表、索引或键；生成 SQL、DSL 或命令时优先针对该对象，不要让用户重复指出。对于服务启动、停止、重启等请求，先检查当前工作目录中的候选脚本，并说明选择依据。工具失败或信息不足时，明确缺少什么信息并给出安全下一步。只生成用户可审阅的产物，绝不自动执行；危险操作必须标记 destructive=true。不得索要、回显或推断密码、私钥、API Key、Token、DSN。最终回复必须输出 JSON artifact（见当前模式说明）。`

const systemPromptJDBC = systemPrompt + `当前为 JDBC/SQL 模式。最终回复必须输出：{"type":"sql","content":"...","summary":"...","destructive":false}。`

const systemPromptSSH = systemPrompt + `当前为 SSH/Shell 模式。最终回复必须输出：{"type":"shell","content":"...","summary":"...","destructive":false}。`

const systemPromptNative = systemPrompt + `当前为缓存/搜索等原生数据源（Redis、Elasticsearch 等）。不要输出 type=sql。可执行查询必须放在最终 JSON artifact 的 content 中：{"type":"native_query","content":{...DSL 或命令...},"summary":"...","destructive":false}。写入/删除用 {"type":"native_mutation","content":"{\"operation\":\"...\",\"parent\":\"\",\"name\":\"\",\"payload\":\"{}\"}","summary":"...","destructive":true}。说明用自然语言即可；不要在 artifact 之外再贴一份无 type 的查询 JSON，也不要把 type/summary/destructive 字段当作查询语句的一部分。Elasticsearch 日期可用 date math（如 now-7d），数值型时间戳字段需配合 format（如 epoch_millis）。`

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

// ScopedSchemaReader lists tables and columns in the catalog/schema the user currently has open.
type ScopedSchemaReader interface {
	SchemaReader
	ListTablesInScope(sessionID, database, schema string) ([]string, error)
	GetTableSchemaInScope(sessionID, database, schema, table string) (*config.TableSchema, error)
}

// NativeReader loads cache/search/resource metadata for copilot tools.
type NativeReader interface {
	ListResources(sessionID string) ([]NativeResourceInfo, error)
	ListChildResources(sessionID, parent string) ([]NativeResourceInfo, error)
	DescribeResource(sessionID, parent, name string) (*NativeResourceView, error)
}

// NativeQuerier runs read-only native queries for copilot tools.
type NativeQuerier interface {
	ExecuteQuery(sessionID, parent, name, query string) (string, error)
}

type NativeResourceInfo struct {
	Kind string `json:"kind"`
	Name string `json:"name"`
}

type NativeResourceView struct {
	Kind    string `json:"kind"`
	Name    string `json:"name"`
	Summary string `json:"summary"`
	Content string `json:"content"`
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
	Schema        string
	ObjectKind    string
	ObjectName    string
	ObjectParent  string
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
	native   NativeReader
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
		native:   s.native,
		commands: s.commands,
		gate:     s.gate,
		runtime:  s.runtime,
	}
}

func (s *Service) WithNative(n NativeReader) *Service {
	if s == nil {
		return &Service{
			native: n,
			gate:   &sessionGate{inflight: make(map[string]context.CancelFunc)},
		}
	}
	clone := *s
	clone.native = n
	return &clone
}

func (s *Service) WithRuntimeSettings(v RuntimeSettings) *Service {
	clone := *s
	clone.runtime = NormalizeRuntimeSettings(v)
	return &clone
}

func (s *Service) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	ctx, finish, err := s.beginChat(ctx, req.SessionID)
	if err != nil {
		return nil, err
	}
	defer finish()

	messages := initialMessages(req)
	last, notes, messages, err := s.runChatRounds(ctx, req, messages)
	if err != nil {
		return nil, err
	}
	if len(last.ToolCalls) > 0 {
		last, err = s.requestFinalReply(ctx, req.Model, messages, last)
		if err != nil {
			return nil, err
		}
	}
	return buildChatResponse(req, last, notes), nil
}

func (s *Service) beginChat(ctx context.Context, sessionID string) (context.Context, func(), error) {
	ctx, cancel := context.WithCancel(ctx)
	s.gate.mu.Lock()
	if _, busy := s.gate.inflight[sessionID]; busy {
		s.gate.mu.Unlock()
		cancel()
		return nil, nil, fmt.Errorf("已有生成进行中")
	}
	s.gate.inflight[sessionID] = cancel
	s.gate.mu.Unlock()
	return ctx, func() {
		s.gate.mu.Lock()
		delete(s.gate.inflight, sessionID)
		s.gate.mu.Unlock()
		cancel()
	}, nil
}

func initialMessages(req ChatRequest) []Message {
	messages := []Message{{Role: "system", Content: systemPromptForRequest(req)}}
	messages = append(messages, req.History...)
	messages = append(messages, Message{Role: "user", Content: buildUserPrompt(req)})
	return messages
}

func systemPromptForRequest(req ChatRequest) string {
	if strings.EqualFold(strings.TrimSpace(req.Mode), "ssh") {
		return systemPromptSSH
	}
	if IsNativeDBType(req.DBType) {
		return systemPromptNative
	}
	return systemPromptJDBC
}

func (s *Service) runChatRounds(ctx context.Context, req ChatRequest, messages []Message) (Message, []string, []Message, error) {
	tools := toolsForRequest(req)
	var notes []string
	var last Message
	for round := 0; round < s.runtime.MaxToolRounds; round++ {
		msg, err := s.provider.Chat(ctx, req.Model, messages, tools)
		if err != nil {
			return Message{}, nil, nil, err
		}
		last = msg
		if len(msg.ToolCalls) == 0 {
			break
		}
		messages = append(messages, msg)
		for _, tc := range msg.ToolCalls {
			result, note := s.runTool(req, tc.Name, tc.Arguments)
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
	return last, notes, messages, nil
}

func (s *Service) requestFinalReply(ctx context.Context, model string, messages []Message, last Message) (Message, error) {
	messages = append(messages, last)
	finalMsg, err := s.provider.Chat(ctx, model, messages, nil)
	if err != nil {
		return Message{}, err
	}
	return finalMsg, nil
}

func buildChatResponse(req ChatRequest, last Message, notes []string) *ChatResponse {
	reply := last.Content
	var artifact *Artifact
	if len(last.ToolCalls) == 0 {
		artifact, reply = ExtractArtifact(reply, req.DBType)
	}
	if strings.TrimSpace(reply) == "" {
		if artifact != nil && strings.TrimSpace(artifact.Summary) != "" {
			reply = artifact.Summary
		} else {
			reply = maxReplyFallbackText
		}
	}
	if notes == nil {
		notes = []string{}
	}
	return &ChatResponse{Reply: reply, Artifact: artifact, ToolNotes: notes}
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
	fmt.Fprintf(&b, "Schema: %s\n", req.Schema)
	fmt.Fprintf(&b, "ObjectKind: %s\n", req.ObjectKind)
	fmt.Fprintf(&b, "ObjectName: %s\n", req.ObjectName)
	fmt.Fprintf(&b, "ObjectParent: %s\n", req.ObjectParent)
	fmt.Fprintf(&b, "WorkingDir: %s\n", req.WorkingDir)
	fmt.Fprintf(&b, "EditorContent: %s\n", Redact(req.EditorContent))
	fmt.Fprintf(&b, "TerminalTail: %s\n", Redact(req.TerminalTail))
	b.WriteString("\n")
	b.WriteString(req.Message)
	return b.String()
}
