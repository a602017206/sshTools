# Docker 容器连接 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 实现 Docker 三种接入（本机 / SSH 隧道+CLI 回退 / TCP）的测连、保存、会话与容器列表工作区。

**Architecture:** 新增 `internal/docker` 薄 Engine API HTTP 客户端（Unix socket / TCP，不引入 moby SDK）；`SSH` 路径用 `golang.org/x/crypto/ssh` 的 `unix` Dial + 本地 TCP 转发；`DockerService` 管会话；前端改 `AddAssetDialog` 按 `transport` 显隐字段，新增 `DockerWorkspace`，顶栏增加 `docker` 模式。

**Tech Stack:** Go 1.24、Wails v2、Svelte、Node `node --test`、`go test`、标准库 `net/http`（Docker API）

**规格:** `docs/superpowers/specs/2026-10-23-docker-connection-design.md`

## Global Constraints

- 文档正文中文；变更写 `docs/changes/features/`
- 不内嵌 Docker 引擎；不增加 `github.com/docker/docker` 依赖（第一期用 HTTP）
- Docker 表单不做「用户名+密码」API 认证
- 第一期只做 Ping + ListContainers；不做日志/exec/启停
- 动到终端相关文件时必须跑剪贴板契约三件套；本计划默认不改 Terminal 受保护区
- SSH 隧道会话结束必须 Close，禁止端口泄漏

## 文件结构

| 文件 | 职责 |
|------|------|
| `internal/docker/types.go` | `Config`、`ContainerSummary`、`Transport` 常量 |
| `internal/docker/client.go` | Engine API：`Ping`、`ListContainers`（HTTP over Unix/TCP） |
| `internal/docker/cli_parse.go` | 解析 `docker ps --format '{{json .}}'` 行 |
| `internal/docker/tunnel.go` | 本地 TCP → 远端 unix socket 转发 |
| `internal/docker/*_test.go` | 上述单测 |
| `internal/ssh/client.go` | 导出 `SSHClient()` 或 `Dial(n,addr)` 供隧道/CLI |
| `internal/service/docker_service.go` | 会话：Connect/Close/Test/DetectLocal/List |
| `app.go` | 装配服务并导出 Wails API |
| `frontend/src/lib/dockerTransport.js` | transport 校验与默认 metadata（可单测） |
| `frontend/src/lib/workspaceTabs.js` | `docker` 独立模式；`modeForAsset` |
| `frontend/src/components/AddAssetDialog.svelte` | Docker 接入方式 UI |
| `frontend/src/components/DockerWorkspace.svelte` | 容器列表工作区 |
| `frontend/src/components/TerminalPanel.svelte` / `App.svelte` | 打开 docker 资产、挂载工作区 |
| `frontend/src/components/AssetList.svelte` | SSH 右键「添加 Docker」、本机提示入口 |
| `docs/changes/features/2026-10-23-docker-connection-impl.md` | 实现变更记录 |

---

### Task 1: Docker 类型与 Engine API 客户端（本机/TCP）

**Files:**
- Create: `internal/docker/types.go`
- Create: `internal/docker/client.go`
- Create: `internal/docker/client_test.go`

**Interfaces:**
- Produces:
  - `const TransportLocal/TransportSSH/TransportTCP = "local"|"ssh"|"tcp"`
  - `type Config struct { Transport, Host string; Port int; SocketPath, DockerHost string; TLS bool }`
  - `type ContainerSummary struct { ID, Name, Image, Status, Ports string; Created int64 }`
  - `type Client struct{...}`；`NewClient(cfg Config) (*Client, error)`；`(c *Client) Ping(ctx) error`；`(c *Client) ListContainers(ctx context.Context, all bool) ([]ContainerSummary, error)`；`(c *Client) Close() error`
- `NewClient`：`local` → 默认 socket（`DOCKER_HOST` / `SocketPath` / 平台默认）；`tcp` → `http://host:port`；`ssh` 本任务不实现（由 Task 3 注入 `BaseURL` 或 `httpClient`）

