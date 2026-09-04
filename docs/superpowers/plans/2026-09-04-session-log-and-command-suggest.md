# 会话日志与常用命令提示 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 SSH 会话增加默认开启的会话日志（搜索/导出/保留/脱敏）、按连接的常用命令提示（Tab/点击填入），并将文件管理器目录跟踪默认改为开启，同步更新 release 说明。

**Architecture:** Go 侧新增 `SessionLogService`（`~/.ahasshtools/session-logs/`）与 `CommandHistoryService`（`~/.ahasshtools/commands/`）；`ConnectSSH` 输出回调挂钩日志；前端用可测的行缓冲/建议模块驱动 xterm 浮层。配置扩展 `AppSettings`；FM `DirectoryTracking` 默认 `true`。

**Tech Stack:** Go 1.24、Wails v2、Svelte 4、xterm.js、Node 内置 test runner、`go test`

## Global Constraints

- 文档正文使用中文（路径、标识符可英文）
- 变更需写 `docs/changes/features/`；设计已在 `docs/designs/` 与 `docs/superpowers/specs/`
- 脱敏非审计级保证；日志失败不得阻断 SSH
- Tab/点击只填入；Enter 提交当前缓冲（不自动采纳建议）
- 命令史按 `connectionId`；多行粘贴按行拆分记录

## 文件结构

| 文件 | 职责 |
|------|------|
| `internal/service/session_log_redact.go` | 会话日志脱敏（扩展 password/PEM/Bearer/长 token） |
| `internal/service/session_log_service.go` | 追加、列表、搜索、导出、保留清理 |
| `internal/service/command_history_service.go` | 按连接记录/建议 |
| `internal/config/config.go` | settings 字段 + FM 默认 |
| `app.go` | 服务装配、导出 API、输出挂钩、`session→connection` 映射 |
| `frontend/src/lib/commandLineBuffer.js` | 行缓冲与提交拆分（可单测） |
| `frontend/src/lib/commandSuggest.js` | 防抖查询与选中填入辅助 |
| `frontend/src/components/Terminal.svelte` | 浮层 UI 与按键 |
| `frontend/src/components/TerminalPanel.svelte` | 传入 connectionId、RecordCommand |
| `frontend/src/components/SessionLogPanel.svelte` | 日志列表/搜索/导出 |
| `frontend/src/components/GlobalSettingsDialog.svelte` | 日志与提示开关 |
| `frontend/src/stores.js` / `FileManager.svelte` | directoryTracking fallback |
| `.github/workflows/release.yml` | 发布说明 |
| `docs/changes/features/2026-09-04-session-log-and-command-suggest.md` | 变更记录 |

规格：`docs/superpowers/specs/2026-09-04-session-log-and-command-suggest-design.md`

---

### Task 1: 配置默认值与 settings 字段

**Files:**
- Modify: `internal/config/config.go`
- Modify: `internal/config/config_test.go`
- Test: `internal/config/config_test.go`

**Interfaces:**
- Produces: `AppSettings` 增加 `SessionLogEnabled bool`、`SessionLogRetentionDays int`、`SessionLogRedactEnabled bool`、`CommandSuggestEnabled bool`、`CommandSuggestLimit int`（json tag 与规格一致）；`DefaultFileManagerSettings().DirectoryTracking == true`；`UpdateSettings` 支持上述 key

- [ ] **Step 1: 写失败测试**

在 `internal/config/config_test.go` 追加：

```go
func TestDefaultSettingsIncludeSessionLogAndCommandSuggest(t *testing.T) {
	s := DefaultSettings()
	if !s.SessionLogEnabled {
		t.Fatal("session log should be enabled by default")
	}
	if s.SessionLogRetentionDays != 30 {
		t.Fatalf("retention days = %d, want 30", s.SessionLogRetentionDays)
	}
	if !s.SessionLogRedactEnabled {
		t.Fatal("redact should be enabled by default")
	}
	if !s.CommandSuggestEnabled {
		t.Fatal("command suggest should be enabled by default")
	}
	if s.CommandSuggestLimit != 8 {
		t.Fatalf("suggest limit = %d, want 8", s.CommandSuggestLimit)
	}
}

func TestDefaultFileManagerDirectoryTrackingEnabled(t *testing.T) {
	fm := DefaultFileManagerSettings()
	if !fm.DirectoryTracking {
		t.Fatal("directory tracking should default to true")
	}
}

func TestUpdateSettingsPersistsSessionLogFields(t *testing.T) {
	dir := t.TempDir()
	cm, err := NewConfigManager(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := cm.UpdateSettings(map[string]interface{}{
		"session_log_enabled":         false,
		"session_log_retention_days":  float64(7),
		"session_log_redact_enabled":  false,
		"command_suggest_enabled":     false,
		"command_suggest_limit":       float64(5),
	}); err != nil {
		t.Fatal(err)
	}
	s := cm.GetSettings()
	if s.SessionLogEnabled || s.SessionLogRedactEnabled || s.CommandSuggestEnabled {
		t.Fatalf("flags not updated: %+v", s)
	}
	if s.SessionLogRetentionDays != 7 || s.CommandSuggestLimit != 5 {
		t.Fatalf("numeric fields not updated: %+v", s)
	}
}
```

