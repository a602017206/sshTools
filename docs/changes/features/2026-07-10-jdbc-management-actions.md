# JDBC 管理操作与状态展示

## 背景

JDBC 驱动管理界面此前仍通过浏览器 `prompt` 输入本地路径，“安装 JRE”只切换运行时模式，没有执行真实下载。Agent 状态也使用静态文字，无法反映启动、运行或失败状态。

## 范围

- 增加 JDBC Agent 的 `stopped`、`starting`、`running`、`failed` 状态，以及运行时类型和最后错误信息。
- 增加托管 JRE 在线安装、JRE 归档导入、运行时归档选择、离线驱动包选择和 Java 可执行文件选择 API。
- 文件选择通过可注入适配器调用 Wails 对话框，后端测试无需启动图形界面。
- JDBC 驱动管理页和数据库错误操作区移除路径 `prompt`，统一使用系统文件选择器。
- “安装 JRE”调用真实托管运行时安装流程，Agent 状态栏展示 supervisor 的实际状态和最后错误。
- 切换系统 Java 或托管 JRE 时复用现有 `RuntimeService`，避免 supervisor 保留失效的旧运行时引用。
- 使用 Wails CLI 根据 Go 导出方法重新生成前端 bindings。

## 修改文件

- `app.go`
- `app_jdbc_test.go`
- `internal/service/jdbc_agent_supervisor.go`
- `internal/service/jdbc_api_models.go`
- `internal/service/jdbc_runtime.go`
- `frontend/src/components/JDBCDriverManager.svelte`
- `frontend/src/components/DatabasePanel.svelte`
- `frontend/build/assets/index.js`
- `frontend/wailsjs/go/main/App.js`
- `frontend/wailsjs/go/main/App.d.ts`
- `frontend/wailsjs/go/models.ts`

## 验证

- `go test . -run TestJDBCManagementAPIReturnsAgentAndRuntimeState -v`
- `go test ./...`
- `/Users/dingwei/go/bin/wails generate module`
- `cd frontend && npm run build`
- 检查两个 JDBC 前端组件，确认不再包含 `window.prompt`。

## 工具链阻塞记录

首次执行 `/Users/dingwei/go/bin/wails generate module` 时，沙箱禁止读取 `~/Library/Caches/go-build`，生成过程失败。最小修复是在允许读取本机 Go 构建缓存的环境中重跑同一命令；未手工伪造 bindings，也未修改生成流程。

## 剩余风险

- 文件选择器依赖 Wails 桌面运行时，浏览器单独预览时无法完成原生文件选择。
- Agent 状态为当前进程内快照，应用重启后从 `stopped` 重新开始，不持久化历史错误。
- 前端构建仍报告仓库既有的 Svelte 可访问性和大包告警，本次未扩大范围处理。
