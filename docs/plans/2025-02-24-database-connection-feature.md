# 数据库连接功能开发计划

> **创建日期**: 2025-02-24
> **优先级**: 高
> **目标**: 实现 MySQL 和 PostgreSQL 数据库连接、查询、管理功能

---

## 📋 概述

在 SSHTools 应用中添加数据库连接功能，支持用户通过 GUI 界面连接、查询和管理 MySQL 和 PostgreSQL 数据库。

### 核心需求
- ✅ **数据库类型**: 支持 MySQL 和 PostgreSQL
- ✅ **查询界面**: GUI 风格（查询输入框 + 结果表格）
- ✅ **连接模式**: 一个连接 = 一个数据库（需要指定数据库名）
- ✅ **查询结果**: 表格格式显示（支持分页、排序）
- ✅ **密码加密**: 复用现有的 AES-GCM 加密存储

---

## 🏗️ 架构设计

### 后端架构

```
app.go (Wails 绑定层)
    ↓
service/database_service.go (业务逻辑层)
    ↓
config/database.go (数据模型)
    ↓
sql.DB (数据库连接)
```

### 前端架构

```
App.svelte (主应用)
    ↓
DatabasePanel.svelte (数据库查询面板)
    ↓
stores.js (状态管理)
    ↓
wailsjs/go/main/App.js (Wails 绑定)
```

---

## 📦 文件结构

```
sshTools/
├── internal/
│   ├── config/
│   │   ├── config.go          # 扩展 ConnectionConfig
│   │   └── database.go        # 新建：数据库配置
│   ├── service/
│   │   └── database_service.go # 新建：数据库服务
│   └── ssh/
│       └── manager.go          # 扩展：支持数据库会话管理
├── frontend/src/
│   ├── components/
│   │   ├── DatabasePanel.svelte # 新建：数据库查询面板
│   │   ├── AssetList.svelte     # 修改：onConnect 逻辑
│   │   └── App.svelte          # 修改：集成数据库面板
│   └── stores.js               # 修改：添加数据库会话状态
└── app.go                      # 修改：添加数据库 API
```

---

## 🎯 实施计划

### P0: MVP 阶段（核心功能）

#### 1. 数据模型扩展

**文件**: `internal/config/config.go`

**变更**:
```go
type ConnectionConfig struct {
    ID       string            `json:"id"`
    Name     string            `json:"name"`
    Host     string            `json:"host"`
    Port     int               `json:"port"`
    User     string            `json:"user"`
    AuthType string            `json:"auth_type"` // "password" or "key"
    KeyPath  string            `json:"key_path,omitempty"`
    Tags     []string          `json:"tags,omitempty"`
    Type     string            `json:"type"`         // 新增: "ssh", "database", "docker"
    Metadata map[string]string `json:"metadata,omitempty"` // 新增: {"database": "db_name", "db_type": "mysql"}
}
```

**验证**:
- [ ] 添加 `Type` 字段
- [ ] 添加 `Metadata` 字段
- [ ] 向后兼容（现有 SSH 连接不受影响）
- [ ] 配置导入/导出测试

---

#### 2. 数据库配置结构

**文件**: `internal/config/database.go`（新建）

**内容**:
```go
package config

import "time"

// DatabaseConfig 数据库连接配置
type DatabaseConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    DBType   string // "mysql" 或 "postgresql"
    Database string // 数据库名称（必填）
    Timeout  time.Duration
}

// TableSchema 表结构信息
type TableSchema struct {
    TableName string            `json:"table_name"`
    Columns   []ColumnSchema   `json:"columns"`
}

// ColumnSchema 列信息
type ColumnSchema struct {
    Name         string `json:"name"`
    Type         string `json:"type"`
    Nullable     bool   `json:"nullable"`
    IsPrimaryKey bool   `json:"is_primary_key"`
}
```

---

#### 3. 数据库驱动依赖

**文件**: `go.mod`

**添加依赖**:
```go
require (
    github.com/go-sql-driver/mysql v1.7.1
    github.com/lib/pq v1.10.9
)
```

**安装命令**:
```bash
go get github.com/go-sql-driver/mysql
go get github.com/lib/pq
```

---

#### 4. 数据库服务层

**文件**: `internal/service/database_service.go`（新建）

**核心方法**:

