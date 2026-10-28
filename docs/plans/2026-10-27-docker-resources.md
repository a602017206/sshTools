# Docker 镜像 / 网络 / 卷（只读 + 删除）Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在已连接的 Docker 工作区内切换浏览镜像、网络、卷，支持刷新、详情侧栏与删除（镜像可强制）；API 与 CLI 回退双路径。

**Architecture:** 扩展 `internal/docker` 薄 HTTP 列表/Inspect/删除 + CLI `images`/`network`/`volume` 命令解析；`DockerService` 按 session mode 分支（对齐 `runContainerOp`）；Wails 导出；`DockerWorkspace` 顶部资源 Tab，复用可拖详情壳，详情无日志/Exec。

**Tech Stack:** Go、Wails、Svelte、`go test`、`node --test`、标准库 `net/http`（无 moby SDK）

**规格:** `docs/superpowers/specs/2026-10-27-docker-phase3-exec-resources-design.md`（刀 3）

## Global Constraints

- 只读浏览 + 删除；不做 pull / create / run / prune / Compose
- API 与 CLI 字段映射到同一前端模型；缺失显示 `-`
- 删除前确认；镜像可勾选强制；网络/卷失败展示引擎错误
- 复用既有 Docker 会话与 CLI 回退；不引入 moby SDK
- 不改 Terminal 剪贴板受保护区；不改 exec 行为
- 文档中文；变更写 `docs/changes/features/`；工作在当前 `master`
- 本计划**仅刀 3**

## 文件结构

| 文件 | 职责 |
|------|------|
| `internal/docker/types.go` | `ImageSummary`/`ImageDetail`、`NetworkSummary`/`NetworkDetail`、`VolumeSummary`/`VolumeDetail` |
| `internal/docker/resource_ops.go` | Engine API：List/Inspect/Remove images|networks|volumes |
| `internal/docker/resource_ops_test.go` | httptest 单测 |
| `internal/docker/cli_resources.go` | CLI 命令构建与 JSON 行解析 |
| `internal/docker/cli_resources_test.go` | 解析/转义单测 |
| `internal/service/docker_resources.go` | Service 层 List/Inspect/Remove |
| `internal/service/docker_resources_test.go` | API/CLI mock |
| `app.go` + `frontend/wailsjs/go/main/App.js` / `App.d.ts` | 导出与绑定 |
| `frontend/src/lib/dockerResources.js` | Tab 定义、字段取值、删除确认文案、Size 格式化 |
| `frontend/src/components/DockerResourceDetail.svelte` | 通用资源概览侧栏（无日志） |
| `frontend/src/components/DockerWorkspace.svelte` | 资源 Tab + 列表切换 + 删除 |
| `docs/changes/features/2026-10-27-docker-resources-impl.md` | 实现变更记录 |
| `.github/workflows/release.yml` | 更新说明加资源浏览一句 |

### 数据模型（跨任务约定）

```go
type ImageSummary struct {
    ID      string   // 短或长 ID
    Tags    []string // RepoTags；可空
    Size    int64    // 字节
    Created int64    // unix 秒
    Via     string   `json:"via,omitempty"`
}

type ImageDetail struct {
    ImageSummary
    Architecture string
    Os           string
    Author       string
    Comment      string
    // 可选：History 不做本期
}

type NetworkSummary struct {
    ID     string
    Name   string
    Driver string
    Scope  string
    Via    string `json:"via,omitempty"`
}

type NetworkDetail struct {
    NetworkSummary
    Internal   bool
    Attachable bool
    Ingress    bool
    // IPAM 简化为一行字符串即可
    IPAM string
}

type VolumeSummary struct {
    Name       string
    Driver     string
    Mountpoint string
    Created    int64 // unix 秒；CLI 可能为 0
    Via        string `json:"via,omitempty"`
}

type VolumeDetail struct {
    VolumeSummary
    Labels map[string]string // 可空；前端折叠展示
    Scope  string
}
```

### Wails API

| 方法 | 说明 |
|------|------|
| `ListDockerImages(sessionID) ([]ImageSummary, error)` | |
| `InspectDockerImage(sessionID, idOrRef string) (ImageDetail, error)` | id 或 `repo:tag` |
| `RemoveDockerImage(sessionID, idOrRef string, force bool) error` | |
| `ListDockerNetworks(sessionID) ([]NetworkSummary, error)` | |
| `InspectDockerNetwork(sessionID, idOrName string) (NetworkDetail, error)` | |
| `RemoveDockerNetwork(sessionID, idOrName string) error` | |
| `ListDockerVolumes(sessionID) ([]VolumeSummary, error)` | |
| `InspectDockerVolume(sessionID, name string) (VolumeDetail, error)` | |
| `RemoveDockerVolume(sessionID, name string) error` | |

