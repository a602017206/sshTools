# JDBC 运行时切换入口缺失

## 背景

真实桌面验收时发现，当系统已有可用 Java 运行时，数据库驱动设置页只显示运行时状态、刷新和 Agent 重启操作。托管 JRE、运行时归档导入和系统 Java 选择按钮仅在 `RUNTIME_MISSING` 错误分支出现，导致正常状态无法切换运行时。

## 范围

- 在运行时状态操作区常驻显示托管 JRE、导入 JRE和系统 Java 三个切换入口。
- 保留 `RUNTIME_MISSING` 错误分支中的恢复操作，不修改后端激活事务。
- 增加前端契约测试，确认正常状态也暴露三个切换处理函数。

## 修改文件

- `frontend/src/components/JDBCDriverManager.svelte`
- `frontend_contract_test.go`

## 验证

- 修复前运行 `go test . -run TestJDBCDriverManagerAlwaysExposesRuntimeSwitching -v`，测试因三项入口均缺失而失败。
- 修复后运行 `go test . -run 'TestJDBCDriverManager(AlwaysExposesRuntimeSwitching|IncludesPollingAndLogViewer)' -v`，两个契约测试均通过。
- 运行 `cd frontend && npm run build`，Vite 与 Gradle `shadowJar` 均成功。
- 运行 `/Users/dingwei/go/bin/wails build`，生产应用打包成功。

## 剩余风险

- 运行时切换会启动文件选择器或下载流程，仍需在真实桌面应用中继续验证。