（若 `NewConfigManager` 签名不同，对照现有 `config_test.go` 调整构造方式。）

- [ ] **Step 2: 运行确认失败**

Run: `go test ./internal/config -run 'TestDefaultSettingsIncludeSessionLog|TestDefaultFileManagerDirectoryTracking|TestUpdateSettingsPersistsSessionLog' -v`

Expected: FAIL（缺字段或默认不符）

- [ ] **Step 3: 最小实现**

在 `AppSettings` 增加：

```go
SessionLogEnabled        bool `json:"session_log_enabled"`
SessionLogRetentionDays  int  `json:"session_log_retention_days"`
SessionLogRedactEnabled  bool `json:"session_log_redact_enabled"`
CommandSuggestEnabled    bool `json:"command_suggest_enabled"`
CommandSuggestLimit      int  `json:"command_suggest_limit"`
```

`DefaultSettings()` 赋值：`true, 30, true, true, 8`。

`DefaultFileManagerSettings()`：`DirectoryTracking: true`。

在 `UpdateSettings` 增加对应 `updates[...]` 分支（数字兼容 `float64`/`int`，与现有 pattern 一致）。

- [ ] **Step 4: 测试通过**

Run: 同 Step 2  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "$(cat <<'EOF'
feat(config): 会话日志与命令提示默认配置，目录跟踪默认开启

EOF
)"
```

---

### Task 2: 会话日志脱敏

**Files:**
- Create: `internal/service/session_log_redact.go`
- Create: `internal/service/session_log_redact_test.go`

**Interfaces:**
- Produces: `func RedactSessionLog(text string) string`
- 规则：password 赋值、PEM 私钥、Bearer token、长 hex（≥32）/长 base64（≥40）→ `***` 或等价占位；可参考 `internal/service/copilot/redact.go` 但放在 `service` 包供日志使用（勿循环依赖 copilot）

- [ ] **Step 1: 写失败测试**

```go
package service

import "testing"

func TestRedactSessionLogPassword(t *testing.T) {
	got := RedactSessionLog("export password=secret123")
	if got == "export password=secret123" || contains(got, "secret123") {
		t.Fatalf("password not redacted: %q", got)
	}
}

