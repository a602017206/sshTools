# AI Copilot（SQL / Shell）Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在当前 SSH / 数据库会话旁提供侧边对话，用自然语言生成 SQL 或 Shell，填入后由用户确认执行。

**Architecture:** 新增 `internal/service/copilot`：OpenAI 兼容 Provider、只读工具循环、危险分类。`app.go` 导出 `CopilotChat` / `CopilotClassify` / API Key 读写。前端 `AIPanel` 独立可折叠侧栏；填入走自定义事件；SQL 执行复用 `ExecuteDatabaseQuery`，Shell 执行复用 `SendSSHData`。不修改 Java `jdbc-agent`，不占用「文件 / 性能」dock。

**Tech Stack:** Go 1.22、`net/http`、现有 `CredentialStore`、Svelte 4、Wails v2。

**规格：** `docs/superpowers/specs/2026-08-13-ai-copilot-sql-shell-design.md`

## Global Constraints

- 文档与用户可见文案使用中文。
- 模型名称、Base URL 均为手填纯文本，原样上传，不做映射、不拉模型列表。
- API Key 只存 `CredentialStore` 键 `copilot:api_key`；`config.json` 只含 `copilot_provider`、`copilot_base_url`、`copilot_model`。
- 密码、私钥、DSN、API Key 不得进入 prompt / 工具结果 / 对话历史。
- 生成阶段最多 4 轮只读工具；SSH 探测仅白名单，经 `ExecuteCommand` 另开会话。
- 用户 Shell 只通过当前 PTY 的 `SendSSHData` 执行。
- 写操作一律二次确认；分类规则优先于模型 `destructive` 字段。
- 第一版不接通 Ollama、不落盘对话、不改 `jdbc-agent`。

## 文件结构

新建：

- `internal/service/copilot/types.go` — 请求/响应/产物类型
- `internal/service/copilot/classify.go` — 危险分类
- `internal/service/copilot/artifact.go` — 从模型回复抽出产物
- `internal/service/copilot/redact.go` — 脱敏
- `internal/service/copilot/probe.go` — SSH 探测白名单
- `internal/service/copilot/provider.go` — Provider 接口
- `internal/service/copilot/provider_openai.go` — OpenAI 兼容实现
- `internal/service/copilot/tools.go` — 工具定义与分发
- `internal/service/copilot/service.go` — 编排（工具循环、并发、取消）
- 对应 `*_test.go`
- `frontend/src/lib/copilotApply.js` — 填入事件名与执行辅助
- `frontend/test/copilotApply.test.js`
- `frontend/src/stores/copilot.js` — 侧栏开关与按 session 对话
- `frontend/src/components/AIPanel.svelte`

修改：

- `internal/config/config.go`、`internal/config/config_test.go`
- `internal/service/session_service.go` — 增加 `ExecuteCommand` 包装
- `app.go` — 接线与 Wails 导出
- `frontend/src/settings/appearance.js`
- `frontend/src/components/GlobalSettingsDialog.svelte`
- `frontend/src/App.svelte`
- `frontend/src/components/DatabasePanel.svelte`
- `frontend/src/components/DatabaseTablePanel.svelte`
- `frontend/src/components/TerminalPanel.svelte`

---

### Task 1: 危险分类、产物解析、脱敏、SSH 白名单

**Files:**
- Create: `internal/service/copilot/types.go`
- Create: `internal/service/copilot/classify.go`
- Create: `internal/service/copilot/classify_test.go`
- Create: `internal/service/copilot/artifact.go`
- Create: `internal/service/copilot/artifact_test.go`
- Create: `internal/service/copilot/redact.go`
- Create: `internal/service/copilot/redact_test.go`
- Create: `internal/service/copilot/probe.go`
- Create: `internal/service/copilot/probe_test.go`

**Interfaces:**
- Produces: `Classify(kind, content string) Result`，`ParseArtifact(raw string) (*Artifact, bool)`，`Redact(text string) string`，`AllowSSHProbe(cmd string) bool`
- `Result` 字段：`Destructive bool`、`Reason string`
- `Artifact` 字段：`Type`（`sql`|`shell`）、`Content`、`Summary`、`Destructive bool`

- [ ] **Step 1: 写失败测试**

