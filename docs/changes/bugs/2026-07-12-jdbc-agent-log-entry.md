# JDBC Agent 自动失败状态缺少日志入口

## 背景

真实桌面验收中，Agent 子进程退出后页面可通过轮询显示“启动失败”，但日志入口只存在于操作错误横幅的 `AGENT_UNAVAILABLE` 分支。自动状态刷新不会生成该横幅，因此用户无法从当前失败状态打开日志对话框。

## 范围

- 在 Agent 状态操作区常驻显示“查看日志”按钮。
- 复用现有日志读取、刷新、复制和关闭逻辑，不修改日志文件访问边界。
- 增加前端契约测试，确认正常状态和自动失败状态均有日志入口。

## 修改文件

- `frontend/src/components/JDBCDriverManager.svelte`
- `frontend_contract_test.go`

## 验证

- 修复前运行 `go test . -run TestJDBCDriverManagerAlwaysExposesAgentLog -v`，测试因状态操作区缺少日志入口而失败。
- 修复后运行 `go test . -run 'TestJDBCDriverManager' -v`，三个前端契约测试均通过。
- 运行 `cd frontend && npm run build`，Vite 与 Gradle `shadowJar` 均成功；保留仓库既有的可访问性、动态导入和大分块警告。
- 运行 `/Users/dingwei/go/bin/wails build`，生产应用打包成功。
- 真实桌面复验中，Agent 已停止和自动失败状态均可从状态操作区打开日志对话框。

## 剩余风险

- 日志读取仍受固定路径和最大 1 MiB 限制，这是既定安全边界。