### Engine HTTP（参考）

- Images：`GET /images/json`；`GET /images/{name}/json`；`DELETE /images/{name}?force=1`
- Networks：`GET /networks`；`GET /networks/{id}`；`DELETE /networks/{id}`
- Volumes：`GET /volumes`（响应 `{Volumes:[...]}`）；`GET /volumes/{name}`；`DELETE /volumes/{name}`

### CLI（参考）

- `docker images --format '{{json .}}'` → 解析 `ID/Repository/Tag/Size/CreatedAt`（Size 可能是人类可读，需解析或改为 `--format` 含 `Size` 数字；优先 API 字段；CLI 可用 `docker image inspect` 补 Size，或接受 Size 字符串转估——**推荐**：CLI list 用 `{{json .}}`，`Size` 字段若为人读字符串则解析 `GB/MB/kB` 粗转字节，失败则为 0）
- `docker network ls --format '{{json .}}'`
- `docker volume ls --format '{{json .}}'`
- Inspect：`docker image/network/volume inspect <id>`
- Remove：`docker rmi [-f]` / `docker network rm` / `docker volume rm`（均 `shellQuote`）

ID 校验：镜像允许 `:` `/` `@` 等引用字符；网络/卷名用较宽字符集 + `shellQuote`。抽出 `validateDockerRef(kind, id)` 于 service 层。

---

### Task 1: 类型 + Engine API 镜像

**Files:**
- Modify: `internal/docker/types.go`
- Create: `internal/docker/resource_ops.go`
- Create: `internal/docker/resource_ops_test.go`

**Interfaces:**
- `ListImages(ctx) ([]ImageSummary, error)`
- `InspectImage(ctx, ref string) (ImageDetail, error)`
- `RemoveImage(ctx, ref string, force bool) error`

- [ ] **Step 1:** httptest：list 映射 `Id/RepoTags/Size/Created`；inspect；delete `?force=1`

- [ ] **Step 2:** `go test ./internal/docker -run Image -count=1` FAIL

- [ ] **Step 3:** 实现

- [ ] **Step 4:** PASS

- [ ] **Step 5: Commit** `feat(docker): add Engine API image list inspect remove`

---

### Task 2: Engine API 网络 + 卷

**Files:**
- Modify: `internal/docker/types.go`（若 Task 1 未加全）
- Modify: `internal/docker/resource_ops.go`
- Modify: `internal/docker/resource_ops_test.go`

**Interfaces:**
- `ListNetworks` / `InspectNetwork` / `RemoveNetwork`
- `ListVolumes` / `InspectVolume` / `RemoveVolume`（注意 list 响应包在 `Volumes` 键下）

- [ ] **Step 1–4:** TDD httptest

- [ ] **Step 5: Commit** `feat(docker): add Engine API network and volume ops`

---

### Task 3: CLI 资源命令与解析

**Files:**
- Create: `internal/docker/cli_resources.go`
- Create: `internal/docker/cli_resources_test.go`

**Interfaces:**
- `BuildDockerImagesCommand()` → `docker images --format '{{json .}}'`
- `BuildDockerNetworksCommand()` → `docker network ls --format '{{json .}}'`
- `BuildDockerVolumesCommand()` → `docker volume ls --format '{{json .}}'`
- `BuildDockerRmiCommand(ref, force)` / `BuildDockerNetworkRmCommand` / `BuildDockerVolumeRmCommand`
- `BuildDockerImageInspectCommand` / Network / Volume
- `ParseDockerImagesJSONLines` / Networks / Volumes（过滤非 `{` 行）
- `CLIListImages/Networks/Volumes`、`CLIInspect*`、`CLIRemove*`（复用 `CLISession` + `cliRun`）

- [ ] **Step 1–4:** 解析样例 JSON + shellQuote 注入测

- [ ] **Step 5: Commit** `feat(docker): add CLI image network volume commands`

---

### Task 4: DockerService 资源方法

**Files:**
- Create: `internal/service/docker_resources.go`
- Create: `internal/service/docker_resources_test.go`
- 可选：小改 `docker_service.go` 抽出共用 `withSession` 若减少重复

**行为:**
- 每个 List/Inspect/Remove：`lookupSession` → API 或 CLI
- CLI 路径设 `Via: "cli"`
- 超时：list/inspect 用 `dockerOpTimeout`；remove 可用 `dockerLifecycleTimeout`

