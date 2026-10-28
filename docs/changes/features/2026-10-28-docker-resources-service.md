# Docker 镜像/网络/卷：DockerService 方法接线

## 背景

Phase 3 Task 4：`internal/docker` 已提供镜像、网络、卷的 API 与 CLI 实现，需要在 `DockerService` 暴露给上层（后续由 `app.go` 绑定到前端）。

## 范围

- 新增 `ListDockerImages/Networks/Volumes`、`InspectDockerImage/Network/Volume`、`RemoveDockerImage(sessionID, ref, force)`、`RemoveDockerNetwork`、`RemoveDockerVolume`。
- 流程：`lookupSession` → API（带超时）或 CLI（`cliSession` 超时包装）；CLI 结果 `Via: "cli"`。
- list/inspect 使用 `dockerOpTimeout`，remove 使用 `dockerLifecycleTimeout`。
- 入参校验：镜像引用允许 `:` `/` `@`；网络/卷名称非空、不以 `-` 开头、仅 `[A-Za-z0-9_.-]`。
- `RemoveDockerVolume` 始终 `force=false`。
- 不涉及 `app.go` 绑定与前端。

## 修改文件

- `internal/service/docker_resources.go`（新增）
- `internal/service/docker_resources_test.go`（新增）

## 验证

- `go test ./internal/service -run TestDocker` 通过（API mock httptest + fake CLI；覆盖 list/inspect/remove、force 参数、校验拒绝、未知会话）。

## 剩余风险

- 校验较严格：含空格或特殊字符的卷/网络名称（Docker 本身通常也不允许）会被拒绝。
- `app.go` 尚未导出这些方法，需后续任务接线。