- [ ] **Step 1: 写失败测试**

```go
package docker_test

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"AHaSSHTools/internal/docker"
)

func TestPingAndListContainersOverHTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/_ping":
			w.Header().Set("API-Version", "1.44")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		case "/containers/json":
			if r.URL.Query().Get("all") != "1" {
				t.Fatalf("expected all=1, got %q", r.URL.Query().Get("all"))
			}
			_ = json.NewEncoder(w).Encode([]map[string]any{
				{
					"Id":      "abcdef0123456789",
					"Names":   []string{"/web"},
					"Image":   "nginx:latest",
					"Status":  "Up 2 hours",
					"Created": int64(1700000000),
					"Ports":   []map[string]any{{"PublicPort": 8080, "PrivatePort": 80, "Type": "tcp"}},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	host, port := mustHostPort(t, srv.URL)
	c, err := docker.NewClient(docker.Config{Transport: docker.TransportTCP, Host: host, Port: port})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if err := c.Ping(context.Background()); err != nil {
		t.Fatal(err)
	}
	list, err := c.ListContainers(context.Background(), true)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Name != "web" || list[0].Image != "nginx:latest" {
		t.Fatalf("unexpected list: %+v", list)
	}
}

func mustHostPort(t *testing.T, raw string) (string, int) {
	t.Helper()
	u := raw // httptest URL like http://127.0.0.1:port
	host, portStr, err := net.SplitHostPort(u[len("http://"):])
	if err != nil {
		t.Fatal(err)
	}
	var port int
	_, _ = fmt.Sscanf(portStr, "%d", &port)
	return host, port
}
```

（实现时把 `fmt` 导入补全；或改用 `url.Parse` + `strconv.Atoi`。）

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./internal/docker -count=1`
Expected: FAIL（包不存在或符号未定义）

- [ ] **Step 3: 最小实现**

`types.go`：常量与结构体如上。

`client.go` 要点：
- TCP：`http.Client` 默认 Transport，Base `http://host:port`
- local：自定义 `http.Transport{DialContext: unix dial}`，Base `http://localhost`（Docker 约定）
- `GET /_ping`；`GET /containers/json?all=1|0`
- 映射 Names 去掉前导 `/`；Ports 格式化为 `8080:80/tcp` 一类可读字符串；短 ID 取前 12 位存 `ID`

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./internal/docker -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/docker/types.go internal/docker/client.go internal/docker/client_test.go
git commit -m "$(cat <<'EOF'
feat(docker): add Engine API client for local and TCP

EOF
)"
```

---

### Task 2: CLI 输出解析（SSH 回退）

**Files:**
- Create: `internal/docker/cli_parse.go`
- Create: `internal/docker/cli_parse_test.go`

**Interfaces:**
- Consumes: `ContainerSummary`
- Produces: `ParseDockerPSJSONLines(stdout string) ([]ContainerSummary, error)` — 按行 JSON；字段兼容 `ID`/`Id`、`Names`/`Names` 字符串、`Image`、`Status`、`Ports`、`CreatedAt`

- [ ] **Step 1: 写失败测试**

```go
func TestParseDockerPSJSONLines(t *testing.T) {
	in := `{"ID":"abc123def456","Names":"web","Image":"nginx","Status":"Up 1 minute","Ports":"80/tcp","CreatedAt":"2024-01-02 03:04:05 +0000 UTC"}
{"ID":"fff111aaa222","Names":"db","Image":"postgres","Status":"Exited (0)","Ports":"","CreatedAt":"2024-01-01 00:00:00 +0000 UTC"}`
	list, err := docker.ParseDockerPSJSONLines(in)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 2 || list[0].Name != "web" || list[1].Status == "" {
		t.Fatalf("%+v", list)
	}
}

