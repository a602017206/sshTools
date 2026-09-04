package copilot

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	"AHaSSHTools/internal/config"
)

type fakeProvider struct {
	mu      sync.Mutex
	calls   []providerCall
	handler func(call int, ctx context.Context, messages []Message, tools []ToolSpec) (Message, error)
}

type providerCall struct {
	model    string
	messages []Message
	tools    []ToolSpec
}

func (p *fakeProvider) Chat(ctx context.Context, model string, messages []Message, tools []ToolSpec) (Message, error) {
	p.mu.Lock()
	n := len(p.calls)
	p.calls = append(p.calls, providerCall{
		model:    model,
		messages: append([]Message(nil), messages...),
		tools:    append([]ToolSpec(nil), tools...),
	})
	h := p.handler
	p.mu.Unlock()
	if h == nil {
		return Message{Role: "assistant", Content: `{"type":"sql","content":"SELECT 1","summary":"ok"}`}, nil
	}
	return h(n, ctx, messages, tools)
}

func (p *fakeProvider) callCount() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.calls)
}

func (p *fakeProvider) firstTools() []ToolSpec {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.calls) == 0 {
		return nil
	}
	return p.calls[0].tools
}

func (p *fakeProvider) lastTools() []ToolSpec {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.calls) == 0 {
		return nil
	}
	return p.calls[len(p.calls)-1].tools
}

func (p *fakeProvider) allMessageText() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	var b strings.Builder
	for _, c := range p.calls {
		for _, m := range c.messages {
			b.WriteString(m.Content)
			b.WriteByte('\n')
		}
	}
	return b.String()
}

type fakeSchema struct {
	mu            sync.Mutex
	listDatabases []string
	listTables    []string
	tableSchema   *config.TableSchema
	listTablesN   int
	listDBN       int
	getSchemaN    int
}

func (f *fakeSchema) ListDatabases(sessionID string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.listDBN++
	return f.listDatabases, nil
}

func (f *fakeSchema) ListTables(sessionID string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.listTablesN++
	return f.listTables, nil
}

func (f *fakeSchema) GetTableSchema(sessionID, table string) (*config.TableSchema, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.getSchemaN++
	return f.tableSchema, nil
}

func (f *fakeSchema) listTablesCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.listTablesN
}

type fakeCommands struct {
	mu    sync.Mutex
	calls []string
}

func (f *fakeCommands) ExecuteCommand(sessionID, cmd string, timeout time.Duration) (string, string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls = append(f.calls, cmd)
	return "ok", "", nil
}

func (f *fakeCommands) callCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.calls)
}

func hasTool(tools []ToolSpec, name string) bool {
	for _, t := range tools {
		if t.Name == name {
			return true
		}
	}
	return false
}

func TestServiceDatabaseModeListTablesNotCommandRunner(t *testing.T) {
	schema := &fakeSchema{listTables: []string{"users", "orders"}}
	cmds := &fakeCommands{}
	provider := &fakeProvider{
		handler: func(call int, ctx context.Context, messages []Message, tools []ToolSpec) (Message, error) {
			if call == 0 {
				return Message{Role: "assistant", ToolCalls: []ToolCall{{
					ID:        "call_1",
					Name:      "list_tables",
					Arguments: "{}",
				}}}, nil
			}
			return Message{Role: "assistant", Content: `{"type":"sql","content":"SELECT id FROM users","summary":"列出用户"}`}, nil
		},
	}
	svc := NewService(provider, schema, cmds)
	resp, err := svc.Chat(context.Background(), ChatRequest{
		SessionID: "db-1",
		Mode:      "database",
		Message:   "列出用户",
		Host:      "db.example",
		User:      "alice",
		DBType:    "mysql",
		Database:  "shop",
	})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if schema.listTablesCalls() != 1 {
		t.Fatalf("ListTables calls = %d, want 1", schema.listTablesCalls())
	}
	if cmds.callCount() != 0 {
		t.Fatalf("CommandRunner must not be called in database mode, got %d", cmds.callCount())
	}
	if hasTool(provider.firstTools(), "ssh_probe") {
		t.Fatal("database mode must not register ssh_probe")
	}
	if !hasTool(provider.firstTools(), "list_tables") {
		t.Fatal("database mode must register list_tables")
	}
	if resp == nil || resp.Artifact == nil {
		t.Fatal("expected parsed SQL artifact")
	}
	if resp.Artifact.Type != "sql" || resp.Artifact.Content != "SELECT id FROM users" {
		t.Fatalf("unexpected artifact: %+v", resp.Artifact)
	}
}