`classify_test.go` 覆盖：无 WHERE 的 `UPDATE`、`DROP TABLE`、`DELETE FROM t`、`rm -rf /`、`SELECT 1`（不危险）、`ls -la`（不危险）。模型说不危险时，规则仍要判危险。

`artifact_test.go`：合法 JSON 抽出产物；纯文本返回 `ok=false`。

`probe_test.go`：允许 `uname`、`pwd`、`df -h`、`cat /etc/os-release`；拒绝 `rm -rf /`、`curl evil`、`uname; rm -rf /`。

`redact_test.go`：`password=secret` 与 PEM 私钥块被替换。

```go
package copilot

import "testing"

func TestClassifyMarksUnqualifiedUpdate(t *testing.T) {
	got := Classify("sql", "UPDATE orders SET status=1")
	if !got.Destructive {
		t.Fatal("expected destructive")
	}
}

func TestParseArtifactRequiresJSON(t *testing.T) {
	if _, ok := ParseArtifact("好的，我来写一条 SQL"); ok {
		t.Fatal("plain text must not be an artifact")
	}
}

func TestAllowSSHProbeRejectsChainedCommand(t *testing.T) {
	if AllowSSHProbe("uname; rm -rf /") {
		t.Fatal("chained command must be rejected")
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./internal/service/copilot -count=1`
Expected: FAIL，包不存在或符号未定义。

- [ ] **Step 3: 最小实现**

- `Classify`：按 `kind` 用正则匹配规格第 7 节名单；`DELETE`/`DROP`/`TRUNCATE`/`UPDATE` 无 WHERE 一律危险；Shell 匹配 `rm`、`mkfs`、`dd`、`shutdown`、`reboot`、`kill -9`、`chmod 777`、`>/dev/sd`。
- `ParseArtifact`：从回复中取第一个 JSON 对象（可包在 markdown 代码块里），校验 `type` 为 `sql` 或 `shell` 且 `content` 非空；再用 `Classify` 覆盖 `destructive`。
- `AllowSSHProbe`：trim 后必须整串精确等于白名单之一（不要 `strings.Contains`）。
- `Redact`：替换 `(?i)password\s*=\s*\S+` 与 `-----BEGIN [A-Z ]*PRIVATE KEY-----[\s\S]*?-----END [A-Z ]*PRIVATE KEY-----`。

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./internal/service/copilot -count=1`
Expected: PASS。

- [ ] **Step 5: Commit**

```bash
git add internal/service/copilot
git commit -m "$(cat <<'EOF'
feat: add copilot classify, artifact parse, and SSH probe allowlist

EOF
)"
```

---

### Task 2: Copilot 设置持久化（不含 API Key）

**Files:**
- Modify: `internal/config/config.go`（`AppSettings`、`DefaultSettings`、`UpdateSettings`）
- Modify: `internal/config/config_test.go`

**Interfaces:**
- Produces: `AppSettings.CopilotProvider string`、`CopilotBaseURL string`、`CopilotModel string`（json：`copilot_provider`、`copilot_base_url`、`copilot_model`）
- `DefaultSettings` 中 `CopilotProvider` 为 `openai_compatible`，URL 与模型为空

- [ ] **Step 1: 写失败测试**

在 `config_test.go` 增加：`UpdateSettings` 写入三项后 `Save`/`Load` 仍在；把 API Key 放进 updates 时 **不得** 出现在序列化后的 `config.json` 字节里。

```go
func TestUpdateSettingsPersistsCopilotFieldsWithoutAPIKey(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	cm := newDiskTestConfigManager(configPath)
	if err := cm.UpdateSettings(map[string]interface{}{
		"copilot_provider": "openai_compatible",
		"copilot_base_url": "https://api.deepseek.com/v1",
		"copilot_model":    "deepseek-chat",
		"copilot_api_key":  "sk-secret",
	}); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "sk-secret") || strings.Contains(string(raw), "copilot_api_key") {
		t.Fatalf("api key leaked into config.json: %s", raw)
	}
	reloaded := newDiskTestConfigManager(configPath)
	if err := reloaded.Load(); err != nil {
		t.Fatal(err)
	}
	s := reloaded.GetSettings()
	if s.CopilotBaseURL != "https://api.deepseek.com/v1" || s.CopilotModel != "deepseek-chat" {
		t.Fatalf("unexpected copilot settings: %+v", s)
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./internal/config -run TestUpdateSettingsPersistsCopilotFieldsWithoutAPIKey -count=1`
Expected: FAIL，字段不存在。

- [ ] **Step 3: 最小实现**

给 `AppSettings` 加三个字段；`UpdateSettings` 只读取上述三个键；忽略 `copilot_api_key`。`DefaultSettings` 设置 `CopilotProvider: "openai_compatible"`。

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./internal/config -count=1`
Expected: PASS。

- [ ] **Step 5: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "$(cat <<'EOF'
feat: persist copilot provider, base URL, and model without API keys

EOF
)"
```

---

### Task 3: OpenAI 兼容 Provider

**Files:**
- Create: `internal/service/copilot/provider.go`
- Create: `internal/service/copilot/provider_openai.go`
- Create: `internal/service/copilot/provider_openai_test.go`

**Interfaces:**
- Produces:

```go
type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