func TestParseDockerPSJSONLinesEmpty(t *testing.T) {
	list, err := docker.ParseDockerPSJSONLines("\n")
	if err != nil || len(list) != 0 {
		t.Fatalf("list=%v err=%v", list, err)
	}
}
```

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./internal/docker -run ParseDockerPS -count=1`
Expected: FAIL

- [ ] **Step 3: 实现 `ParseDockerPSJSONLines`**

跳过空行；单行 `json.Unmarshal` 到临时 struct；`Created` 解析失败则置 `0`（不致整批失败）。

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./internal/docker -run ParseDockerPS -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/docker/cli_parse.go internal/docker/cli_parse_test.go
git commit -m "$(cat <<'EOF'
feat(docker): parse docker ps JSON lines for SSH CLI fallback

EOF
)"
```

---

### Task 3: SSH Unix 隧道 + SSH Client 导出

**Files:**
- Modify: `internal/ssh/client.go`（增加 `Underlying() *cryptossh.Client` 与/或 `Dial(network, addr string) (net.Conn, error)`）
- Create: `internal/docker/tunnel.go`
- Create: `internal/docker/tunnel_test.go`

**Interfaces:**
- Consumes: 已连接的 `*ssh.Client`（本仓库）或 `interface { Dial(network, addr string) (net.Conn, error); Close() error }`
- Produces:
  - `type UnixTunnel struct{...}`
  - `StartUnixTunnel(dialer UnixDialer, remoteSocket string) (localAddr string, tun *UnixTunnel, err error)`
  - `(t *UnixTunnel) Close() error`
  - 行为：`Listen("tcp", "127.0.0.1:0")`，每条连接 `dialer.Dial("unix", remoteSocket)` 后双向 `io.Copy`

- [ ] **Step 1: 写失败测试（本地假 dialer）**

```go
func TestUnixTunnelForwardsToDialer(t *testing.T) {
	// 起一个本地 unix 服务端 echo
	dir := t.TempDir()
	sock := filepath.Join(dir, "docker.sock")
	ln, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		c, _ := ln.Accept()
		if c != nil {
			defer c.Close()
			buf := make([]byte, 4)
			_, _ = io.ReadFull(c, buf)
			_, _ = c.Write(buf)
		}
	}()

	dialer := unixDialerFunc(func(network, addr string) (net.Conn, error) {
		return net.Dial(network, addr)
	})
	addr, tun, err := docker.StartUnixTunnel(dialer, sock)
	if err != nil {
		t.Fatal(err)
	}
	defer tun.Close()

	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_, _ = conn.Write([]byte("ping"))
	out := make([]byte, 4)
	if _, err := io.ReadFull(conn, out); err != nil {
		t.Fatal(err)
	}
	if string(out) != "ping" {
		t.Fatalf("got %q", out)
	}
}
```

（`unixDialerFunc` 为测试辅助类型，实现 `Dial`。）

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./internal/docker -run UnixTunnel -count=1`
Expected: FAIL

- [ ] **Step 3: 实现隧道 + SSH `Dial` 导出**

```go
// internal/ssh/client.go
func (c *Client) Dial(network, addr string) (net.Conn, error) {
	if c.client == nil {
		return nil, fmt.Errorf("ssh client not connected")
	}
	return c.client.Dial(network, addr)
}
```

`StartUnixTunnel` 在后台 accept loop；`Close` 关闭 listener 与 active conns。

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./internal/docker -run UnixTunnel -count=1`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/ssh/client.go internal/docker/tunnel.go internal/docker/tunnel_test.go
git commit -m "$(cat <<'EOF'
feat(docker): add SSH unix socket local forward tunnel

EOF
)"
```

---

### Task 4: DockerService（Detect / Test / Connect / List / 回退）

**Files:**
- Create: `internal/service/docker_service.go`
- Create: `internal/service/docker_service_test.go`
- Modify: `app.go`（字段、startup 装配）

