# JDBC agent 构建资源部署

## 背景

Java agent 虽然可以生成 shadow jar，但 Wails 生产构建没有把该 jar 放入应用资源，运行时固定查找 `~/.sshtools/agent/jdbc-agent.jar` 也没有对应部署步骤，导致生产包无法启动 JDBC agent。

## 范围

- 新增 agent jar 原子部署服务。
- 相同内容重复部署时不重写文件。
- Vite 构建完成后运行 Gradle `shadowJar`，并把产物暂存到 Wails 会嵌入的 `frontend/build`。
- 忽略生成的 jar，避免把二进制构建产物提交到仓库。

## 修改文件

- `.gitignore`
- `frontend/package.json`
- `internal/service/jdbc_agent_artifact.go`
- `internal/service/jdbc_agent_artifact_test.go`
- `scripts/stage-jdbc-agent.sh`
- `docs/changes/features/2026-07-10-jdbc-agent-artifact.md`

## 验证

- 红灯：`go test ./internal/service -run TestAgentArtifactInstallerWritesJarAtomically -v` 在实现前因 `NewAgentArtifactInstaller` 未定义而失败。
- 绿灯：`go test ./internal/service -run TestAgentArtifactInstallerWritesJarAtomically -v` 通过。
- 构建：`cd frontend && npm run build` 通过，并生成 `frontend/build/jdbc-agent.jar`。
- 工具链阻塞：沙箱内构建完成 Vite 阶段后，Gradle wrapper 因不能创建 `~/.gradle` 锁文件而失败；未改构建流程，仅在沙箱外重跑同一命令后通过。

## 剩余风险

- 当前原子替换依赖操作系统 `rename` 行为；Windows 对已存在目标文件的替换语义需要在 Windows 构建机补充验证。
- Gradle 仍会报告面向 Gradle 9 的既有弃用警告。
