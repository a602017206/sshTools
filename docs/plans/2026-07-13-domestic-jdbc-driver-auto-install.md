# 国产 JDBC 驱动自动安装 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 让达梦 DM8 与人大金仓 V8/V9 JDBC 驱动可从 Maven Central 校验安装，并在连接配置中选择和使用人大金仓对应的 profile。

**Architecture:** 内置 manifest 保存厂商已发布到 Maven Central 的固定坐标、HTTPS jar URL 与 SHA-256。人大金仓保留 V8 与 V9 两个独立 profile；连接配置将 profile ID 持久化到 metadata，并在 Go profile resolver 中优先解析它，未指定时才回退推荐 profile。

**Tech Stack:** Go、Wails、Svelte、Maven Central、SHA-256、Go `testing`。

---

### Task 1: 核验并配置达梦与人大金仓在线 Profile

**Files:**
- Modify: `internal/service/jdbc_builtin_manifest.json`
- Modify: `internal/service/jdbc_catalog_test.go`
- Create: `docs/changes/features/2026-07-13-domestic-jdbc-driver-sources.md`

**Step 1: Write the failing test**

在 `internal/service/jdbc_catalog_test.go` 添加 `TestDriverCatalogProvidesVerifiedDomesticOnlineProfiles`，加载内置 manifest，并断言：

```go
for _, wanted := range []struct {
	DriverID string
	Version  string
	Class    string
}{
	{"dm", "8.1.5.45", "dm.jdbc.driver.DmDriver"},
	{"kingbase", "8.6.1", "com.kingbase8.Driver"},
	{"kingbase", "9.0.1", "com.kingbase8.Driver"},
} {
	_, profile, err := catalog.GetProfile(wanted.DriverID, wanted.Version)
	if err != nil { t.Fatal(err) }
	if profile.DriverClass != wanted.Class { t.Fatal(profile.DriverClass) }
	for _, jar := range profile.Jars {
		if !strings.HasPrefix(jar.URL, "https://repo.maven.apache.org/") { t.Fatal(jar.URL) }
		if len(jar.SHA256) != 64 { t.Fatal(jar.SHA256) }
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/service -run TestDriverCatalogProvidesVerifiedDomesticOnlineProfiles -v`

Expected: FAIL，现有 `dm-8`、`kingbase-8` profile 没有在线 URL 与 SHA-256，且缺少 V9 profile。

**Step 3: Obtain and verify artifacts**

下载以下 Maven Central jar 到临时目录，通过 `shasum -a 256` 计算摘要，并检查 jar 中存在相应 driver class：

```bash
curl --fail --location --output /tmp/DmJdbcDriver8-8.1.5.45.jar https://repo.maven.apache.org/maven2/com/dameng/DmJdbcDriver8/8.1.5.45/DmJdbcDriver8-8.1.5.45.jar
curl --fail --location --output /tmp/kingbase8-8.6.1.jar https://repo.maven.apache.org/maven2/cn/com/kingbase/kingbase8/8.6.1/kingbase8-8.6.1.jar
curl --fail --location --output /tmp/kingbase8-9.0.1.jar https://repo.maven.apache.org/maven2/cn/com/kingbase/kingbase8/9.0.1/kingbase8-9.0.1.jar
shasum -a 256 /tmp/DmJdbcDriver8-8.1.5.45.jar /tmp/kingbase8-8.6.1.jar /tmp/kingbase8-9.0.1.jar
jar tf /tmp/DmJdbcDriver8-8.1.5.45.jar | rg 'dm/jdbc/driver/DmDriver.class'
jar tf /tmp/kingbase8-8.6.1.jar | rg 'com/kingbase8/Driver.class'
jar tf /tmp/kingbase8-9.0.1.jar | rg 'com/kingbase8/Driver.class'
```

Expected: 所有下载、SHA-256 计算和类检查成功。若 Maven Central 与官方文档的驱动类不匹配，停止实施并记录阻塞点与最小修复方案。

**Step 4: Write minimal implementation**

更新 `jdbc_builtin_manifest.json`：