func TestRedactSessionLogBearerAndPEM(t *testing.T) {
	pem := "-----BEGIN RSA PRIVATE KEY-----\nABC\n-----END RSA PRIVATE KEY-----"
	got := RedactSessionLog("Authorization: Bearer abcdefghijklmnop " + pem)
	if contains(got, "abcdefghijklmnop") {
		t.Fatal("bearer not redacted")
	}
	if contains(got, "BEGIN RSA PRIVATE KEY") && contains(got, "ABC") {
		t.Fatal("pem body should be redacted")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || stringIndex(s, sub) >= 0)
}
func stringIndex(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
```

（可用 `strings.Contains` 简化。）

- [ ] **Step 2: 运行确认失败**

Run: `go test ./internal/service -run TestRedactSessionLog -v`  
Expected: FAIL

- [ ] **Step 3: 实现 `RedactSessionLog`**

```go
package service

import "regexp"

var (
	sessionLogPasswordRe = regexp.MustCompile(`(?i)(password|passwd)\s*=\s*\S+`)
	sessionLogBearerRe   = regexp.MustCompile(`(?i)(bearer\s+)[a-z0-9._\-]+`)
	sessionLogPEMRe      = regexp.MustCompile(`-----BEGIN [A-Z ]*PRIVATE KEY-----[\s\S]*?-----END [A-Z ]*PRIVATE KEY-----`)
	sessionLogLongHexRe  = regexp.MustCompile(`\b[a-fA-F0-9]{32,}\b`)
	sessionLogLongB64Re  = regexp.MustCompile(`\b[A-Za-z0-9+/]{40,}={0,2}\b`)
)

func RedactSessionLog(text string) string {
	text = sessionLogPasswordRe.ReplaceAllString(text, "${1}=***")
	text = sessionLogBearerRe.ReplaceAllString(text, "${1}***")
	text = sessionLogPEMRe.ReplaceAllString(text, "***")
	text = sessionLogLongHexRe.ReplaceAllString(text, "***")
	text = sessionLogLongB64Re.ReplaceAllString(text, "***")
	return text
}
```

按测试微调替换串。

- [ ] **Step 4: 测试通过**

Run: `go test ./internal/service -run TestRedactSessionLog -v`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/service/session_log_redact.go internal/service/session_log_redact_test.go
git commit -m "$(cat <<'EOF'
feat(service): 会话日志写入前脱敏

EOF
)"
```

---

### Task 3: SessionLogService

**Files:**
- Create: `internal/service/session_log_service.go`
- Create: `internal/service/session_log_service_test.go`

**Interfaces:**
- Produces:
  - `type SessionLogService struct`
  - `func NewSessionLogService(rootDir string) *SessionLogService`
  - `func (s *SessionLogService) Append(connectionID, sessionID string, data []byte, redact bool) error`
  - `func (s *SessionLogService) List(connectionID string) ([]SessionLogInfo, error)`
  - `func (s *SessionLogService) Search(connectionID, query string, limit int) ([]SessionLogHit, error)`
  - `func (s *SessionLogService) Export(logID, destPath string) error`
  - `func (s *SessionLogService) Delete(logID string) error`
  - `func (s *SessionLogService) PurgeExpired(retentionDays int) (int, error)`
  - `type SessionLogInfo struct { ID, ConnectionID, SessionID, Path string; Size int64; ModTime time.Time }`
  - `type SessionLogHit struct { LogID string; Line int; Text string }`
  - logID 约定：`{connectionID}/{filename}` 或绝对路径相对 root 的 rel path
  - 文件路径：`{root}/{connectionID}/{time}_{sessionID}.log`，time 用 `2006-01-02T15-04-05`
  - Append：同一 session 追加到同一文件（内存 map `sessionID→*os.File` 或按名 reopen）；目录不可写返回 error，调用方吞掉
  - Search：逐文件扫描，limit 条命中后停止

- [ ] **Step 1: 写失败测试（temp dir）**

```go
func TestSessionLogAppendListSearchPurge(t *testing.T) {
	root := t.TempDir()
	svc := NewSessionLogService(root)
	if err := svc.Append("c1", "s1", []byte("hello password=secret\n"), true); err != nil {
		t.Fatal(err)
	}
	list, err := svc.List("c1")
	if err != nil || len(list) != 1 {
		t.Fatalf("list=%v err=%v", list, err)
	}
	if strings.Contains(mustRead(t, list[0].Path), "secret") {
		t.Fatal("expected redaction on disk")
	}
	hits, err := svc.Search("c1", "hello", 10)
	if err != nil || len(hits) == 0 {
		t.Fatalf("hits=%v err=%v", hits, err)
	}
	old := filepath.Join(root, "c1", "2000-01-01T00-00-00_old.log")
	_ = os.MkdirAll(filepath.Dir(old), 0o755)
	_ = os.WriteFile(old, []byte("old\n"), 0o644)
	n, err := svc.PurgeExpired(30)
	if err != nil || n < 1 {
		t.Fatalf("purge n=%d err=%v", n, err)
	}
}
```

- [ ] **Step 2: 运行确认失败**

Run: `go test ./internal/service -run TestSessionLogAppendListSearchPurge -v`  
Expected: FAIL

- [ ] **Step 3: 实现服务**

实现上述接口；`Export` 用 `os.ReadFile` + `os.WriteFile`；`Delete` 删文件；打开文件失败打日志风格返回 error。注意并发：`sync.Mutex` 保护 open writers map。

- [ ] **Step 4: 测试通过**

Run: 同 Step 2 + `go test ./internal/service -run SessionLog -v`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/service/session_log_service.go internal/service/session_log_service_test.go
git commit -m "$(cat <<'EOF'
feat(service): 会话日志落盘、搜索与保留清理

EOF
)"
```

---

### Task 4: CommandHistoryService

**Files:**
- Create: `internal/service/command_history_service.go`
- Create: `internal/service/command_history_service_test.go`

**Interfaces:**
- Produces:
  - `func NewCommandHistoryService(rootDir string) *CommandHistoryService`
  - `func (s *CommandHistoryService) Record(connectionID, command string) error` — trim；空则 no-op
  - `func (s *CommandHistoryService) Suggest(connectionID, prefix string, limit int) ([]CommandHistoryEntry, error)`
  - `type CommandHistoryEntry struct { Command string; Count int; LastUsed time.Time }`
  - 文件：`{root}/{connectionID}.json`；排序 count desc, last_used desc；前缀匹配（`strings.HasPrefix`），不足再用 `Contains`
  - 单连接最多保留 500 条（超出按 last_used 淘汰）

- [ ] **Step 1: 写失败测试**

```go
func TestCommandHistoryRecordAndSuggest(t *testing.T) {
	svc := NewCommandHistoryService(t.TempDir())
	_ = svc.Record("c1", "cd /var/log")
	_ = svc.Record("c1", "cd /var/log")
	_ = svc.Record("c1", "cd /tmp")
	got, err := svc.Suggest("c1", "cd", 8)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) < 2 || got[0].Command != "cd /var/log" || got[0].Count != 2 {
		t.Fatalf("unexpected suggest: %+v", got)
	}
	if err := svc.Record("c1", "   "); err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 2: 运行确认失败**

