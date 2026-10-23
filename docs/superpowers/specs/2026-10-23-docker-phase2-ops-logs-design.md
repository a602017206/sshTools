# Docker 二期：容器操作、详情与日志设计

## 背景

第一期已交付 Docker 三种接入与容器列表（见 `2026-10-23-docker-connection-design.md`）。日常运维在「看见列表」之后还需要启停删、查看详情与日志；exec 成本更高，本期限不做。

## 已定口径（用户拍板）

1. **范围（方案 B）**：启停 / 重启 / 删除 + 行内操作 + 右侧详情侧栏（概览 + 日志）；**不做 exec/attach**。
2. **CLI 回退**：与 API 会话同等支持写操作与日志（远端 `docker start/stop/restart/rm/inspect/logs`）。
3. **日志呈现**：详情侧栏内嵌；日志为侧栏 Tab（默认尾部 N 行，可跟流）。
4. **实现路线**：统一 `ContainerOps` 抽象；API 与 CLI 双实现；跟流经 Wails EventsEmit。

## 目标

- 在已连接的 Docker 工作区内，对单个容器执行启动、停止、重启、删除（可强制）。
- 打开右侧详情：概览展示 Inspect 摘要；日志可拉取尾部并可跟流/停止。
- API 与 CLI 回退路径行为对齐到同一前端模型；失败给出可读错误。
- 关闭 Docker 会话或应用退出时，取消进行中的日志跟流。

## 非目标（二期不做）

- exec / attach、镜像 / 网络 / 卷管理、Compose、K8s
- TCP TLS 证书完整 UI、Windows named pipe、Flutter 移动端 Docker
- 日志落盘、多容器并排日志、日志全文搜索/高亮
- 用用户名密码作为 Docker API 认证（延续一期）

## 操作模型

| 能力 | Engine API | CLI 回退 |
|------|------------|----------|
| Start / Stop / Restart | `POST /containers/{id}/{start\|stop\|restart}` | `docker start/stop/restart {id}` |
| Remove | `DELETE /containers/{id}?force=` | `docker rm [-f] {id}` |
| Inspect | `GET /containers/{id}/json` | `docker inspect {id}` → 同一摘要结构 |
| Logs（尾部） | `GET .../logs?tail=N&stdout=1&stderr=1` | `docker logs --tail N {id}` |
| Logs（跟流） | 同上 + `follow=1`，后端读流并 emit | `docker logs -f --tail N`，SSH 拉 stdout 并 emit |

**规则**

- 操作目标使用列表中的容器 ID（短 ID 可接受）。
- 启停重启成功后刷新列表；删除成功后若侧栏正显示该容器则关闭侧栏，再刷新。
- 删除前前端确认；默认非强制；可勾选「强制删除」或在「仍在运行」失败后二次确认强制删。
- 一次性命令超时复用现有 `dockerOpTimeout`；跟流不靠该超时结束，靠显式 Stop / 关会话。
- `CloseDocker` / `CloseAll` / 应用 `shutdown` 必须停止该会话全部 log follow。

## 数据模型

### ContainerDetail（侧栏概览）

至少包含（缺失则空字符串/空列表，不阻塞打开）：

- `id`、`name`、`image`、`state`、`status`、`created`
- `ports`（可读字符串或结构化均可，前端统一展示）
- `mounts`（源→目标路径列表）
- `network_mode`
- `env`、`cmd`、`entrypoint`（可截断；UI 可折叠）

与一期 `ContainerSummary` 并存：列表仍用 Summary；点开详情再 Inspect。

## UI

### 列表

- 行点击或「详情」打开侧栏并选中该容器。
- 右键或行末菜单：启动、停止、重启、删除、查看详情。
- 按状态禁用：运行中禁用启动；已停止禁用停止；删除始终可选（确认框）。

### 详情侧栏

- 主区表格保留；右侧滑出（默认宽约 360–420px，可拖）。
- 标题：名称 + 短 ID + 状态点；关闭按钮。
- **Tab 概览**：Inspect 摘要；Env/Cmd 折叠。
- **Tab 日志**：
  - 工具条：尾部行数（默认 200，可选 100/500/1000）、跟流开关、清空缓冲、复制可见内容。
  - 等宽只读区域；默认自动滚底；用户上滚则暂停自动滚（跟流仍接收）。
  - 进入日志 Tab 先 `GetDockerContainerLogs`；打开跟流再 `StartDockerContainerLogs`；关 Tab / 关侧栏 / 关会话 → `StopDockerContainerLogs`。

### 事件

- 推送：`docker:logs:{sessionID}:{containerID}`，payload 含文本块（如 `chunk`）。
- 前端节流拼接，避免高频 emit 卡死 UI。

## 后端 API（Wails 导出）

| 方法 | 职责 |
|------|------|
| `StartDockerContainer(sessionID, id)` | 启动 |
| `StopDockerContainer(sessionID, id)` | 停止 |
| `RestartDockerContainer(sessionID, id)` | 重启 |
| `RemoveDockerContainer(sessionID, id, force bool)` | 删除 |
| `InspectDockerContainer(sessionID, id) (ContainerDetail, error)` | 详情 |
| `GetDockerContainerLogs(sessionID, id, tail int) (string, error)` | 尾部 |
| `StartDockerContainerLogs(sessionID, id, tail int) error` | 开始跟流 |
| `StopDockerContainerLogs(sessionID, id) error` | 停止跟流 |

代码位置建议：

- `internal/docker/`：HTTP 容器操作与日志读取
- `internal/docker/cli_ops.go`（或同类）：CLI 命令与 inspect/logs 解析
- `internal/service/docker_service.go`：按 session mode 分支、跟流生命周期
- `app.go`：导出并 `EventsEmit` 日志块

## 与一期关系

- 复用既有 Docker 会话、`transport`、CLI 回退与 `DockerWorkspace` 列表。
- 不改变一期连接资产模型；不引入 moby SDK（延续薄 HTTP + CLI）。

## 验证（实现阶段）

- Go：API/CLI 各路径的启停删（可 mock）、Inspect 映射、尾部日志、跟流启动/停止与 Close 清理。
- 前端：菜单禁用态、删除确认、侧栏 Tab、跟流开关与取消订阅。
- 手工：本机 API、SSH 隧道 API、CLI 回退各测一轮启停与日志跟流。

## 风险与接受

| 风险 | 处理 |
|------|------|
| CLI `logs -f` 断线 | 依赖 Stop / CloseDocker 打断；UI 显示断开错误 |
| CLI Inspect 字段较少 | 侧栏允许空字段 |
| 强制删除 | 确认框 + force 勾选 |
| 日志事件量大 | 前端节流；后端可按行或块 emit |

## 后续（非本规格）

- exec / attach、镜像与网络卷、Compose、TLS UI、Windows 本机 pipe、移动端。
