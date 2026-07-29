# Oracle ZHS16GBK 字符集加载修复

## 背景

使用 Oracle `ZHS16GBK` 数据库连接时，JDBC agent 报 ORA-17056，提示需要在类路径中添加 `orai18n.jar`。原因是 Oracle 推荐 profile 只安装了核心 `ojdbc11.jar`。

## 范围

将官方 Maven Central 的同版本 `orai18n.jar` 加入 Oracle profile，并提升清单版本。驱动状态检查改为验证 profile 的全部 jar；既有缺失 NLS jar 的安装会显示为未安装，用户可直接重新安装。

## 修改文件

- `internal/service/jdbc_builtin_manifest.json`
- `internal/service/jdbc_catalog.go`
- `internal/service/jdbc_catalog_test.go`
- `docs/designs/2026-07-29-oracle-jdbc-nls-dependency.md`

## 验证

目录测试覆盖 Oracle profile 的两个官方 jar、下载 URL 与 SHA-256；另一个测试覆盖仅有 `ojdbc11.jar` 的旧安装不再被标为已安装。

## 剩余风险

该修复覆盖 Oracle 官方的 NLS 字符集支持 jar。若服务端使用自定义字符集映射，仍可能需要由数据库管理员提供额外的 Oracle 映射包。
