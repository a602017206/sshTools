# 运维工具多域导航、品牌图标与工作区拆分设计

## 背景

当前资产统一按 `type: ssh | database` 管理，Redis / Elasticsearch / Kafka 等被归入「数据库」并共用 `NativeDatabasePanel`。结果是：

- 图标同构换色，难以区分 MySQL、Redis、ES、Kafka
- 能力开关（`canQuery` / `canWrite`）无法表达各域真实交互，出现功能错位
- 后期 Docker 等模块若继续塞进同一抽象，包体积与内存压力会持续上升

本设计将产品明确为多域运维客户端，并给出左侧域切换轨、官方风格图标、工作区插件化与体积控制策略。

## 目标

1. 窗口最左侧增加窄轨域切换：全部 / SSH / 数据库 / 缓存 / 搜索 / 消息队列 / Docker
2. 资产树按当前域过滤；图标采用官方识别色与简化官方图形（含 NoSQL）
3. 各域使用独立工作区契约，不再把 Redis/ES/Kafka 假装成关系型库
4. 为 Docker 预留域位，但默认不内嵌重依赖，控制二进制与运行时开销

## 非目标（本阶段不做）

- 不实现完整 Docker 引擎或内嵌 Docker Desktop
- 不一次性重写 JDBC 关系型工作区
- 不拆成多个独立安装包（可列为远期选项）

## 域模型

| 域 ID | 显示名 | 资产来源 | 工作区 |
|-------|--------|----------|--------|
| `all` | 全部 | 全部资产 | 不强制切换主工作区 |
| `ssh` | SSH | `type=ssh`（及非 data 类主机） | Terminal / SFTP / Monitor |
| `database` | 数据库 | JDBC / SQLite / 关系型 + Mongo/Cassandra/Couchbase/Influx/Neo4j（文档/宽表类可暂挂此或后续再拆） | JDBC 工作区；文档库后续专属页 |
| `cache` | 缓存 | Redis、Memcached、KeyDB | RedisWorkspace 等 |
| `search` | 搜索 | Elasticsearch、OpenSearch | ElasticsearchWorkspace |
| `mq` | 消息队列 | Kafka | KafkaWorkspace |
| `docker` | Docker | 预留：本地 Docker 上下文 / 远程 Docker Host | DockerWorkspace（后期） |

**首期映射（明确）：**

- Redis / Memcached → `cache`
- Elasticsearch → `search`
- Kafka → `mq`
- MySQL / PostgreSQL / Oracle / SQL Server / 达梦 / Kingbase / openGauss / SQLite → `database`
- MongoDB / Cassandra / Couchbase / InfluxDB / Neo4j → 首期仍显示在「数据库」域，但工作区按原生类型路由；后续可再拆「文档/图」域

资产持久化增加：

```text
metadata.domain = cache | search | mq | database | ssh | docker
```

若缺失，则由 `type` + `db_type` 推导（兼容旧配置）。

## 布局：左侧域切换轨

```text
┌────┬──────────────┬─────────────────────────────┐
│轨  │ 资产树        │ 主工作区                      │
│标  │ AssetList     │ Terminal / Redis / ES / ...  │
│签  │               │                              │
└────┴──────────────┴─────────────────────────────┘
```

### 行为

- 轨宽固定约 **48–56px**，仅图标 + tooltip（或极短字）
- 选中态高亮；「全部」为默认
- 切换域时：
  - 资产树只显示该域资产（分组仍保留）
  - **不自动断开**已有会话；仅过滤列表
  - 顶部 `WorkspaceNavigation`（SSH / 数据库）逐步演进为与域一致，或由域轨取代重复入口（实现时二选一，推荐：**域轨负责资产过滤，工作区导航负责主舞台模式**，避免双轨打架）
- 折叠侧栏时：域轨可保留为最左一条「迷你轨」，或与资产树一并折叠（推荐：**一并折叠**，展开时恢复上次域）

### 新建连接

- 点「添加」时，默认预选当前域对应的资产类型（在缓存域则默认 Redis）
- 对话框内类型列表按域分组展示，并使用品牌图标