Run: `go test ./internal/service -run TestCommandHistoryRecordAndSuggest -v`  
Expected: FAIL

- [ ] **Step 3: 实现**

JSON 读写加 `sync.Mutex`；原子写可用 temp + rename。

- [ ] **Step 4: 测试通过**

Run: 同 Step 2  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/service/command_history_service.go internal/service/command_history_service_test.go
git commit -m "$(cat <<'EOF'
feat(service): 按连接记录与建议常用命令

EOF
)"
```

---

### Task 5: App 装配与 Wails API

**Files:**
- Modify: `app.go`
- Modify: `main.go`（若需 Bind 列表；Wails 通常自动导出 App 方法）

**Interfaces:**
- Consumes: Task 1–4 服务
- Produces 导出方法：
  - `BindSessionConnection(sessionID, connectionID string)`
  - `ListSessionLogs(connectionID string) ([]service.SessionLogInfo, error)`
  - `SearchSessionLogs(connectionID, query string, limit int) ([]service.SessionLogHit, error)`
  - `ExportSessionLog(logID string) (string, error)` — 内部 `SaveFileDialog` 后 Export，返回路径
  - `DeleteSessionLog(logID string) error`
  - `PurgeExpiredSessionLogs() (int, error)` — 读 settings 保留天数
  - `RecordCommand(connectionID, command string) error`
  - `SuggestCommands(connectionID, prefix string, limit int) ([]service.CommandHistoryEntry, error)`
- `ConnectSSH` 回调中：若 enabled，用 `sessionConnection[sessionID]` 调 `Append`；无 mapping 则跳过
- `startup`：`PurgeExpiredSessionLogs` 一次；root = `filepath.Join(home, ".ahasshtools", "session-logs")` 与 `commands`

查找现有 config 根目录获取方式（credential/config 初始化），复用同一 `~/.ahasshtools`。

- [ ] **Step 1: 在 App 结构体增加字段并在 `NewApp`/startup 初始化服务**

```go
sessionLogService     *service.SessionLogService
commandHistoryService *service.CommandHistoryService
sessionConnectionMu   sync.Mutex
sessionConnection     map[string]string
```

- [ ] **Step 2: 实现 Bind + 日志/命令 API**（完整方法体写入 app.go，错误吞掉策略与规格一致）

- [ ] **Step 3: 修改 ConnectSSH 输出回调**

```go
err := a.sessionService.ConnectSSH(..., func(data []byte) {
    // existing cwd + emit
    a.appendSessionLog(sessionID, data)
}, ...)
```

```go
func (a *App) appendSessionLog(sessionID string, data []byte) {
    if a.sessionLogService == nil || a.settingsService == nil {
        return
    }
    settings := a.settingsService.GetSettings()
    if !settings.SessionLogEnabled {
        return
    }
    a.sessionConnectionMu.Lock()
    connID := a.sessionConnection[sessionID]
    a.sessionConnectionMu.Unlock()
    if connID == "" {
        return
    }
    _ = a.sessionLogService.Append(connID, sessionID, data, settings.SessionLogRedactEnabled)
}
```

`CloseSSH` 时删除 mapping；可选关闭该 session 的 log file handle（若服务提供 `CloseSession(sessionID)`）。

- [ ] **Step 4: 编译**

Run: `go build -o /dev/null .`  
Expected: 成功

- [ ] **Step 5: Commit**

```bash
git add app.go
git commit -m "$(cat <<'EOF'
feat(app): 暴露会话日志与命令建议 API 并挂钩 SSH 输出

