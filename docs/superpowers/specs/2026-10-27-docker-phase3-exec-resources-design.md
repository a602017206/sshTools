# Docker 三期：侧栏可拖、Exec 页签、镜像/网络/卷

## 背景

二期已交付容器启停删、详情侧栏与日志跟流。侧栏设计写明「可拖」但未实现，日志区过窄；运维下一步需要进容器（exec）以及镜像/网络/卷的浏览与清理。用户拍板：先可拖侧栏，再 exec 新页签，再资源只读+删除；不做 Compose / attach / pull / create。

## 已定口径

1. **交付顺序（三刀）**：① 详情侧栏可拖 → ② 容器 exec 新开终端页签 → ③ 镜像/网络/卷只读浏览+删除。
2. **详情展示**：可拖拽右侧侧栏（对齐数据库对象浏览器）；不做「详情新页签」。
3. **Exec**：顶栏新开完整 xterm 页签（`type: docker-exec`）；不做侧栏内嵌终端；默认 `/bin/sh`；本期无 shell 选择器。
4. **资源管理**：列表 + 简要详情 + 删除；不做 pull / create / run / prune 全家桶。
5. **CLI 回退**：与 API 会话同等支持本规格内写操作与 exec（与一/二期一致）。

## 目标

- 详情侧栏可拖宽，日志可读；宽度会话内记忆。
- 对运行中容器一键 Exec，获得与 SSH 终端同级的交互体验。
- 在同一 Docker 工作区切换查看并删除镜像、网络、卷。

## 非目标

- attach、自定义 exec 命令 UI、user/root 切换
- 镜像 pull / run、网络/卷 create、Compose、K8s
- TLS 证书完整 UI、Windows named pipe、Flutter 移动端 Docker
- 侧栏宽度写入 `config.json`（可后续加）
- 关 Docker 父会话时级联关闭已开的 exec 页签

---

## 刀 1：详情侧栏可拖

### 交互

- 侧栏左侧竖向拖动手柄（`role="separator"`，`aria-label` 说明用途）。
- 默认宽度 **420px**；拖拽范围 **320–720px**，且不超过主内容区约 **55%**。
- 宽度存在 `DockerWorkspace` 组件状态；关侧栏再开保留本次会话内宽度；**不落盘**。

### 实现

- 拖拽逻辑放在 `DockerWorkspace.svelte`；外层包住 `DockerContainerDetail` 并设 `style="width: …px"`。
- 去掉详情组件内写死的 `width: 400px` / `max-width: 45%`。
- 抽 `clampDockerDetailWidth(width, viewportWidth)` 纯函数便于单测。
- 不改日志跟流、Inspect、菜单行为。

### 验收

- 拖动手柄可连续改宽；松手后宽度稳定。
- 极值被夹紧；窄窗口不超过约 55%。
- 切换容器 / 关开侧栏后宽度仍为拖后值（同工作区实例内）。

---

## 刀 2：容器 Exec（新页签）

### 交互

- 行菜单与详情头增加「Exec」；**仅运行中**可点，否则禁用。
- 点击后顶栏新开页签：标题 `容器名 · exec`，副标题短 ID；同一容器允许多开。
- 复用现有 `Terminal.svelte`（输入、缩放、复制粘贴契约不变）。
- 关 exec 页签 → `CloseDockerExec`；关 Docker 工作区页签 **不**强制关闭已开 exec。

### 会话模型（前端）

- `connectionsStore` 增加 `type: 'docker-exec'`。
- 字段至少：`sessionId`（exec 会话 ID）、`parentDockerSessionId`，`containerId`，`connection`（展示用，可复用 Docker 资产摘要），`connected`。
- `TerminalPanel`：该类型渲染 Terminal；数据通道走 Docker Exec bindings，而非 SSH。

### 后端 API（Wails）

| 方法 | 职责 |
|------|------|
| `StartDockerExec(dockerSessionID, containerID, cols, rows) (execSessionID, error)` | 创建并启动交互式 exec |
| `SendDockerExecData(execSessionID, data)` | 写入 stdin |
| `ResizeDockerExec(execSessionID, cols, rows)` | 调整 TTY |
| `CloseDockerExec(execSessionID)` | 结束 exec |

