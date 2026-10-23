# Docker 容器连接设计

## 背景

多域导航已预留 Docker 域与「添加连接」中的 Docker 页签，但尚未实现真实连通。当前表单沿用 SSH 式「主机 / 端口 / 用户名 / 密码」（默认 2375），无法覆盖本机 Docker（Unix socket / named pipe），也容易把 Docker API 误当成需要 root 密码的服务。

产品对标 Hexhub：可从已有 SSH 关联添加 Docker，也可手工配置；同时必须能连本机 Docker Desktop / 本机 daemon。

## 已定口径（用户拍板）

1. **第一期范围**：三种接入入口都具备；能力只做到「测连 / 保存 / 打开工作区 / 列容器」。
2. **SSH 路径**：优先 SSH 隧道到远端 `docker.sock` + Docker Engine API；失败且允许时回退远端 `docker` CLI。
3. **资产形态**：独立 `type=docker` 资产；侧栏可新建，SSH 右键也可「添加 Docker」；底层同一模型。
4. **本机**：检测到可用后提示「添加本机 Docker」；用户确认后再写入资产列表。
5. **实现路线**：统一 Docker API 客户端抽象；CLI 仅作 SSH 回退，不作为全局默认。

## 目标

- 本机 / 关联 SSH / 直连 TCP 三种方式均可创建并测连。
- 打开 `DockerWorkspace` 后能列出容器（含停止中的，可筛「仅运行中」）。
- 与现有 `connections[]` + `metadata` + 凭据库模式对齐，不另起一套配置文件。
- 不内嵌 Docker 引擎；不把 Docker 工作区挂到 SSH 终端栈。

## 非目标（第一期不做）

- 容器日志、exec / attach、启停删、镜像 / 网络 / 卷、Compose、K8s
- 用「用户名 + 密码」作为 Docker Engine API 认证
- TCP TLS 客户端证书完整配置（可预留字段；第一期以明文 2375 为主）
- 内嵌引擎或强制依赖 Docker Desktop UI
- Flutter / 移动端 Docker

## 资产模型

Docker 存于现有 `AppConfig.Connections`，`type: "docker"`，类型字段进 `metadata`。

| 字段 | 说明 |
|------|------|
| `id` / `name` / `tags` | 与 SSH、数据库一致；分组仍用 tags / 既有分组约定 |
| `type` | 固定 `"docker"` |
| `metadata.domain` | `"docker"` |
| `metadata.transport` | `local` \| `ssh` \| `tcp` |
| `metadata.ssh_connection_id` | 仅 `ssh`：关联的 SSH 资产 ID |
| `metadata.docker_host` | 可选 Host URL 覆盖 |
| `metadata.socket_path` | 本机非默认 socket 时使用 |
| `metadata.remote_socket` | SSH 隧道远端 socket，默认 `/var/run/docker.sock` |
| `metadata.cli_fallback` | 默认 `true`；`false` 时禁止 CLI 回退 |
| `metadata.tls` | 仅 `tcp`；第一期可存布尔，证书路径后期 |
| `host` / `port` | 仅 `tcp`：API 地址（默认 2375 / 2376） |

### 表单行为

1. 先选接入方式：本机 / 关联 SSH / 直连 API。
2. **本机**：几乎只填名称与分组；不出现主机、端口、用户、密码。
3. **关联 SSH**：下拉已有 SSH；名称默认「`{SSH名}-Docker`」。
4. **直连 API**：主机、端口；第一期去掉易误导的用户名/密码字段。

### 侧栏与入口

- Docker 域只列出 `type === 'docker'`。
- SSH 资产右键「添加 Docker」→ 预填 `transport=ssh` + `ssh_connection_id`。
- 本机探测成功 → 提示「添加本机 Docker」→ 预填 `transport=local`；确认后落盘。

## 连通设计

统一接口（第一期）：

- `Ping`
- `ListContainers(all bool) []ContainerSummary`

