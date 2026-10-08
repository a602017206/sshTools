# 缺陷：数据库空闲断开后无法自动恢复、无重连提示

## 背景

人大金仓连接空闲一段时间后被服务端关闭，再加载对象列表时显示：

```
加载失败：[DB_CONNECT_FAILED] rpc error: code = Unknown desc = This _connection has been closed.
```

后端 `managedGatewayCall` 本已支持"检测到连接失效 → 关闭并重开 JDBC 会话 → 重试一次"，但 `isJDBCSessionStale` 的关键词只覆盖了 Oracle 和少量通用文案，没有 PostgreSQL 系驱动（人大金仓、openGauss、PostgreSQL）的 `connection has been closed`，也没有 MySQL 的 `Communications link failure` 等，导致自动恢复没有触发，错误直接透出到界面且没有任何操作入口。

## 范围

- 后端：
  - 提取 `staleConnectionPatterns`，补充 PostgreSQL/人大金仓/openGauss、MySQL、达梦等驱动的断连报错，以及 agent 侧 `session not found`。
  - 新增错误码 `CONNECTION_LOST`，命中断连关键词的错误归入此类。
  - 自动重开会话失败时，返回 `CONNECTION_LOST`，消息为"数据库连接已断开，自动重连失败: ..."。
- 前端：
  - 新增 `isConnectionLostError` 统一识别断连错误。
  - 对象列表、表数据页显示"数据库连接已断开，可能是长时间空闲被服务器关闭"和"重新连接"按钮；原始错误放在悬停提示中。
  - 数据库面板按 `CONNECTION_LOST` 错误码提供"重新连接""查看原始错误"操作。
  - "重新连接"即重新执行当前加载，由后端自动重开会话。

## 修改文件

- `internal/service/jdbc_errors.go`
- `internal/service/jdbc_errors_test.go`
- `internal/service/jdbc_managed_gateway.go`
- `internal/service/jdbc_managed_gateway_test.go`
- `frontend/src/lib/databaseConnectionError.js`
- `frontend/src/components/SelectedDatabaseObjects.svelte`
- `frontend/src/components/DatabaseTablePanel.svelte`
- `frontend/src/components/DatabasePanel.svelte`
- `frontend/test/databaseConnectionError.test.js`
- `.github/workflows/release.yml`（发布说明）
- 本文档

## 验证

```bash
go test ./internal/service/
cd frontend && node --test test/*.test.js && npm run build
```

- Go 测试全部通过，新增人大金仓断连自动重开、重开失败返回 `CONNECTION_LOST` 两个用例。
- 前端 265 个测试中 263 个通过；失败的 2 个位于 `test/nativeDatabasePanelLayout.test.js`（Redis/Kafka 工作区），与本次改动无关，属既有失败。
- 未在桌面端对真实人大金仓空闲断连做手工验证。

## 剩余风险

- 断连识别依赖驱动报错文案；未收录的驱动文案仍会按普通错误展示，需要按实际报错继续补充关键词。
- Redis、Kafka 等原生（非 JDBC）连接不在本次范围内。
- 自动重连只重试一次；若重开时数据库仍不可达，需用户稍后点击"重新连接"。