| 方法名 | 功能 | 参数 | 返回值 |
|--------|------|------|--------|
| `NewDatabaseService` | 初始化服务 | config, store | *DatabaseService |
| `ConnectDatabase` | 建立连接 | sessionID, host, port, user, password, dbType, database | error |
| `ExecuteQuery` | 执行 SQL 查询 | sessionID, query | *QueryResult, error |
| `ListTables` | 列出所有表 | sessionID | []string, error |
| `GetTableSchema` | 获取表结构 | sessionID, table | *TableSchema, error |
| `CloseDatabase` | 关闭连接 | sessionID | error |
| `TestConnection` | 测试连接 | host, port, user, password, dbType, database | error |

**数据结构**:
```go
type DatabaseSession struct {
    ID        string
    Config    config.DatabaseConfig
    DB        *sql.DB
    Connected bool
}

type QueryResult struct {
    Columns []string
    Rows    [][]interface{}
    Affected int
}
```

---

#### 5. 数据库会话管理器

**文件**: `internal/ssh/manager.go`（扩展现有文件）

**新增方法**:
```go
type SessionManager struct {
    // ... 现有字段
    databaseSessions map[string]*DatabaseSession
}

func (sm *SessionManager) AddDatabaseSession(id string, session *DatabaseSession)
func (sm *SessionManager) GetDatabaseSession(id string) (*DatabaseSession, error)
func (sm *SessionManager) RemoveDatabaseSession(id string)
func (sm *SessionManager) ListDatabaseSessions() []string
```

---

#### 6. 后端 API 集成

**文件**: `app.go`

**新增导出方法**:

```go
type App struct {
    // ... 现有服务
    databaseService *service.DatabaseService
}

func (a *App) startup(ctx context.Context) {
    // ... 现有初始化
    a.databaseService = service.NewDatabaseService(a.configManager, a.credentialStore)
}

// ========== 数据库连接 ==========

func (a *App) ConnectDatabase(sessionID, host, port, user, password, dbType, database string) error

func (a *App) ExecuteDatabaseQuery(sessionID, query string) (string, error)

func (a *App) ListDatabaseTables(sessionID string) ([]string, error)

func (a *App) GetTableColumns(sessionID, table string) ([]string, error)

func (a *App) CloseDatabase(sessionID string) error

// ========== 数据库测试 ==========

func (a *App) TestDatabaseConnection(host string, port int, user, password, dbType, database string) error
```

**事件发送**:
```go
runtime.EventsEmit(a.ctx, "db:output:"+sessionID, result)
runtime.EventsEmit(a.ctx, "db:tables:"+sessionID, tables)
runtime.EventsEmit(a.ctx, "db:columns:"+sessionID, columns)
```

---

### P1: 前端开发

#### 7. 数据库查询面板

**文件**: `frontend/src/components/DatabasePanel.svelte`（新建）

**UI 布局**:
```
┌─────────────────────────────────────────────┐
│ [Production DB - MySQL]            ▲ │
├────────────────────────────────────────────┤
│  Tables           │  Query Results    │
│  ├─ users      │  ┌─────────────┐  │
│  ├─ orders     │  │ SELECT * FROM│  │
│  ├─ products   │  │ users LIMIT 10│  │
│  └─ ...        │  └─────────────┘  │
│                 │         [执行] [清除]  │
│                 │                         │
│  Query History  │  ┌─────────────┐  │
│  1. SELECT...  │  │ id │ name    │  │
│  2. SELECT...  │  │ 1  │ John ... │  │
│  3. SELECT...  │  └─────────────┘  │
│  └─ ...        │                         │
├────────────────────────────────────────────┤
│  Export: [CSV]  显示 10 条          │
└─────────────────────────────────────────────┘
```

**核心功能**:
- SQL 查询输入框（多行文本）
- 执行按钮
- 清除按钮
- 结果表格（分页、排序、复制）
- 表列表侧边栏（双击表名填入查询）
- 查询历史记录（点击历史项填入查询）
- 导出按钮（CSV）

**事件监听**:
```javascript
onMount(() => {
  // 监听查询结果
  window.addEventListener('db:output:' + sessionId, handleQueryResult);

  // 监听表列表更新
  window.addEventListener('db:tables:' + sessionId, handleTablesUpdate);

  // 监听列信息更新
  window.addEventListener('db:columns:' + sessionId, handleColumnsUpdate);
});
```

---

#### 8. 状态管理

**文件**: `frontend/src/stores.js`

**新增**:
```javascript
// 数据库会话状态
export const databaseSessionsStore = writable(new Map());
export const activeDatabaseSessionIdStore = writable(null);

// 查询历史
export const queryHistoryStore = writable([]);

// 当前表的元数据
export const currentTablesStore = writable([]);
export const currentColumnsStore = writable([]);
```

---

#### 9. 资产列表集成

**文件**: `frontend/src/components/AssetList.svelte`