## 品牌图标

### 策略

- 侧栏 16–18px：官方识别色 + **简化图形**（非完整营销 Logo 拼贴）
- 新建对话框 24–28px：同套图标 + 类型名
- 未知类型：通用几何图标（不再用统一蓝圆柱）

### 覆盖类型

SSH、MySQL、PostgreSQL、SQLite、Oracle、SQL Server、达梦、Kingbase、openGauss、Redis、MongoDB、Elasticsearch、Memcached、Cassandra、Couchbase、InfluxDB、Neo4j、Kafka；Docker 预留。

### 合规说明

使用简化品牌识别图用于工具内连接类型标识；UI 不冒充官方产品。实现时优先 SVG 组件（类似现有 `DatabaseTypeIcon`），按类型拆分路径，避免「同构图换色」。

## 工作区拆分

### 公共层（保留）

- 连接 CRUD、凭据、测试连通、会话 ID、关闭会话
- `NativeDatabaseService` 可保留为「原生协议会话总线」，但 UI 不再共用一个胖 Panel

### 私有工作区（插件）

| 工作区 | 入口条件 | 专属能力 |
|--------|----------|----------|
| `JdbcWorkspace` | JDBC 关系型 | SQL、表结构、对象树 |
| `RedisWorkspace` | Redis | DB 下拉、多类型键编辑 |
| `ElasticsearchWorkspace` | ES | 集群概览、索引搜索、Mapping、DSL |
| `KafkaWorkspace` | Kafka | Topic / 分区元数据（后续消费组） |
| `SshWorkspace` | SSH | 现有终端栈 |
| `DockerWorkspace` | 后期 | 容器/镜像列表（CLI 或 API） |

`TerminalPanel` 按 `workspaceKind` 挂载对应组件；能力矩阵写在各 workspace 配置中，**禁止**跨域默认开启无关按钮。

## 包体积与性能

### 原则

1. **壳薄、域隔离、按需加载**
2. 重依赖优先外置进程或系统 CLI（JDBC Agent 模式已验证）
3. Docker **不内嵌引擎**；调用本机 `docker` / remote API

### 具体措施

| 手段 | 说明 |
|------|------|
| 前端按域懒加载工作区组件 | 减小首屏 JS |
| Go 可选 build tag | 如 `docker`、`kafka` 实验特性 |
| 会话销毁即关客户端 | 避免 ES/Redis 连接常驻 |
| 列表虚拟滚动 | 大索引/大键列表 |
| 结果上限 | 延续现有 preview / size clamp |

### 预期

- 图标与域轨：体积影响可忽略
- 工作区拆分：二进制几乎不变，**降低错误功能与无效 UI 状态**
- Docker 若走 CLI：增加有限前端 + 薄 API 封装；若内嵌 SDK 再评估

## 实现分期

### Phase 1（本设计首批实现）

1. 域推导函数 + 资产过滤
2. 左侧域切换轨 UI
3. 官方风格图标替换（含 NoSQL）
4. 文档：变更记录

### Phase 2

1. 拆出 `RedisWorkspace` / `ElasticsearchWorkspace`（从 `NativeDatabasePanel` 迁出）
2. Kafka 独立工作区骨架
3. 新建连接对话框按域分组

### Phase 3

1. Docker 域入口（检测本机 docker）
2. build tag / 懒加载策略落地
3. 评估 Mongo 等是否从「数据库」再拆域

## 风险

- 旧连接无 `metadata.domain`：必须稳定推导，避免资产「消失」
- 域轨与现有 SSH/数据库顶栏导航可能重复：实现 Phase 1 时明确交互优先级
- 官方图标商标观感：使用简化识别图，避免完整 Logo 墙

## 验收标准

- 切换「缓存」仅见 Redis/Memcached；「搜索」仅见 ES；「消息队列」仅见 Kafka
- Redis / ES / MySQL 图标在侧栏可一眼区分
- 在缓存域打开 Redis 不再出现「像关系库一样」的无关空态文案
- 侧栏折叠/展开后域选择保持