**Interfaces:**
- Consumes: `docker.Client`、`docker.StartUnixTunnel`、`docker.ParseDockerPSJSONLines`、`config.ConfigManager`、`ssh.NewClient`、凭据读取（与现有 SSH 相同）
- Produces:
  - `type DockerConnectOptions` / 从 `config.ConnectionConfig` 解析 metadata
  - `DetectLocalDocker() (ok bool, detail string, err error)`
  - `TestDockerConnection(conn config.ConnectionConfig) error`
  - `ConnectDocker(sessionID string, conn config.ConnectionConfig) error`
  - `CloseDocker(sessionID string) error`
  - `ListDockerContainers(sessionID string, all bool) ([]docker.ContainerSummary, error)`
  - 会话状态含：`mode API|CLI`、可选 `tunnel`、`apiClient`

解析规则：
- `metadata.transport` 缺省：有 `ssh_connection_id` → `ssh`；有 host → `tcp`；否则 `local`
- `cli_fallback` 缺省 `true`（字符串 `"false"` 才关）
- `ssh`：加载 SSH 资产 → Connect → StartUnixTunnel(`remote_socket` 默认 `/var/run/docker.sock`) → `NewClient` 指向隧道 `127.0.0.1:port`；Ping 失败且允许回退 → 标记 CLI，`ssh session` 执行 `docker ps -a --format '{{json .}}'`
- `CloseDocker`：关 API client、关 tunnel、关专用 SSH client（勿误关用户终端会话；Docker 使用独立 SSH client）

- [ ] **Step 1: 写失败测试（CLI 回退路径用假 runner）**

将 SSH/CLI 依赖做成可注入接口，例如：

```go
type sshFactory func(cfg *ssh.Config) (dockerSSHSession, error)

type dockerSSHSession interface {
	Dial(network, addr string) (net.Conn, error)
	CombinedOutput(cmd string) ([]byte, error)
	Close() error
}
```

测试：`Dial` 返回 error → 应回退 CLI → `CombinedOutput` 返回一行 JSON → List 成功。  
另测：`cli_fallback=false` 时 Dial 失败 → Test/Connect 返回错误。

- [ ] **Step 2: 跑测试确认失败**

Run: `go test ./internal/service -run Docker -count=1`
Expected: FAIL

- [ ] **Step 3: 实现 `DockerService` 并在 `app.go` 挂载**

```go
// app.go 示意
func (a *App) DetectLocalDocker() (map[string]interface{}, error) { ... }
func (a *App) TestDockerConnection(conn config.ConnectionConfig) error { ... }
func (a *App) ConnectDocker(sessionID string, conn config.ConnectionConfig) error { ... }
func (a *App) CloseDocker(sessionID string) error { ... }
func (a *App) ListDockerContainers(sessionID string, all bool) ([]docker.ContainerSummary, error) { ... }
```

Wails 导出类型需可 JSON 序列化（`ContainerSummary` 已有导出字段）。

- [ ] **Step 4: 跑测试确认通过**

Run: `go test ./internal/service -run Docker -count=1`  
Run: `go test ./internal/docker ./internal/service -count=1`  
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/service/docker_service.go internal/service/docker_service_test.go app.go
git commit -m "$(cat <<'EOF'
feat(docker): add DockerService with SSH tunnel and CLI fallback

EOF
)"
```

---

### Task 5: 前端 transport 辅助 + 工作区模式

**Files:**
- Create: `frontend/src/lib/dockerTransport.js`
- Create: `frontend/test/dockerTransport.test.js`
- Modify: `frontend/src/lib/workspaceTabs.js`
- Modify: `frontend/src/lib/workspaceTabs.test.js`
- Modify: `frontend/src/stores.js`（若 `APP_MODES` 从别处导入则只改 workspaceTabs）

**Interfaces:**
- Produces:
  - `buildDockerMetadata({ transport, sshConnectionId, socketPath, remoteSocket, cliFallback, tls })`
  - `defaultDockerName(sshAsset)` → `` `${sshAsset.name}-Docker` ``
  - `resolveMode('docker') === 'docker'`；`modeForAsset({type:'docker'}) === 'docker'`
  - `APP_MODES` 增加 `{ id: 'docker', label: '容器' }`（或「Docker」）
  - `sessionMatchesMode`：`type==='docker'` 仅在 docker 模式显示

- [ ] **Step 1: 写失败测试**

```js
import test from 'node:test';
import assert from 'node:assert/strict';
import { buildDockerMetadata, defaultDockerName } from '../src/lib/dockerTransport.js';
import { resolveMode, modeForAsset, sessionMatchesMode } from '../src/lib/workspaceTabs.js';