**修改**:
```javascript
async function onConnect(asset) {
  if (asset.type === 'database') {
    // 打开数据库面板
    const sessionId = 'db-' + Date.now();
    await connectToDatabase(sessionId, asset);
    openDatabasePanel(sessionId, asset);
  } else if (asset.type === 'ssh') {
    // SSH 连接（现有逻辑）
    openTerminal(asset);
  }
}

async function connectToDatabase(sessionId, asset) {
  if (!window.wailsBindings) return;

  const dbType = asset.metadata?.db_type || 'mysql';
  const database = asset.metadata?.database || '';

  try {
    await window.wailsBindings.ConnectDatabase(
      sessionId,
      asset.host,
      asset.port,
      asset.username,
      '', // 密码从加密存储加载
      dbType,
      database
    );
    activeDatabaseSessionIdStore.set(sessionId);
  } catch (error) {
    showError('连接失败', error.message);
  }
}
```

---

#### 10. 主应用集成

**文件**: `frontend/src/components/App.svelte`

**修改**:
```svelte
<script>
  // ... 现有导入
  import DatabasePanel from './DatabasePanel.svelte';

  let showDatabasePanel = false;
  let databaseSessionId = null;
  let databaseAsset = null;

  $: if ($activeDatabaseSessionIdStore !== null && $activeDatabaseSessionIdStore !== databaseSessionId) {
    databaseSessionId = $activeDatabaseSessionIdStore;
    showDatabasePanel = true;
    // 加载资产信息
    loadDatabaseAsset(databaseSessionId);
  }
</script>

{#if showDatabasePanel && databaseAsset}
  <div class="absolute inset-0 z-20 bg-black/50">
    <DatabasePanel
      sessionId={databaseSessionId}
      asset={databaseAsset}
      onClose={() => {
        showDatabasePanel = false;
        databaseSessionId = null;
        activeDatabaseSessionIdStore.set(null);
      }}
    />
  </div>
{/if}
```

---

## 🎨 UI 设计细节

### 数据库面板

**主面板**:
- 固定高度：100vh（全屏）
- 左右分栏：
  - 左：表列表（200px，可调整）
  - 右：查询结果（flex-1）

**查询区域**:
- 顶部：SQL 输入框（代码字体，等宽）
  - 语法高亮（简化版：关键字高亮）
  - 行号显示
- 工具栏：执行、清除、导出按钮

**结果表格**:
- 列排序功能
- 行高亮（悬停效果）
- 复制整行功能
- 导出为 CSV

**历史记录**:
- 下拉列表
- 最近 50 条
- 点击历史项填入查询框

### 颜色主题

**深色模式**:
- 背景：`#1e1e1e`
- 面板背景：`#2d2d2d`
- 表头：`#3b3b3b`
- 边框：`#404040`

**浅色模式**:
- 背景：`#ffffff`
- 面板背景：`#f9fafb`
- 表头：`#f3f4f6`
- 边框：`#e5e7eb`

---

## 🔐 安全考虑

### 密码加密存储

- 复用现有的 AES-256-GCM 加密
- 密钥派生基于机器特征
- 存储位置：`~/.sshtools/credentials.enc`

### SQL 注入防护

- 使用参数化查询（`Prepare` + `Exec`）
- 禁止直接拼接 SQL 字符串
- 查询执行超时（30 秒）

### 连接安全

- TLS/SSL 连接支持（MySQL: `?tls=true`，PostgreSQL: `sslmode=require`）
- 主机密钥验证（生产环境启用）
- 连接超时（15 秒）

---

## 🧪 测试计划

### 单元测试

**文件**: `internal/service/database_service_test.go`

**测试用例**:
```go
func TestConnectMySQL(t *testing.T)
func TestConnectPostgreSQL(t *testing.T)
func TestExecuteQuery(t *testing.T)
func TestListTables(t *testing.T)
func TestGetTableSchema(t *testing.T)
func TestCloseDatabase(t *testing.T)
func TestInvalidQuery(t *testing.T)
func TestConnectionTimeout(t *testing.T)
```

### 集成测试

- 测试完整的连接流程（前端 → 后端 → 数据库）
- 测试跨数据库类型切换（MySQL ↔ PostgreSQL）
- 测试并发会话管理
- 测试密码加密/解密流程

---

## 📊 数据库驱动对比

| 特性 | MySQL | PostgreSQL |
|------|--------|------------|
| 驱动包 | `github.com/go-sql-driver/mysql` | `github.com/lib/pq` |
| DSN 格式 | `user:password@tcp(host:port)/database` | `host=host port=port user=user password=password dbname=database` |
| TLS 支持 | `?tls=true` | `sslmode=require` |
| 批量插入 | ✅ | ✅ |
| 事务支持 | ✅ | ✅ |
| JSON 支持 | ✅ | JSONB |

