# 国产 JDBC 驱动自动安装来源

## 背景

达梦和人大金仓原有 profile 仅支持离线导入，未提供官方在线下载地址和 SHA-256，导致自动安装不可用。

## 范围

- 达梦 DM8 使用 `com.dameng:DmJdbcDriver8:8.1.5.45`。
- 人大金仓提供 V8 `8.6.1` 与 V9 `9.0.1` 两个独立 profile，默认推荐 V8。
- 三个 jar 都从 Maven Central HTTPS 地址下载，并使用固定 SHA-256 校验。
- 连接配置指定 `DriverProfileID` 时优先使用该 profile；旧连接未指定时继续使用推荐版本。
- 数据库连接新增兼容 API `ConnectDatabaseWithProfile`，将保存的 `driver_profile_id` 传给 JDBC gateway。
- 保留离线导入，供受限网络或服务端版本不匹配时使用。

## 修改文件

- `internal/service/jdbc_builtin_manifest.json`
- `internal/service/jdbc_catalog_test.go`
- `app.go`
- `app_jdbc_test.go`
- `internal/service/database_service.go`
- `internal/service/database_service_test.go`
- `frontend/src/App.svelte`
- `frontend_contract_test.go`
- `docs/changes/features/2026-07-13-domestic-jdbc-driver-sources.md`

## 验证

- 下载实际 Maven Central jar，使用 `shasum -a 256` 计算 SHA-256。
- 使用 `jar tf` 确认达梦包含 `dm.jdbc.driver.DmDriver`，人大金仓 V8/V9 包含 `com.kingbase8.Driver`。
- `go test ./internal/service -run TestDriverCatalogProvidesVerifiedDomesticOnlineProfiles -v`
- `go test . -run TestBuildJDBCServicesUsesConfiguredDriverProfile -v`
- `go test ./internal/service -run TestDatabaseServiceConnectDatabaseWithProfilePassesDriverProfileID -v`
- 临时受控测试在 `t.TempDir()` 中分别安装达梦 DM8 `8.1.5.45`、人大金仓 V8 `8.6.1` 与 V9 `9.0.1`，三者均完成 jar SHA-256 校验、原子安装及 `driver.json` 写入；测试产物已删除。

在隔离 worktree 首次执行 `cd frontend && npm run build` 时，因不存在 `frontend/node_modules` 而找不到 `vite`。最小修复方案是在该 worktree 的 `frontend/` 运行 `npm install` 安装项目锁定依赖后重跑原构建命令；该步骤不修改应用源码或依赖声明。

重跑后，Vite 编译成功，但 JDBC agent staging 调用 Gradle wrapper 时因沙箱无法访问 `~/.gradle/wrapper/dists/gradle-8.5-bin/5t9huq95ubn472n8rpzujfbqh/gradle-8.5-bin.zip.lck` 而停止。最小修复方案是不修改 Gradle、proto 或 Java agent，只在已授权访问现有 Gradle 缓存的环境中重跑相同命令。

## 剩余风险

未使用真实达梦或人大金仓服务器执行连接验证。驱动与服务端相差较大、启用读写分离或使用厂商特定参数时，用户应选择匹配的 V8/V9 profile，必要时通过离线导入厂商配套版本。