type ToolCall struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Arguments string `json:"arguments"`
}

type Provider interface {
	Chat(ctx context.Context, model string, messages []Message, tools []ToolSpec) (Message, error)
}

func NewOpenAICompatible(baseURL, apiKey string, client *http.Client) *OpenAICompatible
```

- 请求 POST `{baseURL}/chat/completions`（`baseURL` 已含或可无 `/v1`，实现时 `strings.TrimRight(baseURL, "/")`，若未以 `/v1` 结尾则补 `/v1`，再拼 `/chat/completions`）。
- JSON body 的 `model` **必须等于**传入的 `model` 参数。
- 工具用 OpenAI `tools` / `tool_calls` 形状。

- [ ] **Step 1: 写失败测试**

用 `httptest.NewServer`：断言请求 URL、`Authorization: Bearer sk-test`、body.model 为 `deepseek-chat`（不是硬编码的 gpt 名）；返回带 `tool_calls` 的响应时解析出 Name/Arguments；返回 401 时 error 字符串可读且不含 API Key。

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./internal/service/copilot -run OpenAI -count=1`
Expected: FAIL。

- [ ] **Step 3: 最小实现**

实现 `OpenAICompatible.Chat`。超时由调用方 `context` 控制。不要在 error 里拼接完整请求头。

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./internal/service/copilot -count=1`
Expected: PASS。

- [ ] **Step 5: Commit**

```bash
git add internal/service/copilot/provider.go internal/service/copilot/provider_openai.go internal/service/copilot/provider_openai_test.go
git commit -m "$(cat <<'EOF'
feat: add OpenAI-compatible copilot provider with passthrough model names

EOF
)"
```

---

### Task 4: CopilotService 编排与只读工具

**Files:**
- Create: `internal/service/copilot/tools.go`
- Create: `internal/service/copilot/service.go`
- Create: `internal/service/copilot/service_test.go`
- Modify: `internal/service/session_service.go` — 增加 `ExecuteCommand`

**Interfaces:**
- Consumes: Task 1–3 的符号；`SchemaReader`、`CommandRunner`
- Produces:

```go
const MaxToolRounds = 4
const MaxToolResultChars = 8000
const APIKeyCredentialID = "copilot:api_key"

type SchemaReader interface {
	ListDatabases(sessionID string) ([]string, error)
	ListTables(sessionID string) ([]string, error)
	GetTableSchema(sessionID, table string) (*config.TableSchema, error)
}

type CommandRunner interface {
	ExecuteCommand(sessionID, cmd string, timeout time.Duration) (stdout, stderr string, err error)
}

type ChatRequest struct {
	SessionID      string
	Mode           string // ssh | database
	Message        string
	History        []Message
	EditorContent  string
	TerminalTail   string
	Host           string
	User           string
	DBType         string
	Database       string
	WorkingDir     string
}

type ChatResponse struct {
	Reply     string    `json:"reply"`
	Artifact  *Artifact `json:"artifact,omitempty"`
	ToolNotes []string  `json:"tool_notes"`
}