func TestServiceSSHProbeRejectsUnsafeCommand(t *testing.T) {
	schema := &fakeSchema{}
	cmds := &fakeCommands{}
	provider := &fakeProvider{
		handler: func(call int, ctx context.Context, messages []Message, tools []ToolSpec) (Message, error) {
			if call == 0 {
				return Message{Role: "assistant", ToolCalls: []ToolCall{{
					ID:        "call_1",
					Name:      "ssh_probe",
					Arguments: `{"command":"rm -rf /"}`,
				}}}, nil
			}
			return Message{Role: "assistant", Content: `{"type":"shell","content":"uname","summary":"系统信息"}`}, nil
		},
	}
	svc := NewService(provider, schema, cmds)
	resp, err := svc.Chat(context.Background(), ChatRequest{
		SessionID:  "ssh-1",
		Mode:       "ssh",
		Message:    "看看系统",
		Host:       "host.example",
		User:       "alice",
		WorkingDir: "/home/alice",
	})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if cmds.callCount() != 0 {
		t.Fatalf("ExecuteCommand must not run rejected probe, got %d calls", cmds.callCount())
	}
	if schema.listTablesCalls() != 0 {
		t.Fatal("ssh mode must not call schema tools")
	}
	if hasTool(provider.firstTools(), "list_tables") || hasTool(provider.firstTools(), "list_databases") || hasTool(provider.firstTools(), "get_table_schema") {
		t.Fatal("ssh mode must not register schema tools")
	}
	if !hasTool(provider.firstTools(), "ssh_probe") {
		t.Fatal("ssh mode must register ssh_probe")
	}
	found := false
	for _, note := range resp.ToolNotes {
		if strings.Contains(note, "工具被拒绝") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("ToolNotes must contain 工具被拒绝, got %v", resp.ToolNotes)
	}
}

func TestServiceSSHWorkingDirectoryToolUsesRequestDirectory(t *testing.T) {
	cmds := &fakeCommands{}
	provider := &fakeProvider{
		handler: func(call int, ctx context.Context, messages []Message, tools []ToolSpec) (Message, error) {
			if call == 0 {
				if !hasTool(tools, "list_working_directory") {
					t.Fatal("ssh mode must register list_working_directory")
				}
				return Message{Role: "assistant", ToolCalls: []ToolCall{{ID: "dir_1", Name: "list_working_directory", Arguments: "{}"}}}, nil
			}
			return Message{Role: "assistant", Content: `{"type":"shell","content":"./restart.sh","summary":"重启服务"}`}, nil
		},
	}

	resp, err := NewService(provider, nil, cmds).Chat(context.Background(), ChatRequest{
		SessionID: "ssh-1", Mode: "ssh", Message: "重启当前目录服务", WorkingDir: "/srv/app",
	})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if resp.Artifact == nil || resp.Artifact.Content != "./restart.sh" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	cmds.mu.Lock()
	defer cmds.mu.Unlock()
	if got, want := cmds.calls, []string{"cd -- '/srv/app' && LC_ALL=C command ls -la --"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("commands = %#v, want %#v", got, want)
	}
}

