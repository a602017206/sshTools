# JDBC 可操作错误处理

## 背景

JDBC gateway 原先只区分驱动缺失和数据库连接失败，Wails 前端收到的错误文本没有稳定错误码，用户无法根据运行时、驱动、agent 或数据库连接问题采取对应操作。驱动管理页和查询面板也只显示原始错误，没有恢复入口。

## 范围

- 新增统一的 `JDBCError`、五类稳定错误码和基于 agent 消息的分类函数。
- 为错误字符串增加 `[错误码]` 前缀，使 Wails 前端能够稳定识别错误类别。
- 统一声明 JDBC agent、驱动安装和运行时安装日志路径。
- 在数据库查询面板和驱动管理页中增加错误类别对应的行动按钮。
- 将数据库查询面板的“编辑连接”操作接到现有资产编辑对话框。
- 保留原始错误展开能力，并显示驱动文件或日志的本地路径。

## 修改文件

- `internal/service/jdbc_errors.go`
- `internal/service/jdbc_errors_test.go`
- `internal/service/jdbc_gateway.go`
- `frontend/src/App.svelte`
- `frontend/src/components/DatabasePanel.svelte`
- `frontend/src/components/JDBCDriverManager.svelte`
- `frontend/build/assets/index.css`
- `frontend/build/assets/index.js`
- `docs/changes/features/2026-07-09-jdbc-error-handling.md`

## 验证

- 红灯验证：`go test ./internal/service -run TestJDBCErrorMapsRuntimeMissingToActionableCode -v` 在实现前因 `MapJDBCAgentError` 未定义而失败。
- 绿灯验证：`go test ./internal/service -run TestJDBCError -v` 通过。
- 前端验证：`cd frontend && npm run build` 通过。
- 前端构建仍会报告仓库既有的 Svelte 可访问性警告和大分块警告，本次修改没有新增编译失败。

## 剩余风险

- 错误分类依赖稳定错误码和少量常见消息关键词；未识别的厂商错误统一归为 `DB_CONNECT_FAILED`。
- 在线驱动安装和托管 JRE 下载仍是前序任务保留的占位实现，对应按钮会展示后端返回的未实现错误；离线驱动导入和系统 Java 选择可直接执行。
- “查看日志”和“查看文件”当前在界面内显示本地路径，尚未调用操作系统文件管理器或日志查看器。
- 原始数据库错误可能包含厂商返回的连接信息；后续需要补充统一脱敏规则后再用于生产日志。