- 达梦 profile 改为版本 `8.1.5.45`，jar 名称 `DmJdbcDriver8-8.1.5.45.jar`，使用核验后的 URL 与 SHA-256。
- 人大金仓保留推荐版本 `8.6.1`，新增 `9.0.1` profile；每个 profile 使用各自 jar 名称、URL、SHA-256、driver class 与 URL template。
- 仅使用 Maven Central HTTPS URL；不得保留空 checksum 的在线 profile。

**Step 5: Run test to verify it passes**

Run: `go test ./internal/service -run TestDriverCatalogProvidesVerifiedDomesticOnlineProfiles -v`

Expected: PASS。

**Step 6: Write change record and commit**

创建中文变更记录，说明官方来源、版本固定、校验策略和离线导入兜底。

```bash
git add internal/service/jdbc_builtin_manifest.json internal/service/jdbc_catalog_test.go docs/changes/features/2026-07-13-domestic-jdbc-driver-sources.md
git commit -m "feat: add verified domestic jdbc sources"
```

### Task 2: 按配置的 Profile ID 解析 JDBC 驱动

**Files:**
- Modify: `app.go`
- Modify: `app_test.go`
- Modify: `docs/changes/features/2026-07-13-domestic-jdbc-driver-sources.md`

**Step 1: Write the failing test**

在 `app_test.go` 添加 `TestBuildJDBCServicesUsesConfiguredDriverProfile`。创建包含 `kingbase` V8/V9 profile 的临时 manifest，构造 `DatabaseConfig{DBType: "kingbase", DriverProfileID: "kingbase-9.0.1"}`，调用 profile resolver，并断言返回 V9 profile 与其安装路径。

同时断言 `DriverProfileID == ""` 时返回 `RecommendedVersion` 指向的 V8 profile。

**Step 2: Run test to verify it fails**

Run: `go test . -run TestBuildJDBCServicesUsesConfiguredDriverProfile -v`

Expected: FAIL，当前 resolver 总是调用 `GetRecommendedProfile`。

**Step 3: Write minimal implementation**

在 `buildJDBCServices` 的 `gateway.SetProfileResolver` 中按以下逻辑选择：

```go
driver, profile, err := catalog.GetProfile(cfg.DBType, cfg.DriverProfileID)
if err != nil {
	return config.JDBCDriverProfile{}, err
}
resolved := *profile
resolved.InstallPath = filepath.Join(paths.DriversDir, driver.ID, profile.Version)
return resolved, nil
```

利用 `GetProfile` 既有的空版本回退行为，不新增分支或重复代码。

**Step 4: Run test to verify it passes**

Run: `go test . -run TestBuildJDBCServicesUsesConfiguredDriverProfile -v`

Expected: PASS。

**Step 5: Extend verification and commit**

Run: `go test . -run 'TestBuildJDBCServices|TestJDBCDriverManager' -v`

```bash
git add app.go app_test.go docs/changes/features/2026-07-13-domestic-jdbc-driver-sources.md
git commit -m "feat: resolve jdbc driver profile from connection"
```

### Task 3: 在数据库连接表单中选择并保存 Profile

**Files:**
- Modify: `frontend/src/components/AddAssetDialog.svelte`
- Modify: `frontend_contract_test.go`
- Modify: `docs/changes/features/2026-07-13-domestic-jdbc-driver-sources.md`

**Step 1: Write the failing contract test**

在 `frontend_contract_test.go` 添加 `TestAddAssetDialogPersistsSelectedJDBCProfile`，读取 `AddAssetDialog.svelte`，断言包含：

```go
"let selectedJDBCProfileId = ''"
"selectedJDBCDriver?.profiles"
"driver_profile_id: selectedJDBCProfileId"
"metadata?.driver_profile_id"
```

**Step 2: Run test to verify it fails**

Run: `go test . -run TestAddAssetDialogPersistsSelectedJDBCProfile -v`

Expected: FAIL，当前表单只保存 `db_type` 与 `database`。

**Step 3: Write minimal implementation**

在 `AddAssetDialog.svelte`：

1. 为 `formData` 增加 `driverProfileID`。
2. 根据 `selectedJDBCDriver.profiles` 渲染 profile 下拉框；仅显示已安装 profile。若没有已安装 profile，保留当前缺失驱动提示。
3. 数据库类型切换时默认选推荐版本或首个已安装 profile；编辑连接时从 `metadata.driver_profile_id` 恢复选择。
4. 测试连接前与提交前要求所选 profile 已安装。
5. 在 `metadata` 写入 `driver_profile_id: formData.driverProfileID || undefined`。