EOF
)"
```

---

### Task 6: 前端行缓冲与建议辅助（可单测）

**Files:**
- Create: `frontend/src/lib/commandLineBuffer.js`
- Create: `frontend/src/lib/commandSuggest.js`
- Create: `frontend/test/commandLineBuffer.test.js`
- Create: `frontend/test/commandSuggest.test.js`

**Interfaces:**
- Produces:
  - `createCommandLineBuffer()` → `{ push(data), flushPending(), getLine() }`
    - `push`：处理可打印字符、`\x7f`/`\b` 退格、`\r`/`\n` 返回 `{ submitted: string[] }`（多行粘贴拆分）；Ctrl+C (`\x03`) 清空
  - `pickSuggestFill(currentLine, suggestion)` → 返回应用建议后的完整行（一期：直接用 suggestion 替换整行）
  - `shouldOfferSuggest(line, enabled)` → trim 后非空且 enabled

- [ ] **Step 1: 写失败测试**

```js
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { createCommandLineBuffer } from '../src/lib/commandLineBuffer.js';

test('enter submits line and clears', () => {
  const buf = createCommandLineBuffer();
  buf.push('cd /tmp');
  const { submitted } = buf.push('\r');
  assert.deepEqual(submitted, ['cd /tmp']);
  assert.equal(buf.getLine(), '');
});

test('backspace and multiline paste', () => {
  const buf = createCommandLineBuffer();
  buf.push('ab');
  buf.push('\x7f');
  assert.equal(buf.getLine(), 'a');
  const { submitted } = buf.push('one\ntwo\r');
  assert.deepEqual(submitted, ['aone', 'two']);
});
```

- [ ] **Step 2: 运行确认失败**

Run: `cd frontend && node --test test/commandLineBuffer.test.js`  
Expected: FAIL

- [ ] **Step 3: 实现 buffer + suggest helpers**

- [ ] **Step 4: 测试通过**

Run: `cd frontend && node --test test/commandLineBuffer.test.js test/commandSuggest.test.js`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/commandLineBuffer.js frontend/src/lib/commandSuggest.js frontend/test/commandLineBuffer.test.js frontend/test/commandSuggest.test.js
git commit -m "$(cat <<'EOF'
feat(frontend): 终端命令行缓冲与建议辅助函数

EOF
)"
```

---

### Task 7: Terminal 浮层与 TerminalPanel 接线

**Files:**
- Modify: `frontend/src/components/Terminal.svelte`
- Modify: `frontend/src/components/TerminalPanel.svelte`

**Interfaces:**
- Consumes: `RecordCommand`、`SuggestCommands`、`BindSessionConnection`、Task 6
- `Terminal` 新增 props：`connectionId`、`commandSuggestEnabled`、`commandSuggestLimit`
- 在 `terminal.onData` 前/中：先 `buffer.push`；若有 `submitted`，对每条 `RecordCommand(connectionId, cmd)`（fire-and-forget）
- 输入变化防抖调用 `SuggestCommands`；渲染绝对定位列表于终端容器内
- 有可见建议时：拦截 `Tab` 填入（通过 `terminal.onKey` 或 addon）：计算需发送的退格 + 新文本差量，经 `onData` 发给远端以保持 PTY 同步；**不要**自动发 `\r`
- `ArrowUp/Down` 仅在浮层打开时改选中索引并 `preventDefault`
- `Escape` 关闭浮层
- `Enter` 不采纳建议
- `TerminalPanel`：`ConnectSSH` 前 `BindSessionConnection(sessionId, asset.id)`；把 `connection?.id` 与 settings 传入 `Terminal`

填入 PTY 同步策略（必须遵守）：

1. 设 `n = currentLine.length`
2. 发送 `n` 次 `\x7f`（或 `\b`）清除远端回显行（简单策略；若 prompt 复杂可能偏差，记入风险）
3. 发送 suggestion 字符串
4. 更新本地 buffer 为 suggestion

