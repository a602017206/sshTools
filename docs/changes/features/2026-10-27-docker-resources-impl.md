# Docker 三期：镜像/网络/卷浏览与删除（整体实现）

## 背景

第三阶段设计（见 `2026-10-27-docker-phase3-exec-resources-design.md`、计划 `docs/changes/process/2026-10-27-docker-resources-plan.md`）要求在 Docker 工作区中除容器外，还能浏览并删除镜像、网络、数据卷。本文件汇总该能力从后端到前端的全部改动；后端服务层的单独记录见 `2026-10-28-docker-resources-service.md`（已更新）。

## 范围

后端（`internal/docker`）：
- Engine API：镜像 list/inspect/remove、网络与卷 list/inspect/remove（`resource_ops.go`、`types.go` 新增 `ImageSummary`/`NetworkSummary`/`VolumeSummary` 等类型）。
- CLI 回退：`docker images/network ls/volume ls` 的 JSON 行解析、rmi/network rm/volume rm 与 inspect 命令构造，参数经 `shellQuote` 转义（`cli_resources.go`）。
- CLI 镜像去重修复：`docker images` 每个 tag 输出一行，`ParseDockerImagesJSONLines` 现按镜像 ID 合并为一条 `ImageSummary`，`Tags` 聚合（保持顺序、去重），使 API 与 CLI 路径的删除交互一致（按 ID/单 tag 删除语义相同）。

服务与绑定：
- `internal/service/docker_resources.go`：`DockerService` 的 List/Inspect/Remove 镜像、网络、卷方法，含入参校验与超时。
- `app.go`：已导出上述方法给前端（共 9 个绑定），`RemoveDockerVolume` 不暴露 force。
- `frontend/wailsjs/go/main/App.js`、`App.d.ts`、`frontend/wailsjs/go/models.ts`：重新生成的绑定与模型。

前端：
- `frontend/src/lib/dockerResources.js`：资源 Tab 定义、表格行格式化、删除确认文案（`removeResourceConfirmCopy`）等纯函数；`frontend/test/dockerResources.test.js` 为对应单测。
- `frontend/src/components/DockerWorkspace.svelte`：资源 Tab（容器/镜像/网络/卷，默认容器），按 Tab 加载列表，「仅运行中」仅在容器 Tab 显示，刷新只作用于当前 Tab；行末「删除」经 `ConfirmDialog` 确认，镜像提供「强制删除」勾选，成功后刷新并关闭侧栏。容器 Tab 表格、右键菜单、Exec、日志详情保持不变。
- `frontend/src/components/DockerResourceDetail.svelte`（新增）：只读概览键值对详情栏，仅标题与关闭；inspect 失败时回退为列表行字段并提示原因。

发布说明：
- `.github/workflows/release.yml`：Release Notes 增加「Docker 三期 — 镜像/网络/卷：列表、详情与删除（API+CLI）」。

相关变更文档：
- `docs/changes/features/2026-10-28-docker-resources-service.md`
- `docs/changes/features/2026-10-27-docker-phase3-exec-resources-design.md`（设计）
- `docs/changes/process/2026-10-27-docker-resources-plan.md`（计划）

## 修改文件

- `internal/docker/types.go`、`internal/docker/resource_ops.go`、`internal/docker/resource_ops_test.go`
- `internal/docker/cli_resources.go`、`internal/docker/cli_resources_test.go`（含 CLI 镜像按 ID 合并 tag 及单测 `TestParseDockerImagesMergesTagsByID`）
- `internal/service/docker_resources.go`、`internal/service/docker_resources_test.go`
- `app.go`
- `frontend/wailsjs/go/main/App.js`、`frontend/wailsjs/go/main/App.d.ts`、`frontend/wailsjs/go/models.ts`
- `frontend/src/lib/dockerResources.js`、`frontend/test/dockerResources.test.js`
- `frontend/src/components/DockerWorkspace.svelte`、`frontend/src/components/DockerResourceDetail.svelte`
- `.github/workflows/release.yml`
- `docs/changes/features/2026-10-27-docker-resources-impl.md`（本文件）、`docs/changes/features/2026-10-28-docker-resources-service.md`

## 验证

- `go test ./internal/docker -run 'Image|ParseDockerImages' -count=1`：通过（含同 ID 多 tag 合并用例）。
- `cd frontend && node --test test/dockerResources.test.js`：通过（6 项）。
- 前期记录：`node --test test/dockerResources.test.js test/dockerWorkspace.test.js test/dockerContainerActions.test.js`、终端剪贴板契约测试、`npx vite build` 均通过。
- 未做真机手工点击验证（需连接 Docker 引擎）。

## 剩余风险

- 网络/卷/镜像被引用时的删除失败信息依赖引擎返回，仅以横幅展示。
- 未提供右键菜单删除，仅行末按钮。
- 详情栏宽度仅会话内记忆。
- CLI 路径下同 ID 的 `<none>` 悬空镜像行也会按 ID 合并；`Size`/`Created` 取首行值。
