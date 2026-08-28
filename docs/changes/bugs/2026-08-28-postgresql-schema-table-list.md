# 修复：PostgreSQL 有表但对象列表为空

## 背景

PostgreSQL 连接成功且数据库中存在表时，主对象列表仍可能显示为空。前端把当前数据库名称传入 JDBC 元数据接口，后端又将它作为 `catalog` 传给 PostgreSQL 驱动；PostgreSQL 的对象元数据应在当前连接内按 Schema 枚举，数据库名不是该接口可用的 `catalog` 过滤条件。

## 范围

修复 PostgreSQL 及其兼容数据库（人大金仓、openGauss）的 JDBC 元数据请求范围，使 Schema 下的表、视图等对象不再因错误的 `catalog` 被过滤为空。对于左侧选择的其他 PostgreSQL 数据库，使用同一凭据建立短暂连接读取对象元数据，避免继续误用默认连接库。

## 修改文件

- `internal/service/jdbc_managed_gateway.go`：PostgreSQL 兼容数据库请求 Schema 元数据时传入空 `catalog`；跨数据库浏览时建立并关闭短暂元数据连接。
- `internal/service/jdbc_managed_gateway_test.go`：新增业务 Schema 表对象请求的回归测试。

## 验证

- `go test ./internal/service -run TestManagedJDBCGatewayListsPostgreSQLSchemaObjectsWithoutCatalog -count=1 -v`
- `go test ./internal/service -count=1`

## 剩余风险

跨库对象浏览要求当前保存的账号对目标数据库具备 `CONNECT` 权限，并对目标 Schema 具备对象可见权限。表结构详情和 DDL 查询仍使用当前主连接，跨库对象的详情将在后续单独扩展。
