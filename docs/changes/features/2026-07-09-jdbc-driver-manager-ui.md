# JDBC 驱动管理前端页面

## 背景

后端已经暴露 JDBC 驱动管理 API，前端需要提供可操作的管理入口，让用户查看驱动安装状态、导入离线驱动包、校验或卸载已安装驱动，并确认 JRE 与 agent 状态。

## 范围

- 新增 `JDBCDriverManager.svelte` 驱动管理页面。
- 在全局设置中新增“数据库驱动”页签。
- 驱动页面加载 `ListJDBCDrivers()` 和 `GetJDBCRuntimeStatus()`。
- 左侧支持搜索和按安装状态过滤。
- 右侧展示驱动详情、profile、jar、URL template 和高级属性。
- 操作按钮调用真实 Wails API，包括安装、离线导入、校验、重新安装、卸载和重启 agent。

## 修改文件

- `frontend/src/components/JDBCDriverManager.svelte`
- `frontend/src/components/GlobalSettingsDialog.svelte`
- `frontend/build/assets/index.css`
- `frontend/build/assets/index.js`
- `go.mod`
- `go.sum`

## 验证

- [ ] 打开全局设置，能进入“数据库驱动”页。
- [ ] 左栏能搜索、过滤驱动。
- [ ] 右栏能展示驱动详情、版本、jar、URL template、高级配置。
- [ ] 未安装驱动显示“安装”和“导入离线包”。
- [ ] 已安装驱动显示“校验”“重新安装”“卸载”。
- [ ] 顶部显示 JRE 状态和 agent 状态。
- 已运行 `cd frontend && npm run build`，结果通过。
- 已运行 `/Users/dingwei/go/bin/wails dev`，bindings 生成、前端编译、应用编译和打包均通过，并启动到开发 URL。
- `wails dev` 运行过程中触发 `go mod tidy`，将 JDBC proto 生成代码实际使用的 `google.golang.org/grpc` 和 `google.golang.org/protobuf` 整理为 direct dependency，并移除旧版 gRPC 的无用校验项。
- 尝试使用 Playwright 做点击级验收时，本机默认 Chromium 未安装；改用 Chrome channel 时浏览器进程在 sandbox 内退出并出现 `kill EPERM`，shell 环境也没有可直接加载的 `playwright` 模块，因此本轮未完成点击级截图验收。

## 剩余风险

- 离线包路径首版使用路径输入，后续可改为 Wails 文件选择器。
- 在线安装按钮已接 `InstallJDBCDriver`，但后端当前仍返回在线安装未实现。
- agent 状态首版通过“重启 agent”操作反馈展示，尚未接实时健康检查事件。
- 点击级手工验收仍需在可启动 Chrome/Playwright 或真实 Wails 窗口的环境中补做。
