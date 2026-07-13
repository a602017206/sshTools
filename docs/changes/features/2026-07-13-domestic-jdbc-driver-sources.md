# 国产 JDBC 驱动自动安装来源

## 背景

达梦和人大金仓原有 profile 仅支持离线导入，未提供官方在线下载地址和 SHA-256，导致自动安装不可用。

## 范围

- 达梦 DM8 使用 `com.dameng:DmJdbcDriver8:8.1.5.45`。
- 人大金仓提供 V8 `8.6.1` 与 V9 `9.0.1` 两个独立 profile，默认推荐 V8。
- 三个 jar 都从 Maven Central HTTPS 地址下载，并使用固定 SHA-256 校验。
- 保留离线导入，供受限网络或服务端版本不匹配时使用。

## 修改文件

- `internal/service/jdbc_builtin_manifest.json`
- `internal/service/jdbc_catalog_test.go`
- `docs/changes/features/2026-07-13-domestic-jdbc-driver-sources.md`

## 验证

- 下载实际 Maven Central jar，使用 `shasum -a 256` 计算 SHA-256。
- 使用 `jar tf` 确认达梦包含 `dm.jdbc.driver.DmDriver`，人大金仓 V8/V9 包含 `com.kingbase8.Driver`。
- `go test ./internal/service -run TestDriverCatalogProvidesVerifiedDomesticOnlineProfiles -v`

## 剩余风险

未使用真实达梦或人大金仓服务器执行连接验证。驱动与服务端相差较大、启用读写分离或使用厂商特定参数时，用户应选择匹配的 V8/V9 profile，必要时通过离线导入厂商配套版本。
