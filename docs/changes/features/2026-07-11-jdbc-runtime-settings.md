# JDBC 运行时设置持久化

## 背景

JDBC 运行时模式和系统 Java 路径此前只保存在进程内，应用重启后无法恢复用户选择。后续自动恢复需要先把选择纳入现有应用配置模型。

## 范围

- 在 `AppSettings` 增加 JDBC 运行时模式和系统 Java 路径字段。
- 增加专用配置更新方法，只接受空模式、`managed` 和 `system`。
- `system` 模式拒绝空 Java 路径。
- 配置保存失败时恢复内存旧值，避免状态分叉。
- 旧配置缺少新增字段时保持兼容，字段使用空值。

## 修改文件

- `internal/config/config.go`
- `internal/config/config_test.go`

## 验证

- `go test ./internal/config -run 'TestConfigManager(PersistsJDBCRuntimeSettings|LoadsLegacySettingsWithoutJDBCFields|RejectsInvalidJDBCRuntimeMode)' -v`
- `go test ./internal/config -v`

## 剩余风险

- `ConfigManager.Save` 仍沿用直接写文件方式，尚未统一改为临时文件原子替换。
- 本任务只持久化设置，运行时服务恢复和 Agent 激活由后续任务实现。