func TestServiceChatRequestHasNoPasswordAndRedactsTerminalTail(t *testing.T) {
	rt := reflect.TypeOf(ChatRequest{})
	for i := 0; i < rt.NumField(); i++ {
		name := strings.ToLower(rt.Field(i).Name)
		if strings.Contains(name, "password") || strings.Contains(name, "secret") || strings.Contains(name, "apikey") {
			t.Fatalf("ChatRequest must not have secret field %s", rt.Field(i).Name)
		}
	}

	const secret = "s3cretValue99"
	provider := &fakeProvider{
		handler: func(call int, ctx context.Context, messages []Message, tools []ToolSpec) (Message, error) {
			return Message{Role: "assistant", Content: `{"type":"sql","content":"SELECT 1","summary":"ok"}`}, nil
		},
	}
	svc := NewService(provider, &fakeSchema{}, &fakeCommands{})
	_, err := svc.Chat(context.Background(), ChatRequest{
		SessionID:     "db-redact",
		Mode:          "database",
		Message:       "查询",
		Host:          "db.example",
		User:          "alice",
		DBType:        "mysql",
		Database:      "shop",
		WorkingDir:    "/home/alice",
		EditorContent: "SELECT 1",
		TerminalTail:  "login password=" + secret + " ok",
	})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	text := provider.allMessageText()
	if strings.Contains(text, secret) {
		t.Fatalf("provider messages must not contain redacted secret, got %q", text)
	}
	for _, want := range []string{"db.example", "alice", "mysql", "shop", "/home/alice"} {
		if !strings.Contains(text, want) {
			t.Fatalf("provider messages missing context %q in %q", want, text)
		}
	}
}