`ContainerSummary` 至少包含：短 ID、名称、镜像、状态、端口映射、创建时间。API 与 CLI 回退都映射到同一结构。

### 本机（`local`）

- 默认平台 socket（Linux/macOS Unix socket；Windows named pipe）；尊重 `socket_path` / `DOCKER_HOST`。
- 探测：打开 Docker 域或应用适当时机 `Ping`；成功则提示添加。
- 不经 SSH，不填 IP/密码。

### 关联 SSH（`ssh`）

1. 解析 `ssh_connection_id`，复用现有 SSH 凭据与连接能力。
2. **主路径**：本地随机端口经 SSH 转发到远端 `remote_socket` → 本机 Docker API 访问 `localhost:转发端口`。
3. **回退**：隧道失败且 `cli_fallback` 为真 → SSH 上执行 `docker ps`（JSON 格式）并解析。
4. `cli_fallback=false` 时测连直接失败并给出原因。
5. 关闭 Docker 会话时拆除对应隧道，避免端口泄漏。

### 直连 API（`tcp`）

- `tcp://host:port`；第一期支持无 TLS 的 2375。
- 不做用户名密码认证；TLS 证书配置预留到后续。
- 对明文 2375 测连成功时可弱提示安全风险，不阻断。

### 测连

- 成功：`Ping` OK，或 CLI 回退下 `docker version` / `docker ps` 可执行。
- 失败：可读错误（socket 不存在、SSH 失败、隧道被拒、远端无 docker、TCP 拒绝等）。

## 第一期工作区

- 双击 Docker 资产 → `DockerWorkspace`（`workspaceTabs` 中 docker 不再映射到 `ssh`）。
- 顶栏：连接名、接入方式徽章（本机 / SSH / TCP）、刷新、断开。
- 主区：容器表（名称、短 ID、镜像、状态、端口、创建时间）；默认可含已停止；可筛「仅运行中」。
- 空态区分：无容器 / 未连接 / 当前走 CLI 回退。
- 不做详情、日志、exec、生命周期操作。

## 后端 API（示意）

| 方法 | 职责 |
|------|------|
| `DetectLocalDocker()` | 本机是否可添加 |
| `TestDockerConnection(cfg)` | 测连 |
| `ConnectDocker(sessionID, cfg)` | 建立会话（含隧道如需） |
| `CloseDocker(sessionID)` | 关闭会话并清理隧道 |
| `ListDockerContainers(sessionID, opts)` | 列容器 |

实现位置建议：`internal/docker/`（客户端与隧道）+ `internal/service/docker_service.go`（会话管理）+ `app.go` 导出；不内嵌引擎。

## 与既有设计关系

- 承接 [多域导航设计](../../designs/2026-08-28-multi-domain-nav-and-brand-icons.md) 中 Docker 域预留与「不内嵌引擎」原则。
- 修正该文档 Phase 3「仅域入口」为：本规格定义的完整第一期连接与列表能力。

## 风险与接受

| 风险 | 处理 |
|------|------|
| 远端无 sock 权限 | CLI 回退；列表字段可能略少 |
| 明文 2375 不安全 | 弱提示；后续补 TLS |
| CLI 输出因版本差异解析失败 | 固定 `--format`；失败时明确报错 |
| SSH 隧道端口占用 | 随机端口 + 会话结束必拆 |

## 验证（实现阶段）

- Go：本机 Ping（有/无 daemon）、隧道路径单测（可 mock）、CLI 解析单测、`cli_fallback` 开关。
- 前端：表单按 `transport` 显隐字段；工作区列表渲染与筛选；Docker 不再走 SSH workspace。
- 手工：本机添加提示、SSH 关联列容器、TCP 2375（若有环境）、隧道失败回退 CLI。

## 后续（非本规格范围）

- 日志 / exec / 启停、镜像管理、TLS 证书 UI、compose、移动端。
