# 首批 JDBC 数据库类型

## 背景

JDBC agent 架构需要覆盖首批常见数据库类型，让连接表单可以选择对应类型、自动填充默认端口，并在驱动未安装时提示用户先到驱动管理页处理依赖。

## 范围

- `GetDefaultPort` 覆盖 MySQL、PostgreSQL、SQLite、Oracle、SQL Server、达梦、人大金仓和 openGauss。
- 数据库连接表单新增首批 JDBC 类型。
- SQLite 类型隐藏主机和端口字段，改为填写数据库文件路径或 JDBC URL。
- SQLite 类型隐藏用户名和密码字段。
- 表单在能读取驱动状态时，对未安装的 JDBC 驱动提示先安装。
- 前端构建产物随 `npm run build` 更新。

## 修改文件

- `internal/config/database.go`
- `internal/service/jdbc_catalog_test.go`
- `frontend/src/components/AddAssetDialog.svelte`
- `frontend/build/assets/index.js`

## 验证

- 已运行 `go test ./internal/service -run TestDriverCatalogReturnsDefaultPortsForInitialJDBCTypes -v`，结果通过。
- 已运行 `cd frontend && npm run build`，结果通过。
- 前端构建仍输出既有 a11y 警告和 chunk size 警告，本次未处理这些历史问题。

## 剩余风险

- 驱动安装状态依赖 `ListJDBCDrivers`，如果 manifest 缺失或 API 不可用，表单不会强制阻止连接。
- SQLite 文件选择器尚未接入，首版使用手动输入路径或 JDBC URL。
- 不同国产数据库的连接属性和 URL 模板差异后续仍需在驱动 profile 中细化。