- 输出事件：`docker:exec:{execSessionID}`，payload 为文本块（对齐 `ssh:output` 用法）。
- 默认命令：`/bin/sh`（`-i` 交互）；失败返回可读错误，前端不建页签或建后立即显示错误并关闭。

### 传输路径

| 模式 | 行为 |
|------|------|
| Engine API | `POST /containers/{id}/exec`（AttachStdin/Stdout/Stderr + Tty）→ `POST /exec/{id}/start` 劫持流；resize 走 exec resize API |
| CLI 回退 | 本机或既有 Docker SSH 会话上 `docker exec -i -t {id} /bin/sh`，PTY 读写；resize 映射到 PTY |

- Exec 生命周期独立于日志 follow；`CloseDocker` 应停止该 Docker 会话下仍登记的 exec（若实现上 exec 挂在 dockerSession 内），但前端不依赖「关父页签自动关子页签」的 UX。
- 应用 shutdown / `CloseAll` 必须清理未关的 exec。

### 验收

- API 与 CLI 回退均可进入 shell，输入回显、resize 正常。
- 未运行容器菜单项禁用；无 shell / 权限错误有可读提示。
- 关 exec 页签后后端流结束，无泄漏 goroutine（单测或手工观察）。

---

## 刀 3：镜像 / 网络 / 卷（只读 + 删除）

### 交互

- Docker 工作区顶部资源 Tab：`容器 | 镜像 | 网络 | 卷`（默认容器）。
- 各 Tab：表格 + 刷新；行选中后右侧可拖详情栏展示概览（无日志 Tab）。
- 删除前确认；镜像可勾选「强制」；网络/卷删除失败展示引擎错误（如仍被引用）。

### 列表字段

| 资源 | 字段 |
|------|------|
| 镜像 | ID、Tags、Size、Created |
| 网络 | ID、Name、Driver、Scope |
| 卷 | Name、Driver、Mountpoint、Created |

### 后端 API（Wails）

| 方法 | 职责 |
|------|------|
| `ListDockerImages/Networks/Volumes(sessionID)` | 列表 |
| `InspectDockerImage/Network/Volume(sessionID, id)` | 详情（侧栏；列表字段不足时再调） |
| `RemoveDockerImage(sessionID, id, force)` | 删镜像 |
| `RemoveDockerNetwork(sessionID, id)` | 删网络 |
| `RemoveDockerVolume(sessionID, id)` | 删卷 |

- API：对应 Engine 列表/Inspect/删除端点。
- CLI：`docker images` / `network ls` / `volume ls` 与 `rmi` / `network rm` / `volume rm`（及 `-f` 按需）。

### 验收

- 三种资源可列出、刷新、打开详情、删除（含失败提示）。
- API 与 CLI 回退字段映射到同一前端模型；缺失字段显示为 `-`。

---

## 与一/二期关系

- 复用 Docker 会话、`transport`、CLI 回退、`DockerWorkspace`、二期 `ContainerOps` 模式。
- 不引入 moby SDK；延续薄 HTTP + CLI。
- 不改动终端剪贴板受保护区；exec 页签复用 Terminal 时遵守既有契约测试。

## 风险与接受

| 风险 | 处理 |
|------|------|
| `/bin/sh` 不存在 | 可读错误；后续再加 shell 选择 |
| API hijack 流与 Wails 事件模型复杂 | 对齐现有 SSH/日志 follow 模式；单测 mock |
| CLI `docker exec -t` 与本机/SSH PTY | 复用已有 local/SSH PTY 能力；平台差异记入实现说明 |
| 删网络/卷被占用 | 展示引擎错误，不静默 |
| 侧栏过宽挤列表 | 上限 720px 与 55% 视口约束 |

## 文档与实现节奏

1. 本规格评审通过后：先写刀 1 实现计划并落地；再刀 2、刀 3 各自计划。
2. 变更记录：功能进 `docs/changes/features/`；实现说明进 `docs/development/`。
3. 设计副本：`docs/designs/2026-10-27-docker-phase3-exec-resources-design.md`。