test('buildDockerMetadata local', () => {
  const m = buildDockerMetadata({ transport: 'local' });
  assert.equal(m.transport, 'local');
  assert.equal(m.domain, 'docker');
});

test('docker is its own mode', () => {
  assert.equal(resolveMode('docker'), 'docker');
  assert.equal(modeForAsset({ type: 'docker' }), 'docker');
  assert.equal(sessionMatchesMode({ type: 'docker' }, 'docker'), true);
  assert.equal(sessionMatchesMode({ type: 'docker' }, 'ssh'), false);
});
```

- [ ] **Step 2: 跑测试确认失败**

Run: `cd frontend && node --test test/dockerTransport.test.js src/lib/workspaceTabs.test.js`
Expected: FAIL

- [ ] **Step 3: 实现并改 `workspaceTabs.js`**

从 `legacyModeMap` **删除** `docker: 'ssh'`（或改为映射到自身）；把 `docker` 加入 `APP_MODES`。

- [ ] **Step 4: 跑测试确认通过**

Run: `cd frontend && node --test test/dockerTransport.test.js src/lib/workspaceTabs.test.js`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add frontend/src/lib/dockerTransport.js frontend/test/dockerTransport.test.js frontend/src/lib/workspaceTabs.js frontend/src/lib/workspaceTabs.test.js
git commit -m "$(cat <<'EOF'
feat(docker): add transport helpers and docker workspace mode

EOF
)"
```

---

### Task 6: AddAssetDialog Docker 表单

**Files:**
- Modify: `frontend/src/components/AddAssetDialog.svelte`
- Modify: `frontend/test/connectionFormData.test.js`（若有 docker 克隆/默认端口断言则更新）

**Interfaces:**
- Consumes: `buildDockerMetadata`、`GetConnections`（筛 `type==='ssh'`）、`TestDockerConnection`、`DetectLocalDocker`（可选预填）
- 行为：
  - `assetType==='docker'` 时展示接入方式三选一：`local|ssh|tcp`
  - local：隐藏 host/port/user/password；只名称+分组
  - ssh：SSH 下拉；隐藏 docker 密码字段
  - tcp：host + port（默认 2375）；无用户名密码
  - 提交 `type:'docker'` + metadata；`TestConnection` 按钮对 docker 调 `TestDockerConnection`
  - 打开参数支持 `defaults: { assetType:'docker', transport:'ssh', sshConnectionId }`

- [ ] **Step 1: 手工/轻量断言**

若有纯函数抽离表单 payload，补单测；否则本任务以实现 + `npm run build` 验证。

- [ ] **Step 2: 改 `AddAssetDialog.svelte`**

移除 docker 分支对 `requiresPassword` / 默认密码认证的依赖；`getDefaultPortFor('docker')` 仅 tcp 使用。

- [ ] **Step 3: 构建**

Run: `cd frontend && npm run build`  
Expected: 成功

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/AddAssetDialog.svelte frontend/test/connectionFormData.test.js
git commit -m "$(cat <<'EOF'
feat(docker): transport-aware add connection form

