# JDBC 运行时激活事务

## 背景

运行时模式可以恢复后，切换、在线安装和归档导入仍只修改内存或返回运行时状态，没有统一持久化并立即验证 Agent，失败语义也不一致。

## 范围

- 增加运行时与 Agent 的组合激活结果模型。
- 运行时切换按“应用模式、保存配置、重启 Agent”顺序执行。
- 配置保存失败时恢复运行时快照且不重启 Agent。
- Agent 重启失败时保留已保存的新选择，并返回 supervisor 的失败状态和最后错误。
- 托管 JRE 在线安装和归档导入成功后自动切换 managed 模式并重启 Agent。
- 应用启动时注入现有 `ConfigManager` 作为 JDBC 运行时设置存储。

## 修改文件

- `internal/service/jdbc_api_models.go`
- `app.go`
- `app_jdbc_test.go`

## 验证

- `go test . -run 'Test(SetJDBCRuntimeMode|ManagedRuntimeInstallAndImportActivateManagedMode)' -v`
- `go test ./...`

## 剩余风险

- Agent 重启使用固定 15 秒超时，慢速机器上可能需要后续提供可配置值。
- 托管 JRE 安装完成但配置保存失败时，已下载文件会保留，但当前选择会恢复；用户可稍后再次激活。
