# Docker 二期：容器操作、详情与日志 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在已有 Docker 会话上实现启停/重启/删除、Inspect 详情侧栏与日志尾部/跟流（API + CLI 双路径），不做 exec。

**Architecture:** 扩展 `internal/docker` 薄 HTTP 客户端与 CLI 命令封装为统一 ops；`DockerService` 按 session mode 分支并管理 log follow 生命周期；Wails EventsEmit `docker:logs:{sessionID}:{containerID}`；前端 `DockerWorkspace` 增加行菜单与右侧详情侧栏（概览 + 日志 Tab）。

**Tech Stack:** Go 1.24、Wails v2 Events、Svelte、`node --test`、`go test`、标准库 `net/http`（无 moby SDK）

**规格:** `docs/superpowers/specs/2026-10-23-docker-phase2-ops-logs-design.md`

## Global Constraints

- 文档正文中文；变更写 `docs/changes/features/`
- 不引入 `github.com/docker/docker`；延续薄 HTTP + CLI
- 不做 exec/attach、镜像/网络/卷、Compose、K8s、TLS 证书 UI、Windows pipe、移动端
- CLI 回退必须支持启停删与日志（与 API 同等）
- `CloseDocker` / `CloseAll` / `shutdown` 必须停止跟流
- 不改 Terminal 剪贴板受保护区；动到终端相关则跑契约三件套
- 工作在用户指定的当前分支（默认 master），不强制新建分支

## 文件结构

| 文件 | 职责 |
|------|------|
| `internal/docker/types.go` | 增加 `ContainerDetail`、日志相关类型 |
| `internal/docker/container_ops.go` | API：Start/Stop/Restart/Remove/Inspect/LogsReader |
| `internal/docker/cli_ops.go` | CLI：对应命令与 inspect/logs 解析 |
| `internal/docker/*_test.go` | 上述单测（httptest / 解析） |
| `internal/service/docker_service.go` | 会话级 ops + follow 状态机 |
| `internal/service/docker_service_test.go` | 注入 mock 覆盖 API/CLI/follow 清理 |
| `app.go` | 导出方法 + `emitDockerLogChunk` |
| `frontend/src/lib/dockerContainerActions.js` | 菜单禁用态、确认文案（可单测） |
| `frontend/src/components/DockerContainerDetail.svelte` | 侧栏概览 + 日志 Tab |
| `frontend/src/components/DockerWorkspace.svelte` | 行菜单、侧栏挂载、操作调用 |
| `frontend/wailsjs/go/main/App.js` / `App.d.ts` | 手工补绑定（与一期一致） |
| `docs/changes/features/2026-10-23-docker-phase2-ops-logs-impl.md` | 实现变更记录 |
| `.github/workflows/release.yml` | 更新说明补二期条目 |

---

### Task 1: ContainerDetail 类型 + Engine API 生命周期与 Inspect

**Files:**
- Modify: `internal/docker/types.go`
- Create: `internal/docker/container_ops.go`
- Create: `internal/docker/container_ops_test.go`

**Interfaces:**
- Produces:
  - `type ContainerDetail struct { ID, Name, Image, State, Status, Ports, NetworkMode string; Created int64; Mounts []string; Env []string; Cmd, Entrypoint []string }`
  - `(c *Client) StartContainer(ctx, id string) error`
  - `(c *Client) StopContainer(ctx, id string) error`
  - `(c *Client) RestartContainer(ctx, id string) error`
  - `(c *Client) RemoveContainer(ctx, id string, force bool) error`
  - `(c *Client) InspectContainer(ctx, id string) (ContainerDetail, error)`
- HTTP：`POST /containers/{id}/start|stop|restart`（204/304 视为成功）；`DELETE /containers/{id}?force=1`；`GET /containers/{id}/json` 映射到 Detail（Mounts 格式化为 `Source -> Destination`）

- [ ] **Step 1: 写失败测试**（httptest 断言路径与状态码；Inspect 映射 Name 去 `/`）

- [ ] **Step 2: `go test ./internal/docker -run Container -count=1` 确认失败**

- [ ] **Step 3: 最小实现**

- [ ] **Step 4: 测试通过**

- [ ] **Step 5: Commit** `feat(docker): add Engine API container lifecycle and inspect`

---

### Task 2: Engine API 日志（尾部 + 可读流）

**Files:**
- Modify: `internal/docker/container_ops.go`（或 `container_logs.go`）
- Modify: `internal/docker/container_ops_test.go`

**Interfaces:**
- Produces:
  - `(c *Client) GetContainerLogs(ctx, id string, tail int) (string, error)` — `GET .../logs?stdout=1&stderr=1&timestamps=0&tail=N`，解码 Docker multiplexed stream（header 8 字节）为纯文本
  - `(c *Client) FollowContainerLogs(ctx, id string, tail int) (io.ReadCloser, error)` — 同上 + `follow=1`；调用方负责 Close；ctx 取消应关闭 body

- [ ] **Step 1–4: TDD**（用固定 multiplex 字节测解码；follow 用可取消的 httptest）

- [ ] **Step 5: Commit** `feat(docker): add Engine API container log fetch and follow`

---

### Task 3: CLI ops（生命周期、inspect、logs）

**Files:**
- Create: `internal/docker/cli_ops.go`
- Create: `internal/docker/cli_ops_test.go`