EOF
)"
```

---

### Task 7: DockerWorkspace + 连接打开流

**Files:**
- Create: `frontend/src/components/DockerWorkspace.svelte`
- Modify: `frontend/src/components/TerminalPanel.svelte`（或 `App.svelte`：按 mode 挂载）
- Modify: `frontend/src/App.svelte`（`handleConnect` / 顶栏 mode）

**Interfaces:**
- Consumes: `ConnectDocker`、`ListDockerContainers`、`CloseDocker`
- UI：顶栏名称 + transport 徽章 + 刷新 + 「仅运行中」筛选；表格列按规格
- 空态：未连接 / 无容器；若后端返回 `via: cli`（可选字段）显示小字「CLI 回退」
- `handleConnect(dockerAsset)`：`activeMode='docker'`，创建 session `{ id, type:'docker', connection }`，调用 `ConnectDocker` 后 `ListDockerContainers(true)`

- [ ] **Step 1: 实现 `DockerWorkspace.svelte`**

保持视觉与现有 database 空态/表格风格一致（ops-* class），避免新设计体系。

- [ ] **Step 2: 接线 App / TerminalPanel**

`mode === 'docker'` 时渲染 `DockerWorkspace`，不要把 docker 资产送进 SSH `ConnectSSH`。

- [ ] **Step 3: 构建 + 剪贴板契约（若碰 Terminal）**

Run: `cd frontend && npm run build`  
若改了 `Terminal.svelte` / shortcuts：  
`cd frontend && node --test test/terminalClipboardContract.test.js test/terminalSelection.test.js test/terminalShortcuts.test.js`  
Expected: PASS（本任务应避免改受保护区）

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/DockerWorkspace.svelte frontend/src/components/TerminalPanel.svelte frontend/src/App.svelte
git commit -m "$(cat <<'EOF'
feat(docker): add DockerWorkspace and connect flow

EOF
)"
```

---

### Task 8: 本机探测提示 + SSH 右键添加 + 变更文档

**Files:**
- Modify: `frontend/src/components/AssetList.svelte`（及上下文菜单数据源）
- Modify: `frontend/src/App.svelte`（打开添加对话框的 defaults）
- Create: `docs/changes/features/2026-10-23-docker-connection-impl.md`
- Modify: `docs/development/` 可选短实现说明（若仓库惯例需要）

**Interfaces:**
- Docker 域空态或顶部：若 `DetectLocalDocker().ok` 且列表中尚无 `transport=local` 资产 → 显示「添加本机 Docker」按钮 → `openAddConnection({ assetType:'docker', transport:'local' })`
- SSH 资产右键菜单增加「添加 Docker」→ defaults `transport:'ssh', sshConnectionId: asset.id`
- 变更文档含：背景、范围、修改文件、验证、剩余风险

- [ ] **Step 1: 实现入口 UI**

- [ ] **Step 2: 写变更文档（中文）**

- [ ] **Step 3: 回归命令**

```bash
go test ./internal/docker ./internal/service -count=1
cd frontend && node --test test/dockerTransport.test.js src/lib/workspaceTabs.test.js
cd frontend && npm run build
```

Expected: 全部 PASS / build OK

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/AssetList.svelte frontend/src/App.svelte docs/changes/features/2026-10-23-docker-connection-impl.md
git commit -m "$(cat <<'EOF'
feat(docker): local detect prompt, SSH context add, and change log

EOF
)"
```

---

## Spec 覆盖自检

| 规格项 | 任务 |
|--------|------|
| transport local/ssh/tcp 资产模型 | T4 解析 + T5/T6 表单 |
| 本机 socket / DOCKER_HOST | T1 + T4 Detect |
| SSH 隧道优先 | T3 + T4 |
| CLI 回退与可关闭 | T2 + T4 |
| TCP 无密码、2375 | T1 + T6 |
| Detect 后确认添加 | T8 |
| 侧栏新建 + SSH 右键 | T6 + T8 |
| DockerWorkspace 列表 | T7 |
| 非目标（日志/exec/TLS 完整） | 不实现 |
| 不映射 docker→ssh workspace | T5 |

## Placeholder 扫描

无 TBD/「类似 Task N」占位；类型名在任务间一致：`ContainerSummary`、`Transport*`、`StartUnixTunnel`、`ParseDockerPSJSONLines`、`DockerService` API。
