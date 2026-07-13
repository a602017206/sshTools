# JDBC 清单迁移 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 让已存在的用户 JDBC 驱动清单自动获得新的内置在线 profile，同时保留用户自定义驱动。

**Architecture:** 将内置 manifest 版本提升为 `2`，并在 catalog 启动时读取历史清单。当历史版本较低时，合并内置 driver/profile、保留未知 driver/profile，并以原子写入替换清单。

**Tech Stack:** Go、JSON、Go `testing`。

---

### Task 1: 迁移历史 JDBC 清单

**Files:**
- Modify: `internal/service/jdbc_catalog.go`
- Modify: `internal/service/jdbc_builtin_manifest.json`
- Modify: `internal/service/jdbc_catalog_test.go`
- Create: `docs/changes/bugs/2026-07-13-jdbc-manifest-migration.md`

**Step 1: Write the failing test**

在 `jdbc_catalog_test.go` 创建版本 `1` 的历史 manifest，包含旧 `dm-8`、旧 `kingbase-8` 和自定义 `private` driver。调用 `LoadManifest` 后断言：

```go
_, dm, _ := catalog.GetProfile("dm", "8.1.5.45")
_, kingbaseV8, _ := catalog.GetProfile("kingbase", "8.6.1")
_, kingbaseV9, _ := catalog.GetProfile("kingbase", "9.0.1")
_, custom, _ := catalog.GetProfile("private", "1.0")
if dm.Jars[0].URL == "" || kingbaseV8.Jars[0].URL == "" || kingbaseV9.Jars[0].URL == "" || custom.ID != "private-1.0" {
	t.Fatal("migration did not preserve required profiles")
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/service -run TestDriverCatalogMigratesOutdatedBuiltinProfiles -v`

Expected: FAIL，`ensureManifest` 检测到文件存在后直接返回，旧 profile 保持不变。

**Step 3: Write minimal implementation**

1. 将内置 manifest `version` 改为 `2`。
2. 在 `ensureManifest` 中读取现有 JSON；若版本小于内置版本，按 driver ID 合并：内置 driver 覆盖同 ID 的官方 profile，历史中不在内置内的 profile 追加保留；未知 driver 整体保留。
3. 使用已有临时文件 + `os.Rename` 逻辑原子写回，权限 `0600`。
4. 当前版本或更高版本的清单不写回。

**Step 4: Run test to verify it passes**

Run: `go test ./internal/service -run TestDriverCatalogMigratesOutdatedBuiltinProfiles -v`

Expected: PASS。

**Step 5: Run focused verification and commit**

Run:

```bash
go test ./internal/service -run 'TestDriverCatalog(MigratesOutdatedBuiltinProfiles|ProvidesVerifiedDomesticOnlineProfiles)' -v
go test ./...
```

创建中文变更记录，写明旧清单升级、保留自定义 profile 的边界和验证结果。

```bash
git add internal/service/jdbc_catalog.go internal/service/jdbc_builtin_manifest.json internal/service/jdbc_catalog_test.go docs/changes/bugs/2026-07-13-jdbc-manifest-migration.md
git commit -m "fix: migrate outdated jdbc driver manifest"
```

### Task 2: 构建并验证升级路径

**Files:**
- Modify: `docs/changes/bugs/2026-07-13-jdbc-manifest-migration.md`

**Step 1: Build production application**

Run:

```bash
cd frontend && npm run build
/Users/dingwei/go/bin/wails build
```

如果 Gradle wrapper 因沙箱无法访问现有缓存而失败，先在变更记录中记录阻塞点和“授权访问现有 Gradle 缓存后重跑原命令”的最小修复方案，再重跑。

**Step 2: Record verification and commit**

在变更记录记录构建结果、历史 manifest 迁移测试和剩余风险。

```bash
git add docs/changes/bugs/2026-07-13-jdbc-manifest-migration.md frontend/build/assets/index.js frontend/wailsjs
git commit -m "docs: verify jdbc manifest migration"
```
