# JDBC 运行时恢复验收记录

## 背景

JDBC 运行时持久化、启动恢复、原子切换、Agent 状态轮询和日志查看已经完成实现，需要通过完整自动验证与真实桌面操作确认端到端行为。

## 范围

- 验证旧配置兼容和运行时选择持久化。
- 验证系统 Java 与托管 JRE 在切换和应用重启后的行为。
- 验证 Agent 状态自动刷新、日志对话框和应用退出清理。
- 执行 Go、Gradle、真实 Agent、前端及 Wails 生产构建回归。

## 验收表

| 验收项 | 状态 | 证据 |
| --- | --- | --- |
| 旧配置迁移 | 通过 | `TestConfigManagerLoadsLegacySettingsWithoutJDBCFields` 与 Go 全量测试通过。 |
| 系统 Java 选择保存与应用重启恢复 | 通过 | 桌面选择 JDK 21 后立即显示完整路径和 Agent `running`；重开应用后恢复相同路径。 |
| 托管 JRE 选择保存与应用重启恢复 | 通过 | 桌面安装并切换 `jre-21.0.11-10-jre-darwin-arm64`，重开应用后仍显示托管 JRE。 |
| 切换后 Agent 立即进入 `running` 或 `failed` | 通过 | 系统 Java 与托管 JRE 切换后均在操作完成时显示 `running`。 |
| 页面在 2 秒内自动刷新状态 | 通过 | 契约测试固定 2000 毫秒轮询；真实终止 Agent 后页面自动显示“启动失败 / JDBC agent 进程已退出”。 |
| 日志对话框读取、刷新、复制和截断提示 | 通过 | 桌面显示生命周期日志，刷新成功，复制后显示“已复制”；追加 70000 个测试字节后显示“仅显示最近 64 KiB”。 |
| 应用退出后 Agent 子进程清理 | 通过 | Agent 运行时执行 `Command+Q`，`pgrep -f '[j]dbc-agent.jar'` 无结果。 |

## 修改文件

- `docs/development/2026-07-09-jdbc-driver-management-implementation.md`
- `docs/changes/features/2026-07-11-jdbc-runtime-recovery-completion.md`

## 验证

- `go test ./...`：通过。
- `cd jdbc-agent && ./gradlew test`：通过，保留 Gradle 9.0 兼容性弃用警告。
- `./scripts/test-jdbc-agent.sh`：通过。
- `cd frontend && npm run build`：通过，保留仓库既有的 Svelte 可访问性、动态导入和大分块警告。
- `/Users/dingwei/go/bin/wails build`：通过。
- `test -f build/bin/AHaSSHTools.app/Contents/MacOS/AHaSSHTools`：通过。
- 沙箱内 Gradle 锁文件、Go 构建缓存、`httptest` 监听端口和进程信号曾受权限限制；最小修复均为在沙箱外重跑原命令，没有绕开工具链。
- 桌面验收没有卸载或覆盖用户已有 MySQL 驱动；测试追加的日志文件已在验收后删除。

## 剩余风险

- Agent 日志采用追加写入，尚无轮转策略；查看接口只读取受限尾部，不影响磁盘长期增长。
- 前端仍有仓库既有的可访问性和大分块警告，未纳入本功能范围。