**Interfaces:**
- Consumes: 能 `CombinedOutput(cmd string) ([]byte, error)` 的会话接口（与 service 层对齐）；follow 需要可取消的「长时间 CombinedOutput/管道」——在 service 层用 `ssh.Client` session 流式读时，本包可先提供：
  - `CLIStart/Stop/Restart/Remove(sess, id, force) error` 组装命令字符串并解析 exit
  - `CLIInspect(sess, id) (ContainerDetail, error)` 解析 `docker inspect` JSON 数组首元素
  - `CLILogsTail(sess, id, tail int) (string, error)`
- Follow 的进程管道可放 Task 4 service（调用 `docker logs -f`），本任务至少交付命令拼装纯函数 + inspect/tail 解析测试

- [ ] **Step 1–4: TDD**

- [ ] **Step 5: Commit** `feat(docker): add CLI container ops helpers`

---

### Task 4: DockerService 生命周期 + Inspect + 一次性 Logs

**Files:**
- Modify: `internal/service/docker_service.go`
- Modify: `internal/service/docker_service_test.go`
- Modify: `app.go`

**Interfaces:**
- Produces service + App 方法：
  - `StartDockerContainer` / `Stop` / `Restart` / `RemoveDockerContainer(sessionID, id, force)`
  - `InspectDockerContainer` → `ContainerDetail`
  - `GetDockerContainerLogs(sessionID, id, tail int) (string, error)`
- 按 `sess.mode` 分支 API vs CLI；未知 session 报错；操作后不自动 List（由前端刷新）

- [ ] **Step 1–4: TDD**（mock ssh / 可注入 api client 或在现有工厂上扩展）

- [ ] **Step 5: Commit** `feat(docker): wire container lifecycle and inspect in DockerService`

---

### Task 5: DockerService 日志跟流 + Close 清理

**Files:**
- Modify: `internal/service/docker_service.go`
- Modify: `app.go`（emit 回调）
- Modify: tests

**Interfaces:**
- `StartDockerContainerLogs(sessionID, id string, tail int, emit func(chunk string)) error`
- `StopDockerContainerLogs(sessionID, id string) error`
- 同一 session+container 重复 Start 先停旧的；`close()` / `CloseAll` 停全部 follow
- App：`StartDockerContainerLogs` 内 `runtime.EventsEmit(ctx, "docker:logs:"+sessionID+":"+id, map[string]string{"chunk": chunk})`

- [ ] **Step 1–4: TDD**（假 reader 推几块后 Stop；CloseDocker 后 emit 不再调用）

- [ ] **Step 5: Commit** `feat(docker): add container log follow with session cleanup`

---

### Task 6: 前端动作辅助 + 行菜单/确认

**Files:**
- Create: `frontend/src/lib/dockerContainerActions.js`
- Create: `frontend/test/dockerContainerActions.test.js`
- Modify: `frontend/src/components/DockerWorkspace.svelte`
- Modify: wailsjs stubs for new App methods

**Interfaces:**
- `canStart/canStop(container)`、`removeConfirmCopy(name, force)`、`containerActionMenu(container)`
- Workspace：右键或 ⋯ 菜单调用 Start/Stop/Restart/Remove；成功后 `refresh()`；删除用现有 ConfirmDialog 模式

- [ ] **Step 1–4: 单测 + UI 接线 + `node --test` + vite build（gradle 失败可接受）**

- [ ] **Step 5: Commit** `feat(docker): container row actions for start stop restart remove`

---

### Task 7: 详情侧栏（概览 + 日志 Tab）

**Files:**
- Create: `frontend/src/components/DockerContainerDetail.svelte`
- Modify: `DockerWorkspace.svelte`
- Optional: `frontend/src/lib/dockerLogs.js`（节流 append、自动滚底判定）

**Interfaces:**
- 点行打开侧栏 → Inspect；Tab 日志 → GetLogs；跟流开关 → Start/Stop + `EventsOn('docker:logs:...')`
- 关侧栏/组件 destroy → StopLogs；切换容器先停旧跟流

- [ ] **Step 1–4: 实现 + 单测纯函数 + build**

- [ ] **Step 5: Commit** `feat(docker): container detail sidebar with log follow`

---

### Task 8: 变更文档 + Release 说明

**Files:**
- Create: `docs/changes/features/2026-10-23-docker-phase2-ops-logs-impl.md`
- Modify: `.github/workflows/release.yml` 更新说明（新增：容器启停删、详情与日志跟流）

- [ ] **Step 1: 写中文变更文档（背景/范围/修改文件/验证/剩余风险）**

- [ ] **Step 2: 更新 release body**

- [ ] **Step 3: 回归**

```bash
go test ./internal/docker ./internal/service -count=1
cd frontend && node --test test/dockerContainerActions.test.js test/dockerWorkspace.test.js src/lib/workspaceTabs.test.js
```

- [ ] **Step 4: Commit** `docs(docker): phase2 ops/logs change log and release notes`

---

## Spec 覆盖自检

| 规格项 | 任务 |
|--------|------|
| Start/Stop/Restart/Remove API+CLI | T1, T3, T4, T6 |
| Inspect / ContainerDetail | T1, T3, T4, T7 |
| Logs 尾部 + 跟流 + 事件 | T2, T3, T5, T7 |
| Close 清理跟流 | T5 |
| 侧栏 UI / 菜单禁用 / 删除确认 | T6, T7 |
| 不做 exec 等 | 不实现 |
| 文档与 release | T8 |