- [ ] **Step 1: 接线 BindSessionConnection + props**

- [ ] **Step 2: 实现浮层 DOM/CSS（贴合现有终端主题变量，避免新卡片风格堆砌）**

- [ ] **Step 3: 手工或最小前端测试验证 Tab 不发送 CR**

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/Terminal.svelte frontend/src/components/TerminalPanel.svelte
git commit -m "$(cat <<'EOF'
feat(terminal): 常用命令浮层与按连接记录

EOF
)"
```

---

### Task 8: 设置页与会话日志面板

**Files:**
- Modify: `frontend/src/components/GlobalSettingsDialog.svelte`
- Modify: `frontend/src/settings/appearance.js`（`getDefaultAppSettings` 补默认字段）
- Create: `frontend/src/components/SessionLogPanel.svelte`
- Modify: `frontend/src/components/SessionToolDock.svelte` 或 `TerminalPanel.svelte`（增加「会话日志」入口，传入当前 `connectionId`）

**Interfaces:**
- 设置分区 `session-log`：四个开关/数字（enabled、retention、redact、suggest、limit）
- `SessionLogPanel`：`ListSessionLogs`、搜索框 → `SearchSessionLogs`、导出 → `ExportSessionLog`、清理 → `PurgeExpiredSessionLogs`

- [ ] **Step 1: 扩展 `getDefaultAppSettings` 与 dialog normalizeDraft**

- [ ] **Step 2: 实现 SessionLogPanel UI**

- [ ] **Step 3: 挂到会话工具区；确认 settings 保存走现有 `UpdateSettings`**

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/GlobalSettingsDialog.svelte frontend/src/settings/appearance.js frontend/src/components/SessionLogPanel.svelte frontend/src/components/SessionToolDock.svelte frontend/src/components/TerminalPanel.svelte
git commit -m "$(cat <<'EOF'
feat(ui): 会话日志设置与查看面板

EOF
)"
```

---

### Task 9: 目录跟踪默认 fallback + release + 变更文档

**Files:**
- Modify: `frontend/src/components/FileManager.svelte`（`?? false` → `?? true`）
- Modify: `frontend/src/stores.js`（同上）
- Modify: `.github/workflows/release.yml`
- Create: `docs/changes/features/2026-09-04-session-log-and-command-suggest.md`
- Create: `docs/development/2026-09-04-session-log-and-command-suggest.md`（实现说明）

- [ ] **Step 1: 改 FM fallback**

```js
directoryTracking: config?.directory_tracking ?? true,
```

两处保持一致。

- [ ] **Step 2: 更新 release.yml body**

在「新增功能」下列出：

```text
- 会话日志：自动记录终端输出，支持搜索/导出、可配置保留天数与敏感信息过滤（默认开启，保留 30 天）
- 常用命令提示：按连接统计频率，输入时 Tab/点击填入后再回车执行
- 文件管理器目录跟踪默认开启（已关闭的连接保持用户选择）
```

- [ ] **Step 3: 写中文变更文档（背景、范围、修改文件、验证、剩余风险）**

- [ ] **Step 4: 回归命令**

```bash
go test ./internal/config ./internal/service -count=1
cd frontend && node --test test/commandLineBuffer.test.js test/commandSuggest.test.js
```

Expected: 全部 PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/FileManager.svelte frontend/src/stores.js .github/workflows/release.yml docs/changes/features/2026-09-04-session-log-and-command-suggest.md docs/development/2026-09-04-session-log-and-command-suggest.md
git commit -m "$(cat <<'EOF'
docs: 会话日志与命令提示发布说明与变更记录

EOF
)"
```

---

## Spec 覆盖自检

| 规格项 | 任务 |
|--------|------|
| 自动记录终端输出 | T3, T5 |
| 搜索和导出 | T3, T5, T8 |
| 保留策略 | T1, T3, T5, T8 |
| 敏感过滤 | T2, T3 |
| 常用命令排序/提示 | T4, T6, T7 |
| Tab/点击填入、Enter 执行 | T7 |
| 目录跟踪默认开 | T1, T9 |
| release.yml | T9 |
| 日志失败不阻断 SSH | T5 |
| 多行粘贴按行记录 | T6 |

无 TBD/占位步骤；类型名在任务间一致（`SessionLogInfo`、`CommandHistoryEntry`、`RedactSessionLog`）。
