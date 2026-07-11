# JDBC 运行时恢复界面

## 背景

JDBC 运行时切换完成后，驱动管理界面只在用户主动操作时刷新状态，无法及时反映 Agent 异常退出；错误界面的“查看日志”也只能展示日志路径，诊断仍需离开应用完成。

## 范围

- 驱动管理界面每 2 秒轮询运行时与 Agent 状态，页面不可见时暂停请求。
- 组件销毁时清理轮询定时器，避免后台残留请求。
- 运行时切换、安装和导入成功后立即使用后端激活结果更新界面。
- 状态轮询不刷新驱动列表，也不覆盖用户操作产生的错误信息。
- 增加 Agent 日志对话框，支持刷新、复制和关闭，并显示日志大小、截断状态、空日志及读取错误。
- 重新生成 Wails 绑定，并更新仓库内跟踪的前端构建产物。

## 修改文件

- `frontend/src/components/JDBCDriverManager.svelte`
- `frontend/wailsjs/go/main/App.d.ts`
- `frontend/wailsjs/go/main/App.js`
- `frontend/wailsjs/go/models.ts`
- `frontend/build/assets/index.js`
- `frontend/build/assets/index.css`
- `frontend_contract_test.go`

## 验证

- 先运行 `go test . -run TestJDBCDriverManagerIncludesPollingAndLogViewer -v`，确认新增契约测试因缺少轮询与日志查看功能失败。
- 实现后运行 `go test . -run TestJDBCDriverManagerIncludesPollingAndLogViewer -v`，测试通过。
- 运行 `go test ./...`，全部 Go 测试通过；沙箱内首次运行时因 `httptest` 无权监听本机临时端口失败，最小修复是在沙箱外重跑相同命令。
- 运行 `npm run build`，Vite 构建与 Gradle `shadowJar` 均成功，JDBC Agent 已暂存到前端构建目录。
- 沙箱内首次执行 Gradle 时因无法访问 `~/.gradle` 锁文件失败；最小修复是在沙箱外重跑相同命令，没有修改或绕开 Gradle 工具链。
- 构建仍报告仓库既有的 Svelte 可访问性、动态导入和大分块警告，本次没有扩大处理范围。

## 剩余风险

- 状态轮询采用固定 2 秒间隔，极短暂的 Agent 状态变化可能不会展示。
- 日志复制依赖系统剪贴板权限，失败时会在对话框中显示错误。
- 完整桌面端交互与应用重启后的恢复行为留待最终验收任务验证。