不得更改非数据库连接、SQLite 或 SSH 表单行为。

**Step 4: Run test to verify it passes**

Run: `go test . -run TestAddAssetDialogPersistsSelectedJDBCProfile -v`

Expected: PASS。

**Step 5: Build frontend and commit**

Run: `cd frontend && npm run build`

Expected: Vite 和 JDBC agent staging 成功；记录既有无关 Svelte 可访问性警告，但不得新增本组件编译错误。

```bash
git add frontend/src/components/AddAssetDialog.svelte frontend/build/assets/index.js frontend_contract_test.go docs/changes/features/2026-07-13-domestic-jdbc-driver-sources.md
git commit -m "feat: select jdbc profile for database connections"
```

### Task 4: 将连接元数据的 Profile 选择传入 JDBC Gateway

**Files:**
- Modify: `app.go`
- Modify: `internal/service/database_service.go`
- Modify: `internal/service/database_service_test.go`
- Modify: `docs/changes/features/2026-07-13-domestic-jdbc-driver-sources.md`

**Step 1: Write the failing test**

在 `internal/service/database_service_test.go` 添加测试，调用接收 `driverProfileID` 的连接路径并断言 gateway 收到的 `config.DatabaseConfig.DriverProfileID` 为 `kingbase-9.0.1`。

在 `app_test.go` 添加测试，构造 metadata `driver_profile_id` 后启动数据库连接，断言其传递到 `DatabaseService`。

**Step 2: Run tests to verify they fail**

Run: `go test . ./internal/service -run 'Test.*DriverProfile' -v`

Expected: FAIL，现有 Wails 连接 API 未接收 metadata 的 profile ID。

**Step 3: Write minimal implementation**

新增保持兼容的 Wails API `ConnectDatabaseWithProfile(sessionID, host, port, user, password, dbType, database, driverProfileID)`，并让应用数据库连接入口从 connection metadata 读取 `driver_profile_id` 后调用该方法。`ConnectDatabase` 继续作为空 profile 的兼容包装。

在 `DatabaseService` 新增内部连接方法构造 `DatabaseConfig` 时填充 `DriverProfileID`；原有 API 调用该方法并传空字符串。

**Step 4: Run tests to verify they pass**

Run: `go test . ./internal/service -run 'Test.*DriverProfile' -v`

Expected: PASS。

**Step 5: Run full verification and commit**

Run:

```bash
go test ./...
cd frontend && npm run build
/Users/dingwei/go/bin/wails build
```

Expected: 全部通过；如 Gradle wrapper 缓存权限受沙箱限制，先在变更记录中记录阻塞点和“授权访问现有 `~/.gradle` 缓存后重跑原命令”的最小修复方案，再请求授权重跑。

```bash
git add app.go app_test.go internal/service/database_service.go internal/service/database_service_test.go frontend/wailsjs frontend/build docs/changes/features/2026-07-13-domestic-jdbc-driver-sources.md
git commit -m "feat: pass jdbc profile through database connection"
```

### Task 5: 最终真实来源验证与交付检查

**Files:**
- Modify: `docs/changes/features/2026-07-13-domestic-jdbc-driver-sources.md`

**Step 1: Verify profile installation with real Maven artifacts**

运行临时 Go 验证程序或受控测试，分别调用 `InstallProfile` 安装达梦 DM8、人大金仓 V8、人大金仓 V9，并执行 `ValidateJDBCDriver`。完成后删除临时验证产物。

**Step 2: Verify selected profile resolution**

使用 fake gateway 连接 `kingbase-8.6.1` 和 `kingbase-9.0.1`，断言两次 `OpenSessionRequest.Profile.Id` 分别匹配选择值。

**Step 3: Record verification and commit**

在变更记录补充真实 URL、SHA-256 核验、完整命令及结果，说明未连接真实达梦/人大金仓服务端的风险。

```bash
git add docs/changes/features/2026-07-13-domestic-jdbc-driver-sources.md
git commit -m "docs: verify domestic jdbc driver installation"
```