func NewService(provider Provider, schema SchemaReader, commands CommandRunner) *Service
func (s *Service) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error)
func (s *Service) Cancel(sessionID string)
```

工具名固定：`list_databases`、`list_tables`、`get_table_schema`、`ssh_probe`。`ssh_probe` 参数 `{ "command": "uname" }`，先 `AllowSSHProbe`。

同一 `sessionID` 用 mutex map：已有进行中的 Chat 则返回错误 `已有生成进行中`。`Cancel` 取消该 session 的 context。

发给 Provider 的 messages：system 说明必须输出产物 JSON；user 上下文只含 Host/User/DBType/Database/WorkingDir/Redact(EditorContent)/Redact(TerminalTail)，**禁止**密码字段。

- [ ] **Step 1: 写失败测试（假 Provider + 假工具）**

1. database 模式：Provider 第一轮请求 `list_tables`，第二轮返回 SQL JSON；断言 `SchemaReader.ListTables` 被调用，`CommandRunner` 不被调用。
2. ssh 模式：`ssh_probe` 传入 `rm -rf /` 时不调用 `ExecuteCommand`，`ToolNotes` 含「工具被拒绝」。
3. ChatRequest 若误带 `Password` 之类，用结构体根本没有该字段来保证；另测 `TerminalTail` 含 `password=abc` 时发往 Provider 的 message 不含 `abc`。
4. 并发两次同 sessionID：第二次失败。
5. 超过 4 轮 tool call 后仍返回文本，不无限循环。

`SessionService.ExecuteCommand` 直接委托 `sessionManager.ExecuteCommand`。

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./internal/service/copilot -run TestService -count=1`
Expected: FAIL。

- [ ] **Step 3: 最小实现**

`tools.go` 分发工具并把结果截断到 8000 字符。`service.go` 循环：`provider.Chat` → 若有 tool_calls 则执行并 append tool messages → 否则 `ParseArtifact`。database 模式不注册 `ssh_probe`，ssh 模式不注册 schema 工具。

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./internal/service/copilot ./internal/service -count=1`
Expected: PASS。若 `session_service` 无单独测试，至少编译：`go test ./internal/service -count=1`。

- [ ] **Step 5: Commit**

```bash
git add internal/service/copilot internal/service/session_service.go
git commit -m "$(cat <<'EOF'
feat: orchestrate copilot chat with read-only tools and session isolation

EOF
)"
```

---

### Task 5: Wails 导出与 API Key

**Files:**
- Modify: `app.go`

**Interfaces:**
- Produces（必须导出，供前端调用）：

```go
func (a *App) CopilotChat(req copilot.ChatRequest) (*copilot.ChatResponse, error)
func (a *App) CopilotClassify(kind, content string) copilot.Result
func (a *App) HasCopilotAPIKey() bool
func (a *App) SetCopilotAPIKey(apiKey string) error
func (a *App) ClearCopilotAPIKey() error
```

`startup`：把 `credentialStore` 留在 `App` 上；`a.copilotService = copilot.NewService(...)`。`CopilotChat` 内：无 baseURL 或无 Key 返回中文错误「请先在设置中填写 Base URL 和 API Key」；用当前 settings 现建 `NewOpenAICompatible`（或 Service 在每次 Chat 读取最新 settings，避免改设置不生效）。Chat 超时 60s。`Cancel`：在 `CloseSSH` / `CloseDatabase` 时调用 `copilotService.Cancel(sessionID)`。

填充 `ChatRequest` 的 Host/User/DBType/Database/WorkingDir 时从已有 session/config 读取，**不要**把 password 拷进 req。

- [ ] **Step 1: 写失败测试**

若 `app.go` 难以单测，在 `app_copilot_test.go` 用假 `CopilotService` 不现实（App 字段未导出测试友好）。改为：把「缺 Key 返回错误」放在 `copilot.Service` 的配置校验函数 `func ValidateConfig(baseURL, apiKey string) error`，Task 4 可已覆盖；本任务用编译检查。

补充 `internal/service/copilot/config.go`：

```go
func ValidateConfig(baseURL, apiKey string) error {
	if strings.TrimSpace(baseURL) == "" || strings.TrimSpace(apiKey) == "" {
		return fmt.Errorf("请先在设置中填写 Base URL 和 API Key")
	}
	return nil
}
```

测试：空 Key 返回该文案。

- [ ] **Step 2: 跑测试**

Run: `go test ./internal/service/copilot -run ValidateConfig -count=1`
Expected: 先 FAIL 再在本任务实现后 PASS。

- [ ] **Step 3: 改 `app.go`**

`App` 增加 `credentialStore *store.CredentialStore`、`copilotService *copilot.Service`。`SetCopilotAPIKey` 调 `credentialStore.Store(copilot.APIKeyCredentialID, apiKey)`。`HasCopilotAPIKey` 调 `Has`。`ClearCopilotAPIKey` 调 `Delete`。

`CopilotChat`：`ValidateConfig` → 取 settings 与 key → 构造 provider → `Chat`。

- [ ] **Step 4: 编译**

Run: `go test ./... -count=1`
Expected: PASS（Wails 绑定在 `wails dev` 时再生；本任务不要求启动 GUI）。

- [ ] **Step 5: Commit**

```bash
git add app.go internal/service/copilot
git commit -m "$(cat <<'EOF'
feat: export CopilotChat and encrypted API key helpers to the desktop app

