# JDBC JRE 运行时选择模型

## 背景

JDBC agent 需要由本地 Java 运行时启动。为了支持稳定分发和离线环境，应用需要优先使用受管理的本地 JRE，同时保留高级用户选择系统 Java 的能力。

## 范围

- 新增 `RuntimeService`，负责选择 JDBC agent 使用的 Java 运行时。
- 定义 `managed`、`system`、`missing` 三种运行时状态。
- 默认优先选择 `runtimes/` 目录下最新的托管 JRE。
- 支持显式切换到系统 Java 路径。
- 预留 `ImportRuntimeArchive` 接口，后续任务再实现真实导入。

## 修改文件

- `internal/service/jdbc_runtime.go`
- `internal/service/jdbc_runtime_test.go`

## 验证

- 已运行 `go test ./internal/service -run TestRuntimeService -v`，结果通过。

## 剩余风险

- 当前只通过目录名排序选择最新托管 JRE，尚未解析真实 Java 版本。
- `ImportRuntimeArchive` 仍为占位接口，不能导入真实 JRE 包。
- 系统 Java 只检查路径存在，尚未执行 `java -version` 验证兼容性。