func TestServiceConcurrentChatSameSession(t *testing.T) {
	entered := make(chan struct{})
	unblock := make(chan struct{})
	provider := &fakeProvider{
		handler: func(call int, ctx context.Context, messages []Message, tools []ToolSpec) (Message, error) {
			if call == 0 {
				close(entered)
				select {
				case <-unblock:
				case <-ctx.Done():
					return Message{}, ctx.Err()
				}
			}
			return Message{Role: "assistant", Content: `{"type":"sql","content":"SELECT 1","summary":"ok"}`}, nil
		},
	}
	svc := NewService(provider, &fakeSchema{}, &fakeCommands{})
	req := ChatRequest{SessionID: "same-session", Mode: "database", Message: "hi"}

	firstErr := make(chan error, 1)
	go func() {
		_, err := svc.Chat(context.Background(), req)
		firstErr <- err
	}()

	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("first Chat did not reach provider")
	}

	_, err := svc.Chat(context.Background(), req)
	if err == nil || !strings.Contains(err.Error(), "已有生成进行中") {
		t.Fatalf("second Chat error = %v, want 已有生成进行中", err)
	}

	close(unblock)
	select {
	case err := <-firstErr:
		if err != nil {
			t.Fatalf("first Chat: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("first Chat did not finish")
	}
}

func TestChatRequestsFinalReplyAfterMaxToolRounds(t *testing.T) {
	// 前 MaxToolRounds 轮全部带 tool_calls（触顶），第 MaxToolRounds+1 次无 tools 调用应产出最终产物。
	toolCall := ToolCall{ID: "c1", Name: "list_tables", Arguments: "{}"}
	provider := &fakeProvider{
		handler: func(call int, ctx context.Context, messages []Message, tools []ToolSpec) (Message, error) {
			if call < MaxToolRounds {
				return Message{Role: "assistant", ToolCalls: []ToolCall{toolCall}}, nil
			}
			return Message{Role: "assistant", Content: `{"type":"sql","content":"SELECT 2","summary":"final","destructive":false}`}, nil
		},
	}
	schema := &fakeSchema{listTables: []string{"users"}}
	svc := NewService(provider, schema, nil)
	resp, err := svc.Chat(context.Background(), ChatRequest{SessionID: "s2", Mode: "database"})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if provider.callCount() != MaxToolRounds+1 {
		t.Fatalf("provider calls = %d, want %d (final no-tools reply)", provider.callCount(), MaxToolRounds+1)
	}
	// 最后一次调用必须不带 tools
	if lastTools := provider.lastTools(); len(lastTools) != 0 {
		t.Fatalf("final call tools = %d, want 0", len(lastTools))
	}
	if resp.Artifact == nil || resp.Artifact.Content != "SELECT 2" {
		t.Fatalf("expected final artifact SELECT 2, got %+v", resp.Artifact)
	}
	if resp.Reply != "final" && !strings.Contains(resp.Reply, "SELECT 2") {
		t.Fatalf("expected reply summary or content, got %q", resp.Reply)
	}
}

func TestRuntimeSettingsClampUnsafeValues(t *testing.T) {
	got := NormalizeRuntimeSettings(RuntimeSettings{MaxToolRounds: 99, MaxToolResultChars: 99})
	if got.MaxToolRounds != 8 || got.MaxToolResultChars != 1000 {
		t.Fatalf("unexpected normalized settings: %+v", got)
	}
}

func TestChatParsesArtifactWhenNoToolCalls(t *testing.T) {
	provider := &fakeProvider{
		handler: func(call int, ctx context.Context, messages []Message, tools []ToolSpec) (Message, error) {
			return Message{Role: "assistant", Content: `{"type":"sql","content":"SELECT 1","summary":"ping","destructive":false}`}, nil
		},
	}
	svc := NewService(provider, &fakeSchema{}, nil)
	resp, err := svc.Chat(context.Background(), ChatRequest{SessionID: "s1", Mode: "database"})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if resp.Artifact == nil || resp.Artifact.Content != "SELECT 1" {
		t.Fatalf("expected artifact SELECT 1, got %+v", resp.Artifact)
	}
}

func TestChatStopsWhenModelStopsCallingTools(t *testing.T) {
	// 模型第一轮就给最终产物，不应触发额外的无 tools 调用。
	provider := &fakeProvider{
		handler: func(call int, ctx context.Context, messages []Message, tools []ToolSpec) (Message, error) {
			return Message{Role: "assistant", Content: `{"type":"sql","content":"SELECT 3","summary":"done","destructive":false}`}, nil
		},
	}
	svc := NewService(provider, &fakeSchema{}, nil)
	resp, err := svc.Chat(context.Background(), ChatRequest{SessionID: "s3", Mode: "database"})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if provider.callCount() != 1 {
		t.Fatalf("provider calls = %d, want 1", provider.callCount())
	}
	if resp.Artifact == nil || resp.Artifact.Content != "SELECT 3" {
		t.Fatalf("expected artifact SELECT 3, got %+v", resp.Artifact)
	}
}

func TestServiceCancelStopsInFlightChat(t *testing.T) {
	entered := make(chan struct{})
	provider := &fakeProvider{
		handler: func(call int, ctx context.Context, messages []Message, tools []ToolSpec) (Message, error) {
			close(entered)
			<-ctx.Done()
			return Message{}, ctx.Err()
		},
	}
	svc := NewService(provider, &fakeSchema{}, &fakeCommands{})
	errCh := make(chan error, 1)
	go func() {
		_, err := svc.Chat(context.Background(), ChatRequest{
			SessionID: "cancel-1",
			Mode:      "database",
			Message:   "取消我",
		})
		errCh <- err
	}()
	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("Chat did not reach provider")
	}
	svc.Cancel("cancel-1")
	select {
	case err := <-errCh:
		if err == nil {
			t.Fatal("canceled Chat must return error")
		}
		if !errors.Is(err, context.Canceled) && !strings.Contains(err.Error(), "cancel") {
			t.Fatalf("canceled Chat error = %v, want context canceled", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("canceled Chat did not return")
	}
}

func TestServiceTruncatesToolResults(t *testing.T) {
	huge := strings.Repeat("t", MaxToolResultChars+200)
	schema := &fakeSchema{listTables: []string{huge}}
	provider := &fakeProvider{
		handler: func(call int, ctx context.Context, messages []Message, tools []ToolSpec) (Message, error) {
			if call == 0 {
				return Message{Role: "assistant", ToolCalls: []ToolCall{{
					ID:        "call_1",
					Name:      "list_tables",
					Arguments: "{}",
				}}}, nil
			}
			for _, m := range messages {
				if m.Role == "tool" || m.ToolCallID == "call_1" {
					if utf8.RuneCountInString(m.Content) > MaxToolResultChars {
						t.Errorf("tool result runes = %d, want <= %d", utf8.RuneCountInString(m.Content), MaxToolResultChars)
					}
				}
			}
			return Message{Role: "assistant", Content: `{"type":"sql","content":"SELECT 1","summary":"ok"}`}, nil
		},
	}
	svc := NewService(provider, schema, &fakeCommands{})
	if _, err := svc.Chat(context.Background(), ChatRequest{
		SessionID: "trunc-1",
		Mode:      "database",
		Message:   "表很多",
	}); err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if provider.callCount() < 2 {
		t.Fatal("expected a second provider round with truncated tool result")
	}
}

func TestBuildUserPromptIncludesOpenObject(t *testing.T) {
	prompt := buildUserPrompt(ChatRequest{
		Host:         "db.example",
		User:         "alice",
		DBType:       "postgresql",
		Database:     "shop",
		Schema:       "sales",
		ObjectKind:   "table",
		ObjectName:   "orders",
		ObjectParent: "",
		Message:      "给当前表加索引",
	})
	for _, want := range []string{"Schema: sales", "ObjectKind: table", "ObjectName: orders", "Database: shop", "给当前表加索引"} {
		if !strings.Contains(prompt, want) {
			t.Fatalf("prompt missing %q:\n%s", want, prompt)
		}
	}
}

type fakeScopedSchema struct {
	fakeSchema
	lastDatabase string
	lastSchema   string
	lastTable    string
}

func (f *fakeScopedSchema) ListTablesInScope(sessionID, database, schema string) ([]string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lastDatabase = database
	f.lastSchema = schema
	f.listTablesN++
	return f.listTables, nil
}

func (f *fakeScopedSchema) GetTableSchemaInScope(sessionID, database, schema, table string) (*config.TableSchema, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lastDatabase = database
	f.lastSchema = schema
	f.lastTable = table
	f.getSchemaN++
	return f.tableSchema, nil
}

func TestListTablesUsesRequestSchemaScope(t *testing.T) {
	schema := &fakeScopedSchema{fakeSchema: fakeSchema{listTables: []string{"orders"}}}
	provider := &fakeProvider{
		handler: func(call int, ctx context.Context, messages []Message, tools []ToolSpec) (Message, error) {
			if call == 0 {
				return Message{Role: "assistant", ToolCalls: []ToolCall{{
					ID: "c1", Name: "list_tables", Arguments: "{}",
				}}}, nil
			}
			return Message{Role: "assistant", Content: `{"type":"sql","content":"SELECT 1","summary":"ok"}`}, nil
		},
	}
	svc := NewService(provider, schema, nil)
	if _, err := svc.Chat(context.Background(), ChatRequest{
		SessionID: "db-1",
		Mode:      "database",
		DBType:    "postgresql",
		Database:  "shop",
		Schema:    "sales",
		Message:   "有哪些表",
	}); err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if schema.lastDatabase != "shop" || schema.lastSchema != "sales" {
		t.Fatalf("ListTablesInScope got database=%q schema=%q", schema.lastDatabase, schema.lastSchema)
	}
}

func TestGetTableSchemaDefaultsToOpenTable(t *testing.T) {
	schema := &fakeScopedSchema{fakeSchema: fakeSchema{tableSchema: &config.TableSchema{TableName: "orders"}}}
	provider := &fakeProvider{
		handler: func(call int, ctx context.Context, messages []Message, tools []ToolSpec) (Message, error) {
			if call == 0 {
				return Message{Role: "assistant", ToolCalls: []ToolCall{{
					ID: "c1", Name: "get_table_schema", Arguments: "{}",
				}}}, nil
			}
			return Message{Role: "assistant", Content: `{"type":"sql","content":"SELECT 1","summary":"ok"}`}, nil
		},
	}
	svc := NewService(provider, schema, nil)
	if _, err := svc.Chat(context.Background(), ChatRequest{
		SessionID:  "db-1",
		Mode:       "database",
		DBType:     "oracle",
		Database:   "ORCL",
		Schema:     "PEMS",
		ObjectName: "T_ORDER",
		Message:    "当前表结构",
	}); err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if schema.lastTable != "T_ORDER" || schema.lastSchema != "PEMS" || schema.lastDatabase != "ORCL" {
		t.Fatalf("GetTableSchemaInScope table=%q schema=%q database=%q", schema.lastTable, schema.lastSchema, schema.lastDatabase)
	}
}

type fakeNative struct {
	mu         sync.Mutex
	resources  []NativeResourceInfo
	children   []NativeResourceInfo
	details    *NativeResourceView
	lastParent string
	lastName   string
}

func (f *fakeNative) ListResources(sessionID string) ([]NativeResourceInfo, error) {
	return f.resources, nil
}

func (f *fakeNative) ListChildResources(sessionID, parent string) ([]NativeResourceInfo, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lastParent = parent
	return f.children, nil
}

func (f *fakeNative) DescribeResource(sessionID, parent, name string) (*NativeResourceView, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lastParent = parent
	f.lastName = name
	return f.details, nil
}

func TestNativeModeUsesNativeToolsNotSchemaTools(t *testing.T) {
	native := &fakeNative{details: &NativeResourceView{Name: "logs-2026", Kind: "index"}}
	provider := &fakeProvider{
		handler: func(call int, ctx context.Context, messages []Message, tools []ToolSpec) (Message, error) {
			if call == 0 {
				return Message{Role: "assistant", ToolCalls: []ToolCall{{
					ID: "c1", Name: "describe_resource", Arguments: "{}",
				}}}, nil
			}
			return Message{Role: "assistant", Content: `{"type":"sql","content":"{}","summary":"ok"}`}, nil
		},
	}
	svc := NewService(provider, &fakeSchema{}, nil).WithNative(native)
	if _, err := svc.Chat(context.Background(), ChatRequest{
		SessionID:    "es-1",
		Mode:         "database",
		DBType:       "elasticsearch",
		ObjectKind:   "index",
		ObjectName:   "logs-2026",
		ObjectParent: "",
		Message:      "当前索引 mapping",
	}); err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if hasTool(provider.firstTools(), "list_tables") || hasTool(provider.firstTools(), "get_table_schema") {
		t.Fatal("native database types must not register JDBC schema tools")
	}
	if !hasTool(provider.firstTools(), "describe_resource") {
		t.Fatal("native database types must register describe_resource")
	}
	if native.lastName != "logs-2026" {
		t.Fatalf("DescribeResource name = %q, want logs-2026", native.lastName)
	}
}

func TestServiceChatPassesModel(t *testing.T) {
	provider := &fakeProvider{}
	svc := NewService(provider, &fakeSchema{}, &fakeCommands{})
	_, err := svc.Chat(context.Background(), ChatRequest{
		SessionID: "model-1",
		Mode:      "database",
		Message:   "hi",
		Model:     "deepseek-chat",
	})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if provider.callCount() < 1 {
		t.Fatal("expected provider.Chat to be called")
	}
	provider.mu.Lock()
	got := provider.calls[0].model
	provider.mu.Unlock()
	if got != "deepseek-chat" {
		t.Fatalf("provider model = %q, want deepseek-chat", got)
	}
}
