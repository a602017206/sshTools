# 人大金仓数据库列表加载失败

## 背景

人大金仓连接建立后，数据库面板会调用 JDBC gateway 的数据库列表接口。托管 gateway 仅将 `postgresql` 路由到 PostgreSQL 系统目录查询，遗漏了使用同一目录模型的 `kingbase`，因此连接后的数据库列表加载被拒绝。

## 范围

将 `kingbase` 加入 PostgreSQL 兼容的数据库列表查询分支，使用 `pg_database` 读取非模板数据库。不会改变连接、表列表、查询或其他数据库类型的行为。

## 修改文件

- `internal/service/jdbc_managed_gateway.go`
- `internal/service/jdbc_managed_gateway_test.go`
- 本变更记录。

## 验证

执行 `go test ./internal/service -run TestManagedJDBCGatewayListsKingbaseDatabasesThroughPostgreSQLCatalog -v`，验证人大金仓会通过 PostgreSQL 系统目录查询返回数据库列表；执行 JDBC gateway 相关测试验证既有 MySQL 路由未受影响。

## 剩余风险

不同人大金仓部署的系统目录权限可能限制普通用户读取全部数据库；此时应用会保留 JDBC 驱动返回的原始错误。未在当前环境连接真实人大金仓实例。
