# 设计规格：会话日志记录与常用命令提示

日期：2026-09-04  
状态：待实现  
范围：会话日志（记录 / 搜索导出 / 保留策略 / 敏感过滤）、按连接常用命令提示、文件管理器目录跟踪默认开启、同步 `release.yml` 发布说明

## 1. 背景与目标

SSH 终端目前只有实时 PTY 输入输出（`SendSSHData` / `ssh:output:{sessionID}`），没有会话落盘、也没有按连接的命令统计与输入提示。产品路线图将「会话日志记录」列为核心增强；同时需要在键入时提示该服务器上的常用命令（例如频繁的 `cd` 路径），支持点击或 Tab 填入后再回车执行。

成功标准：

- 默认开启会话输出记录，可搜索、可导出，默认保留 30 天，写入前脱敏
- 按 connectionId 统计命令频率，输入时浮层提示；Tab/点击仅填入，回车执行
- 新连接的文件管理器「目录跟踪」默认开启（不覆盖用户已保存的关闭状态）
- `.github/workflows/release.yml` 发布正文包含本期更新要点

## 2. 已确认决策

| 决策点 | 选择 |
|--------|------|
| 交付范围 | 完整版：日志四项 + 命令提示 + release 同步 |
| 命令数据来源 | 按连接单独记命令史（回车时从前端当前行提取），不从日志反解析 |
| 日志默认策略 | 默认开启；保留 30 天；可关；默认开启脱敏 |
| 目录跟踪 | 仅将现有 SFTP `directory_tracking` 默认改为 `true` |
| 架构路线 | 后端落盘 + 前端补全层（方案 1） |
| Tab/点击行为 | 只填入命令，不自动执行；再按回车才发送 |

明确不在本期：端口转发、跳板机、标签页快捷键、从日志解析命令、命令浮层「智能常去目录」实体（与 FM 目录跟踪不同）。

## 3. 架构

```text
Terminal.svelte ──输入──► SendSSHData / 命令行缓冲
       │                      │
       │ 回车确认命令          │ 输出事件 ssh:output
       ▼                      ▼
 CommandHistoryService    SessionLogService
 (按 connectionId)         (按 connectionId + session)
       │                      │
       ▼                      ▼
 ~/.ahasshtools/commands/    ~/.ahasshtools/session-logs/
       │                      │
       └──── Settings ────────┘
```

| 单元 | 职责 | 非职责 |
|------|------|--------|
| `SessionLogService` | 追加写输出、脱敏、搜索、导出、保留清理 | 不解析 shell 命令 |
| `CommandHistoryService` | 按连接记录 / 排序 / 前缀查询 | 不写完整终端回放 |
| `Terminal.svelte` 补全层 | 行缓冲、回车上报、浮层、Tab/点击填入 | 不直接写磁盘 |
| Config / Settings UI | 开关、保留天数、目录跟踪默认 | 不处理 I/O |

关键约束：

- 命令史键为 **connectionId**（非临时 sessionId）
- 日志按连接分目录；启动及定时执行保留清理
- 脱敏在**写入前**执行
- SSH 连接为优先目标；本地终端若存在稳定键（如 `local`）且成本低则一并支持

## 4. 数据模型与存储

### 4.1 目录布局

```text
~/.ahasshtools/
  session-logs/{connectionId}/YYYY-MM-DDTHH-MM-SS_{sessionId}.log
  commands/{connectionId}.json
  config.json   # 扩展 AppSettings
```

### 4.2 AppSettings 新字段

| 字段 | 类型 | 默认 | 含义 |
|------|------|------|------|
| `session_log_enabled` | bool | `true` | 是否记录会话输出 |
| `session_log_retention_days` | int | `30` | 日志保留天数 |
| `session_log_redact_enabled` | bool | `true` | 写入前敏感过滤 |
| `command_suggest_enabled` | bool | `true` | 常用命令提示 |
| `command_suggest_limit` | int | `8` | 浮层最多条数 |

`DefaultFileManagerSettings.DirectoryTracking` 改为 `true`；前端 fallback 由 `?? false` 改为 `?? true`。已持久化为 `false` 的连接不强制覆盖。

