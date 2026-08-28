# Sonar 高优先级问题修复

## 背景

SonarQube 扫描报告了重复字面量、空实现说明、DOM 属性写法和过高认知复杂度等高优先级问题。

## 范围

修复 CSV 指定的 Go、JavaScript 与 Java 源文件，采用保持行为不变的重构方式。

## 修改文件

- `internal/api/handlers/connection.go`、`internal/api/handlers/session.go`
- `internal/service/copilot/tools.go`、`internal/service/jdbc_gateway.go`
- `internal/service/jdbc_agent_process_windows.go`、`menu.go`、`app.go`
- `internal/service/devtools_service.go`、`internal/service/devtools_service_test.go`、`internal/service/artifact_download.go`、`internal/service/archive_extract.go`
- `internal/service/jdbc_runtime_provider.go`
- `internal/service/jdbc_catalog_test.go`
- `internal/service/jdbc_install.go`
- `internal/service/copilot/tools.go`、`internal/service/copilot/service.go`
- `internal/service/native_redis_operations.go`
- `internal/service/database_service.go`
- `internal/ssh/manager.go`、`internal/ssh/manager_test.go`
- `app.go`
- `frontend/src/lib/fileManagerContextMenu.js`、`frontend/src/lib/tableAlterSQL.js`
- `frontend/src/settings/appearance.js`
- `jdbc-agent/src/main/java/com/sshtools/jdbcagent/MetadataServiceImpl.java`
- 三个 JDBC Agent 测试观察者类
- 本次计划、设计与实施记录

## 验证

- Go 服务、Copilot、SSH 与 handler 目标包测试通过；本轮 Redis 定向测试通过。
- 文件管理器快捷键与表结构变更 SQL 的前端纯函数 Node 校验通过。
- JDBC Agent 三个受影响测试类通过。
- 根包测试存在两项既有前端契约失败，未因本次修改处理。
- Sonar 服务当前无法从终端访问，尚未完成重扫。

## 剩余风险

剩余复杂度重构涉及 SSH 会话、SFTP、数据库、配置更新和连接导入；在补充行为测试前不进行批量重构，以避免引入业务回归。