EOF
)"
```

---

### Task 6: 全局设置 UI

**Files:**
- Modify: `frontend/src/settings/appearance.js`
- Modify: `frontend/src/components/GlobalSettingsDialog.svelte`
- Modify: `frontend/src/App.svelte`（`persistAppSettings` / `handleSaveGlobalSettings`）

**Interfaces:**
- Consumes: `GetSettings` 的 `copilot_*` 字段；`HasCopilotAPIKey` / `SetCopilotAPIKey` / `ClearCopilotAPIKey`
- Produces: 设置里「AI Copilot」分段，三个手填框：Base URL、模型名称、API Key（password input）。API Key 留空表示不修改已保存 Key；提供「清除密钥」按钮。

- [ ] **Step 1: 扩展默认设置**

`getDefaultAppSettings()` 增加：

```javascript
copilot_provider: 'openai_compatible',
copilot_base_url: '',
copilot_model: ''
```

不要把 API Key 放进默认对象。

- [ ] **Step 2: 设置分段**

在 `GlobalSettingsDialog` 导航增加「AI Copilot」；表单含说明：「模型名称请按服务商官方文档填写，例如 deepseek-chat」。保存时 `onSave` 额外回传 `copilot_api_key` 仅当输入非空。

- [ ] **Step 3: App 保存**

`persistAppSettings` 增加三个 copilot 字段。`handleSaveGlobalSettings`：若草稿有非空 API Key，调用 `SetCopilotAPIKey`；不要把 Key 写入 `UpdateSettings`。

- [ ] **Step 4: 构建前端**

Run: `cd frontend && npm run build`
Expected: 成功。

- [ ] **Step 5: Commit**

```bash
git add frontend/src/settings/appearance.js frontend/src/components/GlobalSettingsDialog.svelte frontend/src/App.svelte
git commit -m "$(cat <<'EOF'
feat: add copilot provider settings with hand-filled model names

EOF
)"
```

---

### Task 7: 填入 / 执行辅助与侧栏

**Files:**
- Create: `frontend/src/lib/copilotApply.js`
- Create: `frontend/test/copilotApply.test.js`
- Create: `frontend/src/stores/copilot.js`
- Create: `frontend/src/components/AIPanel.svelte`
- Modify: `frontend/src/components/DatabasePanel.svelte`
- Modify: `frontend/src/components/DatabaseTablePanel.svelte`
- Modify: `frontend/src/components/TerminalPanel.svelte`
- Modify: `frontend/src/App.svelte`

**Interfaces:**
- Produces:

```javascript
export const COPILOT_APPLY_SQL = 'copilot:apply-sql';
export function applySqlEvent(sessionId, content) {
  return new CustomEvent(COPILOT_APPLY_SQL, { detail: { sessionId, content } });
}
export function shellExecutePayload(content) {
  const text = String(content || '').replace(/\n+$/, '');
  return `${text}\n`;
}
```

执行语义（第一版锁死）：

- SQL「填入」：对当前数据库 session 派发 `copilot:apply-sql`，`DatabasePanel` / `DatabaseTablePanel` 监听后设置 `query`。
- SQL「执行」：先 `CopilotClassify('sql', sql)`；危险则 `ConfirmDialog`；确认后用**当前编辑器** `query` 调 `ExecuteDatabaseQuery`。无打开的查询面板时，先填入再执行同一内容。
- Shell「填入」：`SendSSHData(sessionId, content)`，**不**加换行。
- Shell「执行」：分类确认后 `SendSSHData(sessionId, shellExecutePayload(content))`，输出留在终端。

`copilot.js` store：`{ open: false, width: 360 }`；`toggle()`。对话数组按 `sessionID` 存在 store 内存，关会话删掉。

`AIPanel`：无 session 显示「先连接主机或数据库」；缺 Key 显示去设置；输入框发送 `CopilotChat`。回复若有 `artifact` 则显示摘要、代码块、「填入」「执行」。生成中禁用发送。

`App.svelte`：顶栏设置按钮左侧增加 AI 按钮（`title="AI Copilot"`），打开可折叠列，放在主舞台与 SSH 工具坞之间（数据库模式也显示该列，不依赖 `showSessionDock`）。默认收起。

- [ ] **Step 1: 写失败测试**

`frontend/test/copilotApply.test.js`：

```javascript
import test from 'node:test';
import assert from 'node:assert/strict';
import { shellExecutePayload, applySqlEvent, COPILOT_APPLY_SQL } from '../src/lib/copilotApply.js';

