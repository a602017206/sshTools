# JDBC 驱动管理补全验收

## 背景

初始 JDBC 驱动管理实现保留了在线驱动安装、托管 JRE、应用 gateway 接线、文件选择和崩溃恢复缺口。本次补全工作按独立计划逐项实现，并在合并前执行完整自动化验收。

## 范围

- 将 Java agent shadow jar 纳入前端构建和 Wails 嵌入资源。
- 增加 supervisor、真实 gRPC client、运行时状态和应用退出清理。
- 增加托管 JRE 在线安装、安全归档导入和系统文件选择器。
- 增加首批八类数据库内置清单、公开驱动在线安装和受限驱动离线提示。
- 增加真实 Agent 崩溃后的 session 自动恢复集成测试。
- macOS 构建前校验待嵌入的 JDBC agent jar。

## 修改文件

- `internal/service/jdbc_integration_test.go`
- `scripts/test-jdbc-agent.sh`
- `scripts/build-mac.sh`
- `frontend/package.json.md5`
- `docs/development/2026-07-09-jdbc-driver-management-implementation.md`
- `docs/changes/features/2026-07-09-jdbc-driver-management-rollout.md`

## 验证

- `go test ./...`：通过。
- `cd jdbc-agent && ./gradlew test`：通过。
- `./scripts/test-jdbc-agent.sh`：通过，执行基础 H2 闭环和真实崩溃恢复两项测试。
- `cd frontend && npm run build`：通过，生成前端资源并暂存 agent jar。
- `/Users/dingwei/go/bin/wails build`：通过。
- `test -f build/bin/AHaSSHTools.app/Contents/MacOS/AHaSSHTools`：通过。

## 工具链阻塞记录

- 沙箱内 Gradle 无法创建 `~/.gradle` wrapper 锁文件。最小修复是在允许访问用户 Gradle 缓存的环境重跑原命令，未更改 Gradle 配置。
- 首次前端完整构建在 agent 暂存阶段受到同一 Gradle 锁文件限制。最小修复是重跑完整 `npm run build`，未跳过后置脚本。
- Wails bindings 生成曾因无法读取 Go 构建缓存失败。最小修复是在允许访问缓存的环境重跑原命令，bindings 均由 Wails 生成。

## 剩余风险

- 本轮完成自动化 API、集成和生产构建验收，但未操作真实桌面窗口完成逐按钮点击回归。
- 在线安装依赖上游服务可用性；受限厂商驱动仍要求用户取得授权包后离线导入。
- 前端保留仓库既有的 Svelte 可访问性、大分块和混合导入告警。