- [ ] **Step 1:** mock API httptest + fake CLI；List 与 Remove force

- [ ] **Step 2–4:** 实现 PASS

- [ ] **Step 5: Commit** `feat(docker): wire image network volume ops in DockerService`

---

### Task 5: app.go + Wails 绑定

**Files:**
- Modify: `app.go`
- Modify: `frontend/wailsjs/go/main/App.js`、`App.d.ts`

- [ ] **Step 1:** 导出上表 9 个方法；nil service 检查对齐现有 Docker API

- [ ] **Step 2:** `go build -o /dev/null .`

- [ ] **Step 3: Commit** `feat(docker): expose image network volume APIs to frontend`

---

### Task 6: 前端纯函数

**Files:**
- Create: `frontend/src/lib/dockerResources.js`
- Create: `frontend/test/dockerResources.test.js`

**Exports（示例）:**
```javascript
export const DOCKER_RESOURCE_TABS = [
  { id: 'containers', label: '容器' },
  { id: 'images', label: '镜像' },
  { id: 'networks', label: '网络' },
  { id: 'volumes', label: '卷' },
];

export function formatDockerSize(bytes) { /* B/KB/MB/GB */ }
export function imageDisplayTags(image) { /* Tags 数组或 RepoTags → 可读字符串 */ }
export function resourceRowKey(kind, row) { /* images: ID|first tag; networks: ID|Name; volumes: Name */ }
export function removeResourceConfirmCopy(kind, label, force) { /* 中文确认框 */ }
export function resourceDetailFields(kind, detail) { /* [{label, value}] 供侧栏 */ }
```

- [ ] **Step 1–4:** TDD

- [ ] **Step 5: Commit** `feat(docker): add resource tab and delete copy helpers`

---

### Task 7: DockerWorkspace UI

**Files:**
- Create: `frontend/src/components/DockerResourceDetail.svelte`
- Modify: `frontend/src/components/DockerWorkspace.svelte`
- Create: `docs/changes/features/2026-10-27-docker-resources-impl.md`

**行为:**
1. 工具栏下（或标题旁）资源 Tab：`containers|images|networks|volumes`，默认 `containers`
2. 切 Tab：清空选中详情；按 kind 调对应 List；loading/error 复用现有 banner
3. 容器 Tab：保持现有表格与菜单/Exec/详情（`DockerContainerDetail`）
4. 镜像/网络/卷：各自表格列按规格；行点击 → Inspect（失败则用列表行字段）→ 右侧可拖壳内挂 `DockerResourceDetail`（仅概览 KV + 关闭）
5. 行末或右键「删除」→ ConfirmDialog；镜像带强制勾选；成功后刷新并关侧栏
6. 「仅运行中」过滤仅在容器 Tab 显示
7. 刷新按钮：刷新当前 Tab 列表

`DockerResourceDetail` props：`kind`、`title`、`detail`、`detailError`、`onClose`；**无**日志/Exec。

- [ ] **Step 1:** 实现 UI

- [ ] **Step 2:** `cd frontend && node --test test/dockerResources.test.js test/dockerWorkspace.test.js test/dockerContainerActions.test.js`

- [ ] **Step 3:** 写中文变更文档（背景/范围/修改文件/验证/剩余风险）

- [ ] **Step 4: Commit** `feat(docker): browse and delete images networks volumes in workspace`

---

### Task 8: release 说明

**Files:**
- Modify: `.github/workflows/release.yml`

在 Docker 三期条目附近增加：  
`Docker 三期 — 镜像/网络/卷：列表、详情与删除（API+CLI）`

- [ ] **Step 1:** 更新

- [ ] **Step 2: Commit** `docs(docker): note image network volume browse in release notes`

---

## 手工验收

1. 本机 Docker：切镜像/网络/卷 Tab，列表非空时可刷新
2. 点行打开侧栏，字段可读；拖拽宽度仍可用
3. 删除未使用卷成功；删除被占用网络见错误文案
4. 强制删除无 tag 镜像（注意别删正在用的）
5. SSH CLI 回退环境抽测 list + remove

## Spec 覆盖自检

| 规格 | 任务 |
|------|------|
| 资源 Tab 默认容器 | Task 7 |
| 列表字段 | Task 1–3 + 7 |
| 详情侧栏无可日志 | Task 7 |
| 删除确认 + 镜像 force | Task 6–7 |
| Wails API 表 | Task 4–5 |
| API + CLI | Task 1–4 |
| 不做 pull/create/run | 全局约束 |