---

## 🚀 部署清单

### 开发完成后

- [ ] 所有单元测试通过（`go test ./internal/service`）
- [ ] 构建成功（`wails build`）
- [ ] 前端开发模式正常运行（`wails dev`）
- [ ] 主题切换正常工作
- [ ] 密码加密/解密功能正常
- [ ] 跨平台测试（macOS, Windows, Linux）

### 发布前

- [ ] 更新 README.md（添加数据库功能说明）
- [ ] 更新 QUICK_START.md（数据库连接教程）
- [ ] 创建迁移指南（从现有工具迁移）
- [ ] 添加示例截图
- [ ] 性能基准测试

---

## 📝 API 文档

### Wails 绑定方法

| 方法名 | 功能 | 参数 | 返回值 |
|--------|------|------|--------|
| `ConnectDatabase` | 连接数据库 | sessionID, host, port, user, password, dbType, database | error |
| `ExecuteDatabaseQuery` | 执行 SQL 查询 | sessionID, query | string (JSON) |
| `ListDatabaseTables` | 列出表 | sessionID | []string |
| `GetTableColumns` | 获取列信息 | sessionID, table | []string |
| `CloseDatabase` | 关闭连接 | sessionID | error |
| `TestDatabaseConnection` | 测试连接 | host, port, user, password, dbType, database | error |

### 前端事件

| 事件名 | 数据 | 说明 |
|--------|------|------|
| `db:output:{sessionID}` | QueryResult | 查询结果 |
| `db:tables:{sessionID}` | []string | 表列表 |
| `db:columns:{sessionID}` | []ColumnInfo | 列信息 |

---

## 🔄 版本迭代计划

### v1.0.0（MVP）
- ✅ MySQL + PostgreSQL 连接
- ✅ SQL 查询执行
- ✅ 结果表格显示
- ✅ 表列表浏览
- ✅ 密码加密存储

### v1.1.0（增强）
- 📋 查询历史记录
- 📋 结果导出（CSV）
- 📋 表结构查看
- 📋 分页支持
- 📋 结果排序

### v1.2.0（完善）
- 📋 SQL 语法高亮
- 📋 列信息详细查看（类型、长度、是否主键）
- 📋 查询模板保存
- 📋 快捷键支持（Ctrl+Enter 执行）

### v2.0.0（未来）
- 📋 可视化查询构建器
- 📋 数据导入（CSV, SQL dump）
- 📋 查询性能分析
- 📋 MongoDB/Redis 支持
- 📋 迁移工具

---

## 📞 参考资料

### 官方文档

- [MySQL Go Driver](https://github.com/go-sql-driver/mysql)
- [PostgreSQL Go Driver](https://github.com/lib/pq)
- [database/sql 包](https://pkg.go.dev/database/sql)

### 类似项目

- [DBeaver](https://dbeaver.io/) - 数据库管理工具
- [TablePlus](https://tableplus.com/) - 数据库管理工具
- [HeidiSQL](https://www.heidisql.com/) - 数据库管理工具

---

## ❓ 待确认问题

1. **MySQL TLS 配置**：是否需要提供 TLS 选项（`?tls=true`）？
2. **PostgreSQL SSL 模式**：SSL 模式选项（`sslmode=disable/allow/prefer/require/verify-ca/verify-full`）？
3. **查询超时**：默认超时时间（建议 30 秒）？
4. **结果分页**：每页显示条数（建议 100 条）？
5. **历史记录**：查询历史保留条数（建议 50 条）？

---

## ✅ 完成标准

### 功能完整性

- [ ] 可以连接到 MySQL 数据库
- [ ] 可以连接到 PostgreSQL 数据库
- [ ] 可以执行 SELECT/INSERT/UPDATE/DELETE 查询
- [ ] 可以浏览表列表
- [ ] 可以查看表结构
- [ ] 可以导出查询结果
- [ ] 密码加密正常工作

### 性能要求

- [ ] 连接时间 < 3 秒
- [ ] 查询执行时间 < 5 秒（简单查询）
- [ ] 表列表加载时间 < 2 秒
- [ ] 表结构加载时间 < 2 秒

### 用户体验

- [ ] 错误提示友好、准确
- [ ] UI 响应流畅（无卡顿）
- [ ] 支持深色/浅色主题
- [ ] 支持键盘快捷键
- [ ] 支持窗口调整大小

---

**文档创建完成，开始实施 P0 阶段任务！** 🚀
