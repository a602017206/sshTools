# Docker 二期：容器操作、详情与日志（实现）

## 背景

第一期 Docker 连接仅支持列出容器。二期在既有会话模型（本机 / SSH 隧道 API / CLI 回退 / 直连 API）上补齐生命周期操作、Inspect 详情与日志尾部及跟流，且 API 与 CLI 双路径行为一致；不包含 exec/attach 等后续能力。

Task 7 侧栏 UI 的单独变更说明见同目录 `2026-10-23-docker-container-detail-sidebar.md`；本文档为二期整体实现记录。

## 范围

- **后端**：Engine API 与 CLI 的 Start/Stop/Restart/Remove、Inspect（`ContainerDetail`）、日志尾部与 follow 流；multiplex 日志解码；`DockerService` 统一分支与会话级 follow 状态；`CloseDocker` / `CloseAll` / 应用关闭时停止跟流；Wails 导出方法与 `docker:logs:{sessionID}:{containerID}` 事件。
- **前端**：容器行操作菜单（按状态禁用）、删除确认、右侧详情侧栏（概览 + 日志 Tab）、尾部行数、跟流/清空/复制与滚底行为；绑定与 `dockerContainerActions` / `dockerLogs` 辅助模块。
- **发布**：`.github/workflows/release.yml` 更新说明补充二期条目。
- **不在范围**：exec、镜像/网络/卷管理、Compose、K8s、TLS 证书 UI、Windows named pipe、移动端。

## 修改文件

| 区域 | 文件 |
|------|------|
| Engine API | `internal/docker/types.go`、`container_ops.go`、`container_logs.go`、`*_test.go` |
| CLI | `internal/docker/cli_ops.go`、`cli_ops_test.go` |
| 服务层 | `internal/service/docker_service.go`、`docker_logs_follow.go`、`docker_service_test.go` |
| SSH | `internal/ssh/client.go`（CLI follow 流式读） |
| Wails | `app.go`；`frontend/wailsjs/go/main/App.js`、`App.d.ts`、`models.ts` |
| 前端 | `DockerWorkspace.svelte`、`DockerContainerDetail.svelte`、`ConfirmDialog.svelte`、`dockerContainerActions.js`、`dockerLogs.js` |
| 测试 | `frontend/test/dockerContainerActions.test.js`、`dockerLogs.test.js`、`dockerWorkspace.test.js` |
| 文档 | 本文件；设计/计划见 `docs/superpowers/`、`docs/designs/`、`docs/changes/` 下 2026-10-23-docker-phase2-* |

## 验证

- `go test ./internal/docker ./internal/service -count=1`
- `cd frontend && node --test test/dockerContainerActions.test.js test/dockerLogs.test.js test/dockerWorkspace.test.js src/lib/workspaceTabs.test.js`
- 未改终端剪贴板受保护区；若后续动到终端需跑契约三件套。

## 剩余风险

- CLI follow 依赖远端 `docker logs -f` 与 SSH 会话稳定性，弱网或 daemon 重启时可能断流，需用户手动重开跟流。
- 侧栏宽度暂不可拖拽；大规模日志仅前端节流拼接，极长会话仍可能占内存。
- 未在全部 transport 组合（隧道失败纯 CLI、直连 API 等）上做完整手工矩阵验证。