test('shell execute appends a single newline', () => {
  assert.equal(shellExecutePayload('ls -la\n'), 'ls -la\n');
});

test('apply sql event uses session id', () => {
  const event = applySqlEvent('db-1', 'SELECT 1');
  assert.equal(event.type, COPILOT_APPLY_SQL);
  assert.equal(event.detail.sessionId, 'db-1');
});
```

- [ ] **Step 2: 跑测试确认失败**

Run: `cd frontend && node --test test/copilotApply.test.js`
Expected: FAIL，模块不存在。

- [ ] **Step 3: 实现辅助、store、面板与接线**

`DatabasePanel.svelte` / `DatabaseTablePanel.svelte` 在 `onMount` 监听 `COPILOT_APPLY_SQL`，`sessionId` 匹配才赋值 `query`。

`TerminalPanel.svelte` 增加 `export function insertCopilotText(sessionId, text)`：对该 session 调已有 `SendSSHData`。

`AIPanel` 通过 props 接收 `sessionId`、`mode`、`hasSession`、`onOpenSettings`。发送 `CopilotChat` 时带上该 `sessionId` 的历史消息（role/content），不含 API Key。关闭会话时从 store 删除该历史。

- [ ] **Step 4: 测试 + 构建**

Run:

```bash
cd frontend && node --test test/copilotApply.test.js
cd frontend && npm run build
go test ./internal/service/copilot ./internal/config -count=1
```

Expected: 全部 PASS / 构建成功。

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/copilotApply.js frontend/test/copilotApply.test.js frontend/src/stores/copilot.js frontend/src/components/AIPanel.svelte frontend/src/components/DatabasePanel.svelte frontend/src/components/DatabaseTablePanel.svelte frontend/src/components/TerminalPanel.svelte frontend/src/App.svelte
git commit -m "$(cat <<'EOF'
feat: add copilot side panel with apply and confirmed execute

EOF
)"
```

---

### Task 8: 实现说明与变更记录

**Files:**
- Create: `docs/development/2026-08-13-ai-copilot-sql-shell.md`
- Create: `docs/changes/features/2026-08-13-ai-copilot-sql-shell.md`
- Modify: `docs/superpowers/specs/2026-08-13-ai-copilot-sql-shell-design.md`（状态改为实现中/已落地）

**变更文档必须含：** 背景、范围、修改文件、验证、剩余风险。正文中文。

验证命令写入变更文档：

```bash
go test ./internal/service/copilot ./internal/config -count=1
cd frontend && node --test test/copilotApply.test.js
cd frontend && npm run build
```

手工回归（不在 CI）：规格第 9 节四条。若跳过 `wails dev` 手工验证，必须在剩余风险写明。

- [ ] **Step 1: 写文档**
- [ ] **Step 2: 对照规格第 2/7/9/10 节，确认计划内任务均已覆盖**
- [ ] **Step 3: Commit**

```bash
git add docs
git commit -m "$(cat <<'EOF'
docs: record AI copilot SQL/shell implementation

EOF
)"
```

---

## 手工回归清单

1. 设置里填写 DeepSeek/OpenAI 的 Base URL + 官方模型名 + Key，保存后 `~/.ahasshtools/config.json` 无 Key。
2. 数据库会话：说「查某表前 10 行」→ 填入编辑器 → 执行 → 结果在表面板。
3. SSH 会话：说「看磁盘」→ 填入终端无换行 → 执行后命令在 PTY 跑。
4. `DROP TABLE` / `rm -rf` 必须确认；取消后无执行。
5. 未配 Key、错误模型名、JDBC agent 不可用时提示可读。