### 4.3 命令史 JSON

```json
{
  "connection_id": "conn-xxx",
  "updated_at": "2026-09-04T10:55:00Z",
  "entries": [
    {
      "command": "cd /var/log/nginx",
      "count": 12,
      "last_used": "2026-09-04T10:50:00Z"
    }
  ]
}
```

排序：`count` 降序，同分再按 `last_used` 降序；查询时做前缀（及必要的子串）过滤。

### 4.4 导出 API（经 `app.go`）

会话日志：

- `ListSessionLogs(connectionId)`
- `SearchSessionLogs(connectionId, query, limit)`
- `ExportSessionLog(logId, destPath)`
- `DeleteSessionLog(logId)` / `PurgeExpiredSessionLogs()`
- 输出回调内：`SessionLogService.Append(connectionId, sessionId, data)`（已脱敏）

常用命令：

- `RecordCommand(connectionId, command)` — 忽略空行 / 纯空白
- `SuggestCommands(connectionId, prefix, limit)`

搜索结果只返回有限片段，不把整文件经事件推前端。

### 4.5 脱敏

写入前替换常见形态：`password=` / `passwd`、Bearer/API key、`BEGIN.*PRIVATE KEY`、长 hex/base64 token → `***`。可关。文档标明非安全审计级保证。

### 4.6 错误处理

- 日志目录不可写：跳过写入并警告一次，不阻断 SSH
- 命令史失败：降级为无建议，连接可用
- 清理失败：下次启动重试

## 5. UI / 交互

### 5.1 命令浮层

- 当前行非空且开关开启时，防抖 120–200ms 调用 `SuggestCommands`
- 浮层在光标附近，最多 `command_suggest_limit` 条；高亮前缀
- `↑`/`↓` 选择；`Tab`：**仅把选中项填入当前输入行，不发送**；`Esc` 关闭浮层，按键仍可交给终端
- 点击条目 = 与 Tab 相同，只填入不发送
- 有浮层时 `Enter` **不**自动采纳建议，只按终端原义提交当前缓冲（若已 Tab/点击填入，则本次 Enter 即执行该命令）
- 无匹配、IME 组字、多行粘贴选区时不弹
- 真正提交行（`\r`/`\n` 且缓冲有内容）后 `RecordCommand`；一期整行原样记录；多行粘贴按行拆分记录（空行跳过）

### 5.2 会话日志 UI

- 设置页「会话日志」：开关、保留天数、脱敏、命令提示
- 连接/会话入口「查看日志」：列表、搜索、导出、手动清理过期

### 5.3 发布

更新 `.github/workflows/release.yml` 的 `body`「更新内容」：会话日志能力、常用命令提示、目录跟踪默认开启。

## 6. 测试与风险

### 6.1 测试

- Go：追加/脱敏/保留删除/搜索；命令记录与排序建议；配置默认与缺字段兼容；`DirectoryTracking` 默认 true
- 前端：行缓冲与回车上报；有建议时 Tab 只填入；Esc；防抖
- 手工：日志落盘、提示排序与填入执行、设置开关、FM 默认跟踪、release 文案

### 6.2 风险

| 风险 | 缓解 |
|------|------|
| PTY 行缓冲不准（粘贴多行、Ctrl+C、退格） | 简单行缓冲；实现计划明确多行粘贴策略 |
| 脱敏漏报 | 默认开 + 文档声明局限 |
| 磁盘占用 | 30 天清理 + 可关记录 |
| 快捷键抢占 | 仅在有可见建议时拦截 Tab/方向键 |

## 7. 主要修改面（实现时）

- `internal/service/`：新增 session log / command history 服务与测试
- `internal/config/config.go`：settings 与 FM 默认
- `app.go`：导出方法；SSH 输出回调挂钩日志
- `frontend/src/components/Terminal.svelte`（及 Panel）：行缓冲、浮层、绑定 connectionId
- 设置 / 日志查看相关 UI 组件
- `frontend` stores 中 FM directoryTracking fallback
- `.github/workflows/release.yml`
- `docs/changes/features/` 变更记录；必要时 `docs/development/` 实现说明
